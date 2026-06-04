package downloader

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestDownloadSegment_RejectsOverLimitBody proves the per-segment read is
// bounded: a segment body larger than the ceiling must return an error (not a
// truncated buffer), guarding against memory-exhaustion / gzip-bomb DoS from an
// untrusted M3U8. The production ceiling (maxSegmentBytes, 256 MiB) is lowered
// here so the test stays fast and never allocates hundreds of MiB.
func TestDownloadSegment_RejectsOverLimitBody(t *testing.T) {
	const testCap = 1024 // 1 KiB test-only ceiling

	orig := maxSegmentBytes
	maxSegmentBytes = testCap
	t.Cleanup(func() { maxSegmentBytes = orig })

	// Serve a body that is one byte past the cap.
	body := bytes.Repeat([]byte("A"), testCap+1)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	client := makeClient(srv.Client())
	d := &HLSDownloader{}

	_, err := d.downloadSegment(context.Background(), Options{Client: client}, srv.URL+"/seg0.ts", nil)
	if err == nil {
		t.Fatal("expected error for segment body exceeding the cap, got nil")
	}

	if !strings.Contains(err.Error(), "exceeds") {
		t.Errorf("expected an exceeds-limit error, got: %v", err)
	}
}

// TestDownloadSegment_AcceptsBodyAtLimit confirms the fix is not over-eager: a
// body exactly at the ceiling is read in full and returned intact.
func TestDownloadSegment_AcceptsBodyAtLimit(t *testing.T) {
	const testCap = 1024

	orig := maxSegmentBytes
	maxSegmentBytes = testCap
	t.Cleanup(func() { maxSegmentBytes = orig })

	body := bytes.Repeat([]byte("B"), testCap)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	client := makeClient(srv.Client())
	d := &HLSDownloader{}

	got, err := d.downloadSegment(context.Background(), Options{Client: client}, srv.URL+"/seg0.ts", nil)
	if err != nil {
		t.Fatalf("unexpected error for body exactly at cap: %v", err)
	}

	if len(got) != testCap {
		t.Errorf("got %d bytes, want %d", len(got), testCap)
	}
}
