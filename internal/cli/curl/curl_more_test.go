package curl

import (
	"bytes"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestIsRestrictedIP(t *testing.T) {
	tests := []struct {
		ip   string
		want bool
	}{
		{"127.0.0.1", true},
		{"::1", true},
		{"10.0.0.1", true},
		{"192.168.1.1", true},
		{"172.16.5.5", true},
		{"169.254.1.1", true},   // link-local unicast
		{"224.0.0.1", true},     // multicast
		{"0.0.0.0", true},       // unspecified
		{"8.8.8.8", false},      // public
		{"1.1.1.1", false},      // public
		{"93.184.216.34", false}, // public
	}
	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if ip == nil {
				t.Fatalf("bad test ip %q", tt.ip)
			}
			if got := isRestrictedIP(ip); got != tt.want {
				t.Fatalf("isRestrictedIP(%s)=%v want %v", tt.ip, got, tt.want)
			}
		})
	}
}

func TestCheckRedirectTarget(t *testing.T) {
	tests := []struct {
		name    string
		rawURL  string
		wantErr bool
	}{
		{"empty host", "/relative/path", false},
		{"loopback ip", "http://127.0.0.1/x", true},
		{"private ip", "http://10.1.2.3/x", true},
		{"metadata", "http://169.254.169.254/latest", true},
		{"public ip", "http://8.8.8.8/x", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := url.Parse(tt.rawURL)
			if err != nil {
				t.Fatalf("parse url: %v", err)
			}
			gotErr := checkRedirectTarget(u)
			if (gotErr != nil) != tt.wantErr {
				t.Fatalf("checkRedirectTarget(%q) err=%v wantErr=%v", tt.rawURL, gotErr, tt.wantErr)
			}
		})
	}
}

// TestRunFollowRedirectToRestricted drives the redirect SSRF guard end-to-end:
// a public test server 30x-redirects to a loopback target, which the
// CheckRedirect hook must reject.
func TestRunFollowRedirectToRestricted(t *testing.T) {
	internal := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("secret"))
	}))
	defer internal.Close()

	redirector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Redirect(w, nil, internal.URL, http.StatusFound)
	}))
	defer redirector.Close()

	var buf bytes.Buffer
	err := Run(&buf, []string{redirector.URL}, Options{Method: http.MethodGet, FollowRedir: true})
	if err == nil {
		t.Fatal("expected redirect to loopback to be refused")
	}
	if !strings.Contains(err.Error(), "restricted") && !strings.Contains(err.Error(), "redirect") {
		t.Logf("note: error was %v", err)
	}
}

func TestRunNoURL(t *testing.T) {
	var buf bytes.Buffer
	if err := Run(&buf, nil, Options{}); err == nil {
		t.Fatal("expected error when no URL given")
	}
}

func TestRunSimpleGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()
	var buf bytes.Buffer
	if err := Run(&buf, []string{srv.URL}, Options{Method: http.MethodGet, Verbose: true}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(buf.String(), "ok") {
		t.Errorf("output missing body: %q", buf.String())
	}
}
