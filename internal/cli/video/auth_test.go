package video

import (
	"bytes"
	"os"
	"testing"
)

func TestIsYouTubeDomain(t *testing.T) {
	tests := []struct {
		domain string
		want   bool
	}{
		{".youtube.com", true},
		{"youtube.com", true},
		{"www.youtube.com", true},
		{".google.com", false},
		{"example.com", false},
	}

	for _, tt := range tests {
		if got := isYouTubeDomain(tt.domain); got != tt.want {
			t.Errorf("isYouTubeDomain(%q) = %v, want %v", tt.domain, got, tt.want)
		}
	}
}

func TestIsGoogleDomain(t *testing.T) {
	tests := []struct {
		domain string
		want   bool
	}{
		{".google.com", true},
		{"google.com", true},
		{"accounts.google.com", true},
		{".youtube.com", false},
	}

	for _, tt := range tests {
		if got := isGoogleDomain(tt.domain); got != tt.want {
			t.Errorf("isGoogleDomain(%q) = %v, want %v", tt.domain, got, tt.want)
		}
	}
}

func TestFindChrome(t *testing.T) {
	path := findChrome()
	if path == "" {
		t.Skip("Chrome not found on this system")
	}
	t.Logf("Chrome found at: %s", path)
}

func TestFreePort(t *testing.T) {
	port, err := freePort()
	if err != nil {
		t.Fatalf("freePort: %v", err)
	}
	if port < 1024 || port > 65535 {
		t.Errorf("unexpected port: %d", port)
	}
}

func TestWsEncodeTextFrame(t *testing.T) {
	payload := []byte(`{"id":1,"method":"Network.getAllCookies"}`)
	frame := wsEncodeTextFrame(payload)

	// First byte: 0x81 (FIN + text opcode).
	if frame[0] != 0x81 {
		t.Errorf("expected first byte 0x81, got 0x%02x", frame[0])
	}

	// Second byte: length | 0x80 (masked).
	if frame[1] != byte(len(payload))|0x80 {
		t.Errorf("expected length byte %d, got %d", byte(len(payload))|0x80, frame[1])
	}
}

func TestRunAuthIntegration(t *testing.T) {
	if os.Getenv("TEST_AUTH") == "" {
		t.Skip("set TEST_AUTH=1 to run (requires Chrome closed)")
	}

	var buf bytes.Buffer
	err := RunAuth(&buf, nil, Options{})
	t.Log(buf.String())
	if err != nil {
		t.Fatalf("RunAuth: %v", err)
	}
}
