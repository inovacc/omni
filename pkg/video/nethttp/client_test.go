package nethttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewClientDefaults(t *testing.T) {
	c, err := NewClient(ClientOptions{})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	if c.ua != defaultUA {
		t.Errorf("ua = %q, want default", c.ua)
	}

	if c.retries != 3 {
		t.Errorf("retries = %d, want 3", c.retries)
	}
}

func TestNewClientCustom(t *testing.T) {
	c, err := NewClient(ClientOptions{
		UserAgent: "test/1.0",
		Retries:   5,
		Timeout:   10 * time.Second,
		Headers:   map[string]string{"X-Custom": "value"},
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	if c.ua != "test/1.0" {
		t.Errorf("ua = %q, want test/1.0", c.ua)
	}
	if c.retries != 5 {
		t.Errorf("retries = %d, want 5", c.retries)
	}
}

func TestNewClientBadProxy(t *testing.T) {
	_, err := NewClient(ClientOptions{Proxy: "://invalid"})
	if err == nil {
		t.Error("expected error for invalid proxy")
	}
}

func TestGetString(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hello"))
	}))
	defer srv.Close()

	c, _ := NewClient(ClientOptions{Retries: 1, Timeout: 5 * time.Second})
	got, err := c.GetString(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("GetString: %v", err)
	}

	if got != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}
}

func TestGetStringNonOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c, _ := NewClient(ClientOptions{Retries: 1, Timeout: 5 * time.Second})
	_, err := c.GetString(context.Background(), srv.URL)
	if err == nil {
		t.Error("expected error for 404")
	}
}

func TestGetJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"key":"value"}`))
	}))
	defer srv.Close()

	c, _ := NewClient(ClientOptions{Retries: 1, Timeout: 5 * time.Second})
	data, err := c.GetJSON(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("GetJSON: %v", err)
	}

	if string(data) != `{"key":"value"}` {
		t.Errorf("got %q", data)
	}
}

func TestPostJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q", ct)
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	c, _ := NewClient(ClientOptions{Retries: 1, Timeout: 5 * time.Second})
	data, err := c.PostJSON(context.Background(), srv.URL, strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("PostJSON: %v", err)
	}

	if string(data) != "ok" {
		t.Errorf("got %q", data)
	}
}

func TestDoSetsDefaults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ua := r.Header.Get("User-Agent"); ua != "test-ua" {
			t.Errorf("User-Agent = %q, want test-ua", ua)
		}
		if al := r.Header.Get("Accept-Language"); al == "" {
			t.Error("Accept-Language not set")
		}
	}))
	defer srv.Close()

	c, _ := NewClient(ClientOptions{UserAgent: "test-ua", Retries: 1, Timeout: 5 * time.Second})
	req, _ := http.NewRequest("GET", srv.URL, nil)
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	_ = resp.Body.Close()
}

func TestDoRetryOn500(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c, _ := NewClient(ClientOptions{Retries: 3, Timeout: 5 * time.Second})
	req, _ := http.NewRequest("GET", srv.URL, nil)
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if attempts != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
}

func TestBackoff(t *testing.T) {
	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{0, 1 * time.Second},
		{1, 2 * time.Second},
		{2, 4 * time.Second},
		{10, 30 * time.Second}, // capped
	}

	for _, tt := range tests {
		got := backoff(tt.attempt)
		if got != tt.want {
			t.Errorf("backoff(%d) = %v, want %v", tt.attempt, got, tt.want)
		}
	}
}

func TestHTTPClient(t *testing.T) {
	c, _ := NewClient(ClientOptions{})
	if c.HTTPClient() == nil {
		t.Error("HTTPClient returned nil")
	}
}

func TestCookieJar(t *testing.T) {
	c, _ := NewClient(ClientOptions{})
	if c.CookieJar() == nil {
		t.Error("CookieJar returned nil")
	}
}
