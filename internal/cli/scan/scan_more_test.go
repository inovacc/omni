package scan

import (
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

func TestValidateFetchURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"empty", "", true},
		{"bad scheme ftp", "ftp://example.com/x", true},
		{"file scheme", "file:///etc/passwd", true},
		{"missing host", "http://", true},
		{"private ip", "http://10.0.0.1/x", true},
		{"loopback ip", "http://127.0.0.1/x", true},
		{"link-local metadata", "http://169.254.169.254/latest", true},
		{"unparseable", "http://[::1", true},
		{"public ip", "https://8.8.8.8/x", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFetchURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateFetchURL(%q) err=%v wantErr=%v", tt.url, err, tt.wantErr)
			}
		})
	}
}

func TestIsPublicFetchIP(t *testing.T) {
	tests := []struct {
		ip   string
		want bool
	}{
		{"8.8.8.8", true},
		{"1.1.1.1", true},
		{"127.0.0.1", false},
		{"10.1.2.3", false},
		{"192.168.0.1", false},
		{"172.16.0.1", false},
		{"169.254.169.254", false},
		{"0.0.0.0", false},
		{"224.0.0.1", false},
		{"100.64.1.1", false}, // CGNAT
		{"100.127.255.255", false},
		{"::1", false},
		{"2606:4700:4700::1111", true},
	}
	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if ip == nil {
				t.Fatalf("bad test ip %q", tt.ip)
			}
			if got := isPublicFetchIP(ip); got != tt.want {
				t.Fatalf("isPublicFetchIP(%s)=%v want %v", tt.ip, got, tt.want)
			}
		})
	}
}

func TestIsPublicFetchIPLoopbackAllowed(t *testing.T) {
	old := allowLoopbackFetch
	allowLoopbackFetch = true
	defer func() { allowLoopbackFetch = old }()
	if !isPublicFetchIP(net.ParseIP("127.0.0.1")) {
		t.Fatal("loopback should be permitted when allowLoopbackFetch is set")
	}
}

func TestPlural(t *testing.T) {
	if plural(1) != "y" {
		t.Errorf("plural(1)=%q want y", plural(1))
	}
	for _, n := range []int{0, 2, 5} {
		if plural(n) != "ies" {
			t.Errorf("plural(%d)=%q want ies", n, plural(n))
		}
	}
}

func TestClassifyFileErr(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want error
	}{
		{"notexist", os.ErrNotExist, cmderr.ErrNotFound},
		{"permission", os.ErrPermission, cmderr.ErrPermission},
		{"other", errors.New("disk full"), cmderr.ErrIO},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyFileErr(tt.err, "/some/path")
			if !errors.Is(got, tt.want) {
				t.Fatalf("classifyFileErr(%v) -> %v want sentinel %v", tt.err, got, tt.want)
			}
			if !strings.Contains(got.Error(), "/some/path") {
				t.Errorf("error %q should mention path", got)
			}
		})
	}
}

func TestClassifyDBErr(t *testing.T) {
	// Default (unknown error) classifies as ErrInvalidInput.
	got := classifyDBErr(errors.New("boom"))
	if !errors.Is(got, cmderr.ErrInvalidInput) {
		t.Fatalf("default classifyDBErr -> %v want ErrInvalidInput", got)
	}
}

func TestWriteAtomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.bin")
	want := []byte("hello atomic world")
	if err := writeAtomic(path, want); err != nil {
		t.Fatalf("writeAtomic: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("content = %q want %q", got, want)
	}
	// No leftover temp files in the dir.
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".osv-db-") {
			t.Errorf("leftover temp file %s", e.Name())
		}
	}
}

func TestWriteAtomicBadDir(t *testing.T) {
	// A path in a nonexistent directory should fail (CreateTemp fails).
	err := writeAtomic(filepath.Join(t.TempDir(), "nope", "x"), []byte("x"))
	if err == nil {
		t.Fatal("expected error writing into nonexistent dir")
	}
}

func TestManifestSummary(t *testing.T) {
	// Non-zip bytes -> (-1, "").
	if n, g := manifestSummary([]byte("not a zip")); n != -1 || g != "" {
		t.Fatalf("manifestSummary(garbage)=(%d,%q) want (-1,\"\")", n, g)
	}
}

func TestFetchHTTPTest(t *testing.T) {
	old := allowLoopbackFetch
	allowLoopbackFetch = true
	defer func() { allowLoopbackFetch = old }()

	t.Run("ok", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("payload-bytes"))
		}))
		defer srv.Close()
		got, err := fetch(srv.URL)
		if err != nil {
			t.Fatalf("fetch: %v", err)
		}
		if string(got) != "payload-bytes" {
			t.Fatalf("fetch body=%q", got)
		}
	})

	t.Run("non-200", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()
		if _, err := fetch(srv.URL); err == nil {
			t.Fatal("expected error on 500")
		}
	})

	t.Run("ssrf rejected before request", func(t *testing.T) {
		// Even with loopback allowed for IPs, a private 10.x URL is rejected.
		if _, err := fetch("http://10.0.0.1/x"); err == nil {
			t.Fatal("expected SSRF rejection")
		}
	})
}
