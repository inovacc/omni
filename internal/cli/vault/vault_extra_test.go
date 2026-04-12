package vault

import (
	"errors"
	"net/http"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/inovacc/omni/internal/cli/cmderr"
)

func TestClassifyVaultError_Nil(t *testing.T) {
	err := classifyVaultError(nil, "test op")
	if err != nil {
		t.Errorf("classifyVaultError(nil) = %v, want nil", err)
	}
}

func TestClassifyVaultError_ResponseError_Unauthorized(t *testing.T) {
	respErr := &api.ResponseError{StatusCode: http.StatusUnauthorized}
	err := classifyVaultError(respErr, "read secret")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if !errors.Is(err, cmderr.ErrPermission) {
		t.Errorf("classifyVaultError(401) should wrap ErrPermission, got: %v", err)
	}
}

func TestClassifyVaultError_ResponseError_Forbidden(t *testing.T) {
	respErr := &api.ResponseError{StatusCode: http.StatusForbidden}
	err := classifyVaultError(respErr, "write secret")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if !errors.Is(err, cmderr.ErrPermission) {
		t.Errorf("classifyVaultError(403) should wrap ErrPermission, got: %v", err)
	}
}

func TestClassifyVaultError_ResponseError_NotFound(t *testing.T) {
	respErr := &api.ResponseError{StatusCode: http.StatusNotFound}
	err := classifyVaultError(respErr, "read secret")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if !errors.Is(err, cmderr.ErrNotFound) {
		t.Errorf("classifyVaultError(404) should wrap ErrNotFound, got: %v", err)
	}
}

func TestClassifyVaultError_ResponseError_BadRequest(t *testing.T) {
	respErr := &api.ResponseError{StatusCode: http.StatusBadRequest}
	err := classifyVaultError(respErr, "put secret")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if !errors.Is(err, cmderr.ErrInvalidInput) {
		t.Errorf("classifyVaultError(400) should wrap ErrInvalidInput, got: %v", err)
	}
}

func TestClassifyVaultError_ResponseError_OtherStatus(t *testing.T) {
	respErr := &api.ResponseError{StatusCode: http.StatusInternalServerError}
	err := classifyVaultError(respErr, "list secrets")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if !errors.Is(err, cmderr.ErrIO) {
		t.Errorf("classifyVaultError(500) should wrap ErrIO, got: %v", err)
	}
}

func TestClassifyVaultError_PlainError(t *testing.T) {
	plain := errors.New("connection refused")
	err := classifyVaultError(plain, "connect")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if !errors.Is(err, cmderr.ErrIO) {
		t.Errorf("classifyVaultError(plain error) should wrap ErrIO, got: %v", err)
	}
}

func TestNewWithNamespace(t *testing.T) {
	client, err := New(Options{
		Namespace: "mynamespace",
	})
	if err != nil {
		t.Fatalf("New() with namespace failed: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	// Vault client doesn't expose namespace directly, just verify no error
}

func TestNewWithTLSSkip(t *testing.T) {
	client, err := New(Options{
		TLSSkip: true,
	})
	if err != nil {
		t.Fatalf("New() with TLSSkip failed: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestClientAddressAndToken(t *testing.T) {
	client, err := New(Options{
		Address: "http://vault.example.com:8200",
		Token:   "s.abc123",
	})
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if client.Address() != "http://vault.example.com:8200" {
		t.Errorf("Address() = %q, want %q", client.Address(), "http://vault.example.com:8200")
	}
	if client.Token() != "s.abc123" {
		t.Errorf("Token() = %q, want %q", client.Token(), "s.abc123")
	}
}

func TestSetTokenUpdates(t *testing.T) {
	client, err := New(Options{Token: "initial"})
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	client.SetToken("updated")
	if client.Token() != "updated" {
		t.Errorf("SetToken() did not update token, got %q", client.Token())
	}
}

func TestNewKVDefaultMount(t *testing.T) {
	client, err := New(Options{})
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	kv := client.NewKV("")
	if kv == nil {
		t.Fatal("NewKV() returned nil")
	}
	if kv.mount != "secret" {
		t.Errorf("NewKV() default mount = %q, want %q", kv.mount, "secret")
	}
}

func TestNewKVCustomMount(t *testing.T) {
	client, err := New(Options{})
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	kv := client.NewKV("kvv2")
	if kv.mount != "kvv2" {
		t.Errorf("NewKV() custom mount = %q, want %q", kv.mount, "kvv2")
	}
}

func TestGetTokenFile_NonEmpty(t *testing.T) {
	path := getTokenFile()
	if path == "" {
		t.Error("getTokenFile() returned empty string")
	}
}
