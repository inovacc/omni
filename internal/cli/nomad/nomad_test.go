package nomad

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

func newTestClient(t *testing.T, url string) *Client {
	t.Helper()

	c, err := New(Options{Address: url})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	return c
}

func TestNewDefaultAddress(t *testing.T) {
	t.Setenv("NOMAD_ADDR", "")

	c, err := New(Options{})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if c.Address() != defaultAddress {
		t.Errorf("Address() = %q, want %q", c.Address(), defaultAddress)
	}
}

func TestNewEnvAddress(t *testing.T) {
	t.Setenv("NOMAD_ADDR", "http://nomad.example:4646")

	c, err := New(Options{})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if c.Address() != "http://nomad.example:4646" {
		t.Errorf("Address() = %q, want env value", c.Address())
	}
}

// errCase is a shared status→sentinel error matrix for the list endpoints.
type errCase struct {
	name    string
	status  int
	wantErr error
}

var errCases = []errCase{
	{name: "not found", status: http.StatusNotFound, wantErr: cmderr.ErrNotFound},
	{name: "forbidden", status: http.StatusForbidden, wantErr: cmderr.ErrPermission},
	{name: "unauthorized", status: http.StatusUnauthorized, wantErr: cmderr.ErrPermission},
	{name: "bad request", status: http.StatusBadRequest, wantErr: cmderr.ErrInvalidInput},
	{name: "server error", status: http.StatusInternalServerError, wantErr: cmderr.ErrIO},
}

func TestJobList(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/v1/jobs" {
				t.Errorf("path = %q, want /v1/jobs", r.URL.Path)
			}
			_, _ = w.Write([]byte(`[{"id":"j1","name":"web","status":"running"}]`))
		}))
		defer srv.Close()

		c := newTestClient(t, srv.URL)
		jobs, err := c.JobList(context.Background())
		if err != nil {
			t.Fatalf("JobList() error = %v", err)
		}
		if len(jobs) != 1 || jobs[0].Status != "running" {
			t.Errorf("unexpected jobs: %+v", jobs)
		}

		var buf bytes.Buffer
		if err := PrintJobs(&buf, jobs, false); err != nil {
			t.Fatalf("PrintJobs() error = %v", err)
		}
		if !strings.Contains(buf.String(), "web") {
			t.Errorf("output missing job name: %q", buf.String())
		}
	})

	for _, tc := range errCases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
			}))
			defer srv.Close()

			c := newTestClient(t, srv.URL)
			if _, err := c.JobList(context.Background()); !errors.Is(err, tc.wantErr) {
				t.Fatalf("JobList() error = %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestNodeList(t *testing.T) {
	t.Run("happy path includes drain", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/v1/nodes" {
				t.Errorf("path = %q, want /v1/nodes", r.URL.Path)
			}
			_, _ = w.Write([]byte(`[{"id":"n1","name":"node-1","status":"ready","drain":true}]`))
		}))
		defer srv.Close()

		c := newTestClient(t, srv.URL)
		nodes, err := c.NodeList(context.Background())
		if err != nil {
			t.Fatalf("NodeList() error = %v", err)
		}
		if len(nodes) != 1 || !nodes[0].Drain {
			t.Errorf("unexpected nodes: %+v", nodes)
		}

		var buf bytes.Buffer
		if err := PrintNodes(&buf, nodes, false); err != nil {
			t.Fatalf("PrintNodes() error = %v", err)
		}
		if !strings.Contains(buf.String(), "true") {
			t.Errorf("output missing drain: %q", buf.String())
		}
	})

	for _, tc := range errCases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
			}))
			defer srv.Close()

			c := newTestClient(t, srv.URL)
			if _, err := c.NodeList(context.Background()); !errors.Is(err, tc.wantErr) {
				t.Fatalf("NodeList() error = %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestAllocList(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/v1/allocations" {
				t.Errorf("path = %q, want /v1/allocations", r.URL.Path)
			}
			_, _ = w.Write([]byte(`[{"id":"a1","jobID":"web","status":"running"}]`))
		}))
		defer srv.Close()

		c := newTestClient(t, srv.URL)
		allocs, err := c.AllocList(context.Background())
		if err != nil {
			t.Fatalf("AllocList() error = %v", err)
		}
		if len(allocs) != 1 || allocs[0].JobID != "web" {
			t.Errorf("unexpected allocs: %+v", allocs)
		}

		var buf bytes.Buffer
		if err := PrintAllocs(&buf, allocs, true); err != nil {
			t.Fatalf("PrintAllocs(json) error = %v", err)
		}
		if !strings.Contains(buf.String(), `"jobID": "web"`) {
			t.Errorf("json output unexpected: %q", buf.String())
		}
	})

	for _, tc := range errCases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
			}))
			defer srv.Close()

			c := newTestClient(t, srv.URL)
			if _, err := c.AllocList(context.Background()); !errors.Is(err, tc.wantErr) {
				t.Fatalf("AllocList() error = %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestTokenAndQueryParams(t *testing.T) {
	var gotToken, gotNamespace, gotRegion string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotToken = r.Header.Get("X-Nomad-Token")
		gotNamespace = r.URL.Query().Get("namespace")
		gotRegion = r.URL.Query().Get("region")
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	c, err := New(Options{Address: srv.URL, Token: "tok", Namespace: "prod", Region: "us-east"})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if _, err := c.JobList(context.Background()); err != nil {
		t.Fatalf("JobList() error = %v", err)
	}
	if gotToken != "tok" {
		t.Errorf("X-Nomad-Token = %q, want tok", gotToken)
	}
	if gotNamespace != "prod" {
		t.Errorf("namespace = %q, want prod", gotNamespace)
	}
	if gotRegion != "us-east" {
		t.Errorf("region = %q, want us-east", gotRegion)
	}
}

func TestTransportErrorIsIO(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := srv.URL
	srv.Close()

	c := newTestClient(t, url)
	if _, err := c.JobList(context.Background()); !errors.Is(err, cmderr.ErrIO) {
		t.Fatalf("JobList() on dead server error = %v, want ErrIO", err)
	}
}
