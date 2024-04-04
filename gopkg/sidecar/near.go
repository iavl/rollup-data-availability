package sidecar

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type Client struct {
	client *http.Client
	/// The host of the sidecar service
	Host   string
	Config *ConfigureClientRequest
}

// / Config can be null here, the sidecar can be configured in ways outside of this package
func NewClient(host string, config *ConfigureClientRequest) (*Client, error) {
	if host == "" {
		host = "http://localhost:5888"
	}
	client := &Client{
		client: &http.Client{},
		Host:   host,
		Config: config,
	}
	return client, client.Health()
}

type Network string

var OverrideRPCUrl string

const (
	Mainnet  Network = "mainnet"
	Testnet  Network = "testnet"
	Localnet Network = "localnet"
)

type ConfigureClientRequest struct {
	/// The signer's near account id: "anexamplehere.near"
	AccountID string `json:"account_id"`
	/// The ed25519 secret key, bs58 encoded, prefixed with "ed25519:"
	/// Note: you can get this as is from your wallet
	/// Example: ed25519:5HeSvTBxttkRrBgds9aUGjMK5qP7EgBrgENPcRaCg4buBQcDZoSK8nbV9iP6F8TnBELiBTfS6L6wzNRH74wtC64K
	SecretKey string `json:"secret_key"`
	/// The contract id you're submitting to, e.g. "anexamplehere.near"
	ContractID string `json:"contract_id"`
	/// The network to connect to
	Network Network `json:"network"`
	/// Optionally a namespace, this should be registered on the shared blob registry
	Namespace *Namespace `json:"namespace"`
}

type Namespace struct {
	ID      int `json:"id"`
	Version int `json:"version"`
}

type BlobRef struct {
	TransactionID [32]byte `json:"transaction_id"`
}

func (b *BlobRef) Id() string {
	return fmt.Sprintf("%x", b.TransactionID)
}

func (b *BlobRef) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		TransactionID string `json:"transaction_id"`
	}{
		TransactionID: b.Id(),
	})
}

func (b *BlobRef) UnmarshalJSON(data []byte) error {
	var aux struct {
		TransactionID string `json:"transaction_id"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	decodedTransactionID, err := hex.DecodeString(aux.TransactionID)
	if err != nil {
		return err
	}
	copy(b.TransactionID[:], decodedTransactionID)
	return nil
}

func FromBytes(data []byte) (*BlobRef, error) {
	if len(data) != 32 {
		return nil, errors.New("invalid blob ref")
	}
	var blobRef BlobRef
	copy(blobRef.TransactionID[:], data)
	return &blobRef, nil
}

type Blob struct {
	Data []byte `json:"data"`
}

func (b *Blob) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Data string `json:"data"`
	}{
		Data: fmt.Sprintf("%x", b.Data),
	})
}

func (b *Blob) UnmarshalJSON(data []byte) error {
	var aux struct {
		Data string `json:"data"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	decodedData, err := hex.DecodeString(aux.Data)
	if err != nil {
		return err
	}
	b.Data = decodedData
	return nil
}

func (c *Client) ConfigureClient(req *ConfigureClientRequest) error {
	jsonData, err := json.Marshal(req)
	log.Info("ConfigureClient ", "jsonData", string(jsonData))
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequest(http.MethodPut, c.Host+"/configure", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	log.Info("ConfigureClient ", "resp", resp)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to configure client, status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) GetBlob(req BlobRef) (*Blob, error) {
	log.Info("GetBlob ", "tx ", req.Id())
	resp, err := c.client.Get(c.Host + "/blob?transaction_id=" + req.Id())
	log.Info("GetBlob ", "resp", resp)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get blob, status code: %d", resp.StatusCode)
	}

	var blob Blob
	err = json.NewDecoder(resp.Body).Decode(&blob)
	if err != nil {
		return nil, err
	}

	return &blob, nil
}

func (c *Client) SubmitBlob(req Blob) (*BlobRef, error) {
	jsonData, err := req.MarshalJSON()
	log.Info("SubmitBlob ", "jsonData", string(jsonData))
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Post(c.Host+"/blob", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to submit blob, status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Failed to read response body", "err", err)
		return nil, err
	}
	log.Info("SubmitBlob ", "responseBody", string(bodyBytes)) // Create a new reader from the body bytes

	hexBytes, err := hex.DecodeString(string(bodyBytes))
	if err != nil {
		log.Error("Failed to decode hex string", "err", err)
		return nil, err
	}

	blobRef, err := FromBytes(hexBytes)
	log.Info("SubmitBlob", "result", blobRef)
	if err != nil {
		return nil, err
	}

	return &*blobRef, nil
}

func (c *Client) Health() error {
	resp, err := c.client.Get(c.Host + "/health")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("health check failed")
	}

	return nil
}
