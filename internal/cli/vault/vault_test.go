package vault

import (
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	// Test with default options
	client, err := New(Options{})
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if client == nil {
		t.Fatal("expected non-nil client")
	}

	// Default address should be https://127.0.0.1:8200
	addr := client.Address()
	if addr == "" {
		t.Error("expected non-empty address")
	}
}

func TestNewWithAddress(t *testing.T) {
	testAddr := "http://localhost:8200"

	client, err := New(Options{
		Address: testAddr,
	})
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if client.Address() != testAddr {
		t.Errorf("expected address %s, got %s", testAddr, client.Address())
	}
}

func TestNewWithToken(t *testing.T) {
	testToken := "test-token-12345"

	client, err := New(Options{
		Token: testToken,
	})
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if client.Token() != testToken {
		t.Errorf("expected token %s, got %s", testToken, client.Token())
	}
}

func TestSetToken(t *testing.T) {
	client, err := New(Options{})
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	testToken := "new-token-xyz"
	client.SetToken(testToken)

	if client.Token() != testToken {
		t.Errorf("expected token %s, got %s", testToken, client.Token())
	}
}

func TestNewFromEnv(t *testing.T) {
	// Save original env
	origAddr := os.Getenv("VAULT_ADDR")
	origToken := os.Getenv("VAULT_TOKEN")

	defer func() {
		os.Setenv("VAULT_ADDR", origAddr)
		os.Setenv("VAULT_TOKEN", origToken)
	}()

	// Set test env
	os.Setenv("VAULT_ADDR", "http://test-vault:8200")
	os.Setenv("VAULT_TOKEN", "env-token")

	client, err := NewFromEnv()
	if err != nil {
		t.Fatalf("NewFromEnv() failed: %v", err)
	}

	if client.Address() != "http://test-vault:8200" {
		t.Errorf("expected address from env, got %s", client.Address())
	}

	if client.Token() != "env-token" {
		t.Errorf("expected token from env, got %s", client.Token())
	}
}

func TestNewKV(t *testing.T) {
	client, err := New(Options{})
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// Default mount
	kv := client.NewKV("")
	if kv == nil {
		t.Fatal("expected non-nil KV client")
	}

	if kv.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %s", kv.mount)
	}

	// Custom mount
	kv = client.NewKV("kv")
	if kv.mount != "kv" {
		t.Errorf("expected mount 'kv', got %s", kv.mount)
	}
}

func TestGetTokenFile(t *testing.T) {
	tokenFile := getTokenFile()
	if tokenFile == "" {
		t.Error("expected non-empty token file path")
	}

	// Should end with .vault-token
	if len(tokenFile) < 12 {
		t.Errorf("token file path too short: %s", tokenFile)
	}
}
