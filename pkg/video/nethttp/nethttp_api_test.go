package nethttp_test

import (
	"testing"

	"github.com/inovacc/omni/pkg/video/nethttp"
)

func TestNewClient_API(t *testing.T) {
	c, err := nethttp.NewClient(nethttp.ClientOptions{})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if c == nil {
		t.Fatal("NewClient() returned nil")
	}
}

func TestNewClient_WithRetries_API(t *testing.T) {
	c, err := nethttp.NewClient(nethttp.ClientOptions{Retries: 2})
	if err != nil {
		t.Fatalf("NewClient(Retries=2) error = %v", err)
	}
	if c == nil {
		t.Fatal("NewClient(Retries=2) returned nil")
	}
}

func TestDefaultCookiePath_API(t *testing.T) {
	path := nethttp.DefaultCookiePath()
	if path == "" {
		t.Fatal("DefaultCookiePath() returned empty string")
	}
}
