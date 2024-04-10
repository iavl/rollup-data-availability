package sidecar

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"testing"
)

// TODO: setup the sidecar in tests
func initClient(t *testing.T) *Client {
	// Read the configuration from the "http-config.json" file
	configData, err := os.ReadFile("../../http-config.json")
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	// Unmarshal the JSON data into a ConfigureClientRequest struct
	var conf ConfigureClientRequest
	err = json.Unmarshal(configData, &conf)
	if err != nil {
		panic(fmt.Errorf("failed to unmarshal config: %v", err))
	}

	client, err := NewClient("http://localhost:5888", nil)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	return client
}

func TestNewClient(t *testing.T) {
	// Test creating a new client with default host
	client, err := NewClient("", nil)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	// Test creating a new client with custom host and configuration
	config := &ConfigureClientRequest{
		AccountID:  "test_account",
		SecretKey:  "test_secret_key",
		ContractID: "test_contract",
		Network:    Testnet,
		Namespace:  &Namespace{ID: 1, Version: 1},
	}
	client, err = NewClient("http://localhost:5888", config)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()
}

func TestConfigureClient(t *testing.T) {
	client := initClient(t)
	defer client.Close()

	// Test configuring the client
	config := &ConfigureClientRequest{
		AccountID:  "test_account",
		SecretKey:  "test_secret_key",
		ContractID: "test_contract",
		Network:    Testnet,
		Namespace:  &Namespace{ID: 1, Version: 1},
	}
	err := client.ConfigureClient(config)
	if err != nil {
		log.Error("TestConfigureClient - likely already configured ", err)
	}
}

func TestGetBlob(t *testing.T) {
	client := initClient(t)
	defer client.Close()

	// Test getting a blob with a valid reference
	transactionID := generateTransactionID(t)
	blobRef, err := NewBlobRef(transactionID)
	log.Info("TestGetBlob blobRef ", blobRef)
	if err != nil {
		t.Fatalf("failed to create blob reference: %v", err)
	}
	blob, err := client.GetBlob(*blobRef)
	log.Info("TestGetBlob blob ", blob)
	if err != nil {
		t.Fatalf("failed to get blob: %v", err)
	}
	if blob == nil {
		t.Fatal("got nil blob")
	}

	// Test getting a blob with an invalid reference
	invalidTransactionID := []byte("invalid_transaction_id")
	log.Info("TestGetBlob invalidTransactionID ", invalidTransactionID)
	invalidBlobRef := &BlobRef{}
	log.Info("TestGetBlob invalidBlobRef ", invalidBlobRef)
	copy(invalidBlobRef.transactionID[:], invalidTransactionID)
	blob, err = client.GetBlob(*invalidBlobRef)
	log.Info("TestGetBlob invalidBlob ", blob)
	if err == nil {
		t.Fatal("expected an error but got none")
	}
	if blob != nil {
		t.Fatalf("expected nil blob but got %v", blob)
	}
}

func TestSubmitBlob(t *testing.T) {
	client := initClient(t)
	defer client.Close()

	// Test submitting a blob
	data := []byte("test_data")
	blob := Blob{Data: data}
	log.Info("TestSubmitBlob blob ", blob)
	blobRef, err := client.SubmitBlob(blob)
	log.Info("TestSubmitBlob blobRef ", blobRef)
	if err != nil {
		t.Fatalf("failed to submit blob: %v", err)
	}
	if blobRef == nil {
		t.Fatal("got nil blob reference")
	}

	// Test submitting an empty blob
	emptyBlob := Blob{}
	blobRef, err = client.SubmitBlob(emptyBlob)
	log.Info("TestSubmitBlob emptyBlob ", emptyBlob)
	if err == nil {
		t.Fatal("expected an error but got none")
	}
	if blobRef != nil {
		t.Fatalf("expected nil blob reference but got %v", blobRef)
	}
}

func TestHealth(t *testing.T) {
	client := initClient(t)
	defer client.Close()

	// Test checking the health of the service
	err := client.Health()
	if err != nil {
		t.Fatalf("health check failed: %v", err)
	}
}

func TestBlobMarshalUnmarshal(t *testing.T) {
	data := []byte("test_data")
	blob := Blob{Data: data}

	// Test marshaling the blob
	jsonData, err := blob.MarshalJSON()
	if err != nil {
		t.Fatalf("failed to marshal blob: %v", err)
	}

	// Test unmarshaling the blob
	var unmarshaled Blob
	err = unmarshaled.UnmarshalJSON(jsonData)
	if err != nil {
		t.Fatalf("failed to unmarshal blob: %v", err)
	}

	if !bytes.Equal(unmarshaled.Data, data) {
		t.Fatalf("unmarshaled blob data does not match original data")
	}
}

func TestNewBlobRefInvalidTransactionID(t *testing.T) {
	invalidTransactionID := []byte("invalid_transaction_id")
	_, err := NewBlobRef(invalidTransactionID)
	if err == nil {
		t.Fatal("expected an error but got none")
	}
}

func generateTransactionID(t *testing.T) []byte {

	hex, err := hex.DecodeString("5d0472abe8eef76f9a44a3695d584af4de6e2ddde82dabfa5c8f29e5aec1270d")
	log.Info("generateTransactionID hex ", hex)
	if err != nil {
		t.Errorf("Failed to decode hex string: %v", err)
	}

	blobRef, err := NewBlobRef(hex)
	log.Info("generateTransactionID blobRef", blobRef)
	if err != nil {
		t.Fatalf("failed to create blob reference: %v", err)
	}
	return blobRef.transactionID[:]
}