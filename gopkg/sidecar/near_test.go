package sidecar

import (
	"encoding/hex"
	"testing"
)

func InitClient() *Client {

	conf := ConfigureClientRequest{
		AccountID:  "blarg233.testnet",
		SecretKey:  "ed25519:FE1gAkdXMr2zJt4aHMgy1R1LbZAdhu6npxYzeKugxYJq2h5MgxahxcVfwXJKGg6RSSRFqFstJRxVfdZGDyTUSxt",
		ContractID: "blarg233.testnet",
		Network:    "testnet",
		Namespace:  nil,
	}

	client := NewClient("http://localhost:5888", &conf)
	return client
}

func TestConfigureClient(t *testing.T) {
	client := InitClient()

	err := client.ConfigureClient(client.Config)
	if err != nil {
		t.Errorf("ConfigureClient failed: %v", err)
	}
}

func TestGetBlob(t *testing.T) {
	client := InitClient()

	hex, err := hex.DecodeString("5d0472abe8eef76f9a44a3695d584af4de6e2ddde82dabfa5c8f29e5aec1270d")
	if err != nil {
		t.Errorf("Failed to decode hex string: %v", err)
	}
	blobRef, err := FromBytes(hex)

	blob, err := client.GetBlob(*blobRef)
	if err != nil {
		t.Errorf("GetBlob failed: %v", err)
	}

	if blob == nil {
		t.Error("GetBlob returned nil blob")
	}
}

func TestSubmitBlob(t *testing.T) {
	client := InitClient()

	req := Blob{
		Data: []byte("test_data"),
	}

	result, err := client.SubmitBlob(req)
	if err != nil {
		t.Errorf("SubmitBlob failed: %v", err)
	}

	if result == nil {
		t.Error("SubmitBlob returned empty result")
	}
}

func TestHealth(t *testing.T) {
	client := InitClient()

	err := client.Health()
	if err != nil {
		t.Errorf("Health check failed: %v", err)
	}
}
