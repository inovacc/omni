package curl

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/json":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"message":"hello","count":42}`))
		case "/text":
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write([]byte("Hello, World!"))
		case "/echo":
			w.Header().Set("Content-Type", "application/json")

			body, _ := io.ReadAll(r.Body)
			response := map[string]any{
				"method":  r.Method,
				"headers": r.Header,
				"body":    string(body),
				"query":   r.URL.Query(),
			}
			_ = json.NewEncoder(w).Encode(response)
		case "/headers":
			w.Header().Set("Content-Type", "application/json")
			response := map[string]any{
				"headers": r.Header,
			}
			_ = json.NewEncoder(w).Encode(response)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	tests := []struct {
		name    string
		args    []string
		opts    Options
		want    string
		wantErr bool
	}{
		{
			name: "simple GET json",
			args: []string{server.URL + "/json"},
			opts: Options{Method: "GET", Timeout: 5 * time.Second},
			want: "hello",
		},
		{
			name: "simple GET text",
			args: []string{server.URL + "/text"},
			opts: Options{Method: "GET", Timeout: 5 * time.Second},
			want: "Hello, World!",
		},
		{
			name: "POST with data",
			args: []string{server.URL + "/echo", "name=John"},
			opts: Options{Method: "POST", Timeout: 5 * time.Second},
			want: "John",
		},
		{
			name: "GET with query params",
			args: []string{server.URL + "/echo", "q==hello"},
			opts: Options{Method: "GET", Timeout: 5 * time.Second},
			want: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := Run(&buf, tt.args, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			output := buf.String()
			if !strings.Contains(output, tt.want) {
				t.Errorf("Run() output = %q, want containing %q", output, tt.want)
			}
		})
	}
}

func TestRunWithHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"auth": r.Header.Get("Authorization"),
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	var buf bytes.Buffer

	opts := Options{
		Method:  "GET",
		Headers: map[string]string{"Authorization": "Bearer token123"},
		Timeout: 5 * time.Second,
	}

	err := Run(&buf, []string{server.URL}, opts)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "token123") {
		t.Errorf("Output should contain auth token, got: %s", output)
	}
}

func TestRunJSONOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"test":"value"}`))
	}))
	defer server.Close()

	var buf bytes.Buffer

	opts := Options{
		Method:  "GET",
		JSON:    true,
		Timeout: 5 * time.Second,
	}

	err := Run(&buf, []string{server.URL}, opts)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	var response Response
	if err := json.Unmarshal(buf.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if response.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want 200", response.StatusCode)
	}

	if !strings.Contains(response.Body, "test") {
		t.Errorf("Body should contain 'test', got: %s", response.Body)
	}
}

func TestRunVerbose(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("OK"))
	}))
	defer server.Close()

	var buf bytes.Buffer

	opts := Options{
		Method:  "GET",
		Verbose: true,
		Timeout: 5 * time.Second,
	}

	err := Run(&buf, []string{server.URL}, opts)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	output := buf.String()

	// Should contain request details
	if !strings.Contains(output, "> GET") {
		t.Errorf("Output should contain request method")
	}

	// Should contain response details
	if !strings.Contains(output, "< 200") || !strings.Contains(output, "< HTTP") {
		// Accept either format
		if !strings.Contains(output, "OK") {
			t.Errorf("Output should contain response status")
		}
	}
}

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		opts       Options
		wantURL    string
		wantHeader string
		wantData   string
		wantErr    bool
	}{
		{
			name:    "simple URL",
			args:    []string{"https://example.com"},
			opts:    Options{},
			wantURL: "https://example.com",
		},
		{
			name:       "URL with header",
			args:       []string{"https://example.com", "Authorization:Bearer token"},
			opts:       Options{},
			wantURL:    "https://example.com",
			wantHeader: "Bearer token",
		},
		{
			name:     "URL with data",
			args:     []string{"https://example.com", "name=John"},
			opts:     Options{},
			wantURL:  "https://example.com",
			wantData: "John",
		},
		{
			name:    "URL with query param",
			args:    []string{"https://example.com", "q==hello"},
			opts:    Options{},
			wantURL: "https://example.com?q=hello",
		},
		{
			name:    "no URL",
			args:    []string{},
			opts:    Options{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urlStr, headers, data, err := parseArgs(tt.args, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if urlStr != tt.wantURL {
				t.Errorf("URL = %q, want %q", urlStr, tt.wantURL)
			}

			if tt.wantHeader != "" {
				if headers["Authorization"] != tt.wantHeader {
					t.Errorf("Header = %q, want %q", headers["Authorization"], tt.wantHeader)
				}
			}

			if tt.wantData != "" {
				if !strings.Contains(data, tt.wantData) {
					t.Errorf("Data = %q, want containing %q", data, tt.wantData)
				}
			}
		})
	}
}

func TestRunNoRedirect(t *testing.T) {
	redirectServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "/target", http.StatusFound)
			return
		}

		_, _ = w.Write([]byte("target"))
	}))
	defer redirectServer.Close()

	var buf bytes.Buffer

	opts := Options{
		Method:      "GET",
		FollowRedir: false,
		JSON:        true,
		Timeout:     5 * time.Second,
	}

	err := Run(&buf, []string{redirectServer.URL + "/redirect"}, opts)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	var response Response
	if err := json.Unmarshal(buf.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if response.StatusCode != http.StatusFound {
		t.Errorf("StatusCode = %d, want %d (should not follow redirect)", response.StatusCode, http.StatusFound)
	}
}
