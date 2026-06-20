package consul

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

// newTestClient returns a client pointed at the given httptest server URL.
func newTestClient(t *testing.T, url string) *Client {
	t.Helper()

	c, err := New(Options{Address: url})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	return c
}

func TestNewDefaultAddress(t *testing.T) {
	t.Setenv("CONSUL_HTTP_ADDR", "")

	c, err := New(Options{})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if c.Address() != defaultAddress {
		t.Errorf("Address() = %q, want %q", c.Address(), defaultAddress)
	}
}

func TestNewEnvAddress(t *testing.T) {
	t.Setenv("CONSUL_HTTP_ADDR", "http://consul.example:8500")

	c, err := New(Options{})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if c.Address() != "http://consul.example:8500" {
		t.Errorf("Address() = %q, want env value", c.Address())
	}
}

func TestMembers(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		body       string
		wantErr    error
		wantRows   int
		wantHeader bool
	}{
		{
			name:       "happy path",
			status:     http.StatusOK,
			body:       `[{"name":"n1","addr":"10.0.0.1","status":1},{"name":"n2","addr":"10.0.0.2","status":1}]`,
			wantRows:   2,
			wantHeader: true,
		},
		{name: "not found", status: http.StatusNotFound, body: ``, wantErr: cmderr.ErrNotFound},
		{name: "forbidden", status: http.StatusForbidden, body: ``, wantErr: cmderr.ErrPermission},
		{name: "unauthorized", status: http.StatusUnauthorized, body: ``, wantErr: cmderr.ErrPermission},
		{name: "bad request", status: http.StatusBadRequest, body: ``, wantErr: cmderr.ErrInvalidInput},
		{name: "server error", status: http.StatusInternalServerError, body: ``, wantErr: cmderr.ErrIO},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/agent/members" {
					t.Errorf("path = %q, want /v1/agent/members", r.URL.Path)
				}
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			c := newTestClient(t, srv.URL)
			members, err := c.Members(context.Background())

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("Members() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Members() unexpected error = %v", err)
			}
			if len(members) != tt.wantRows {
				t.Errorf("got %d members, want %d", len(members), tt.wantRows)
			}

			var buf bytes.Buffer
			if err := PrintMembers(&buf, members, false); err != nil {
				t.Fatalf("PrintMembers() error = %v", err)
			}
			if tt.wantHeader && !strings.Contains(buf.String(), "Node") {
				t.Errorf("text output missing header: %q", buf.String())
			}
		})
	}
}

func TestKVGet(t *testing.T) {
	t.Run("happy path decodes base64", func(t *testing.T) {
		val := base64.StdEncoding.EncodeToString([]byte("hello-world"))
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/v1/kv/myapp/config" {
				t.Errorf("path = %q", r.URL.Path)
			}
			_, _ = w.Write([]byte(`[{"Key":"myapp/config","Value":"` + val + `"}]`))
		}))
		defer srv.Close()

		c := newTestClient(t, srv.URL)
		got, err := c.KVGet(context.Background(), "myapp/config")
		if err != nil {
			t.Fatalf("KVGet() error = %v", err)
		}
		if got != "hello-world" {
			t.Errorf("KVGet() = %q, want hello-world", got)
		}
	})

	t.Run("empty key", func(t *testing.T) {
		c := newTestClient(t, "http://127.0.0.1:1")
		_, err := c.KVGet(context.Background(), "")
		if !errors.Is(err, cmderr.ErrInvalidInput) {
			t.Fatalf("KVGet(empty) error = %v, want ErrInvalidInput", err)
		}
	})

	t.Run("404 not found", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer srv.Close()

		c := newTestClient(t, srv.URL)
		_, err := c.KVGet(context.Background(), "missing")
		if !errors.Is(err, cmderr.ErrNotFound) {
			t.Fatalf("KVGet(missing) error = %v, want ErrNotFound", err)
		}
	})

	t.Run("empty array is not found", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`[]`))
		}))
		defer srv.Close()

		c := newTestClient(t, srv.URL)
		_, err := c.KVGet(context.Background(), "empty")
		if !errors.Is(err, cmderr.ErrNotFound) {
			t.Fatalf("KVGet(empty array) error = %v, want ErrNotFound", err)
		}
	})
}

func TestServices(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{"consul":[],"web":["v1","prod"]}`))
		}))
		defer srv.Close()

		c := newTestClient(t, srv.URL)
		svcs, err := c.Services(context.Background())
		if err != nil {
			t.Fatalf("Services() error = %v", err)
		}
		if len(svcs) != 2 {
			t.Errorf("got %d services, want 2", len(svcs))
		}

		var buf bytes.Buffer
		if err := PrintServices(&buf, svcs, false); err != nil {
			t.Fatalf("PrintServices() error = %v", err)
		}
		if !strings.Contains(buf.String(), "web") || !strings.Contains(buf.String(), "v1,prod") {
			t.Errorf("output missing service/tags: %q", buf.String())
		}
	})

	t.Run("403 forbidden", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		}))
		defer srv.Close()

		c := newTestClient(t, srv.URL)
		_, err := c.Services(context.Background())
		if !errors.Is(err, cmderr.ErrPermission) {
			t.Fatalf("Services() error = %v, want ErrPermission", err)
		}
	})
}

func TestTokenHeaderSent(t *testing.T) {
	var gotToken string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotToken = r.Header.Get("X-Consul-Token")
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	c, err := New(Options{Address: srv.URL, Token: "secret-token"})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if _, err := c.Members(context.Background()); err != nil {
		t.Fatalf("Members() error = %v", err)
	}
	if gotToken != "secret-token" {
		t.Errorf("X-Consul-Token = %q, want secret-token", gotToken)
	}
}

func TestTransportErrorIsIO(t *testing.T) {
	// Point at a closed server to force a transport-level error.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := srv.URL
	srv.Close()

	c := newTestClient(t, url)
	_, err := c.Members(context.Background())
	if !errors.Is(err, cmderr.ErrIO) {
		t.Fatalf("Members() on dead server error = %v, want ErrIO", err)
	}
}

func TestPrintJSON(t *testing.T) {
	var buf bytes.Buffer
	members := []Member{{Name: "n1", Addr: "10.0.0.1", Status: 1}}
	if err := PrintMembers(&buf, members, true); err != nil {
		t.Fatalf("PrintMembers(json) error = %v", err)
	}
	if !strings.Contains(buf.String(), `"name": "n1"`) {
		t.Errorf("json output unexpected: %q", buf.String())
	}
}

func TestMemberStatusName(t *testing.T) {
	tests := []struct {
		status int
		want   string
	}{
		{0, "none"},
		{1, "alive"},
		{2, "leaving"},
		{3, "left"},
		{4, "failed"},
		{9, "unknown(9)"},
	}
	for _, tt := range tests {
		if got := (Member{Status: tt.status}).StatusName(); got != tt.want {
			t.Errorf("StatusName(%d) = %q, want %q", tt.status, got, tt.want)
		}
	}
}
