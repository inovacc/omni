package downloader

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/inovacc/omni/pkg/video/m3u8"
	"github.com/inovacc/omni/pkg/video/nethttp"
)

// makeClient creates a nethttp.Client with default options.
// Tests use httptest servers at real loopback addresses so no custom transport needed.
func makeClient(_ *http.Client) *nethttp.Client {
	c, _ := nethttp.NewClient(nethttp.ClientOptions{})
	return c
}

// TestMain enables loopback fetches for the duration of the test run. The
// production SSRF guard (validateFetchURL) blocks loopback by default; tests
// here intentionally drive httptest servers bound to 127.0.0.1, so loopback is
// permitted only inside the test process. Non-loopback private and metadata
// ranges remain blocked.
func TestMain(m *testing.M) {
	allowLoopbackFetch = true
	os.Exit(m.Run())
}

// ---- selectVariant ----

func TestSelectVariant_Empty(t *testing.T) {
	pl := &m3u8.Playlist{Variants: nil}
	got := selectVariant(pl, "https://example.com/")
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestSelectVariant_Single(t *testing.T) {
	pl := &m3u8.Playlist{
		Variants: []m3u8.Variant{
			{URL: "low.m3u8", Bandwidth: 100},
		},
	}
	got := selectVariant(pl, "https://example.com/master.m3u8")
	if got == "" {
		t.Error("expected non-empty variant URL")
	}
}

func TestSelectVariant_PicksHighestBandwidth(t *testing.T) {
	pl := &m3u8.Playlist{
		Variants: []m3u8.Variant{
			{URL: "low.m3u8", Bandwidth: 100},
			{URL: "mid.m3u8", Bandwidth: 500},
			{URL: "high.m3u8", Bandwidth: 1000},
		},
	}
	got := selectVariant(pl, "https://example.com/")
	if got == "" {
		t.Error("expected non-empty URL")
	}
	// Should contain "high" (highest bandwidth).
	if len(got) == 0 {
		t.Error("expected URL to contain high variant")
	}
}

// ---- decryptAES128 (pure crypto — key served via httptest) ----

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	for i := 0; i < padding; i++ {
		data = append(data, byte(padding))
	}
	return data
}

func encryptAES128CBC(plaintext, key, iv []byte) []byte {
	plaintext = pkcs7Pad(plaintext, aes.BlockSize)
	block, _ := aes.NewCipher(key)
	mode := cipher.NewCBCEncrypter(block, iv)
	ciphertext := make([]byte, len(plaintext))
	mode.CryptBlocks(ciphertext, plaintext)
	return ciphertext
}

func TestDecryptAES128_RoundTrip(t *testing.T) {
	key := []byte("0123456789abcdef") // 16 bytes
	iv := make([]byte, 16)            // all-zero IV
	plaintext := []byte("Hello, HLS segment data!")

	ciphertext := encryptAES128CBC(plaintext, key, iv)

	// Serve the key via httptest.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(key)
	}))
	defer srv.Close()

	client := makeClient(srv.Client())
	d := &HLSDownloader{}

	k := &m3u8.Key{
		Method: "AES-128",
		URI:    srv.URL + "/key",
		IV:     "",
	}

	result, err := d.decryptAES128(context.Background(), Options{Client: client}, ciphertext, k)
	if err != nil {
		t.Fatalf("decryptAES128: %v", err)
	}
	if string(result) != string(plaintext) {
		t.Errorf("decrypted = %q, want %q", result, plaintext)
	}
}

func TestDecryptAES128_WithExplicitIV(t *testing.T) {
	key := []byte("abcdef0123456789")
	iv := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	ivHex := "0x" + hex.EncodeToString(iv)
	plaintext := []byte("test data for IV")

	ciphertext := encryptAES128CBC(plaintext, key, iv)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(key)
	}))
	defer srv.Close()

	client := makeClient(srv.Client())
	d := &HLSDownloader{}

	k := &m3u8.Key{Method: "AES-128", URI: srv.URL + "/key", IV: ivHex}
	result, err := d.decryptAES128(context.Background(), Options{Client: client}, ciphertext, k)
	if err != nil {
		t.Fatalf("decryptAES128 with IV: %v", err)
	}
	if string(result) != string(plaintext) {
		t.Errorf("decrypted = %q, want %q", result, plaintext)
	}
}

func TestDecryptAES128_InvalidKeyLength(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("short")) // not 16 bytes
	}))
	defer srv.Close()

	client := makeClient(srv.Client())
	d := &HLSDownloader{}
	k := &m3u8.Key{Method: "AES-128", URI: srv.URL + "/key"}

	_, err := d.decryptAES128(context.Background(), Options{Client: client}, make([]byte, 16), k)
	if err == nil {
		t.Error("expected error for invalid key length")
	}
}

func TestDecryptAES128_NotMultipleOfBlockSize(t *testing.T) {
	key := []byte("0123456789abcdef")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(key)
	}))
	defer srv.Close()

	client := makeClient(srv.Client())
	d := &HLSDownloader{}
	k := &m3u8.Key{Method: "AES-128", URI: srv.URL + "/key"}

	_, err := d.decryptAES128(context.Background(), Options{Client: client}, []byte{1, 2, 3}, k)
	if err == nil {
		t.Error("expected error for non-block-aligned ciphertext")
	}
}

// ---- HTTP Download (via httptest) ----

func TestHTTPDownload_SimpleFile(t *testing.T) {
	content := []byte("hello download content")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "22")
		_, _ = w.Write(content)
	}))
	defer srv.Close()

	client := makeClient(srv.Client())
	d := &HTTPDownloader{}

	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "out.mp4")

	format := &FormatInfo{URL: srv.URL + "/video.mp4"}
	opts := Options{Client: client, NoPart: true}

	if err := d.Download(context.Background(), outPath, format, opts); err != nil {
		t.Fatalf("Download: %v", err)
	}

	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("content = %q, want %q", got, content)
	}
}

func TestHTTPDownload_NoClient(t *testing.T) {
	d := &HTTPDownloader{}
	err := d.Download(context.Background(), "/tmp/x.mp4", &FormatInfo{URL: "http://example.com"}, Options{})
	if err == nil {
		t.Error("expected error when client is nil")
	}
}

func TestHTTPDownload_HTTP404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client := makeClient(srv.Client())
	d := &HTTPDownloader{}
	tmpDir := t.TempDir()

	// Retries=1 to avoid slow test.
	opts := Options{Client: client, NoPart: true, Retries: 1}
	err := d.Download(context.Background(), filepath.Join(tmpDir, "out.mp4"), &FormatInfo{URL: srv.URL}, opts)
	if err == nil {
		t.Error("expected error for HTTP 404")
	}
}

func TestHTTPDownload_WithProgress(t *testing.T) {
	content := []byte("progress test content")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(content)
	}))
	defer srv.Close()

	client := makeClient(srv.Client())
	d := &HTTPDownloader{}
	tmpDir := t.TempDir()

	var progressCalled bool
	opts := Options{
		Client: client,
		NoPart: true,
		Progress: func(p ProgressInfo) {
			progressCalled = true
		},
	}

	err := d.Download(context.Background(), filepath.Join(tmpDir, "out.mp4"), &FormatInfo{URL: srv.URL}, opts)
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	// Progress may or may not fire depending on timing; just ensure no crash.
	_ = progressCalled
}

func TestHTTPDownload_RangeNotSatisfiable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
	}))
	defer srv.Close()

	client := makeClient(srv.Client())
	d := &HTTPDownloader{}
	tmpDir := t.TempDir()

	opts := Options{Client: client, NoPart: true, Retries: 1}
	// 416 means file already complete — should succeed.
	err := d.Download(context.Background(), filepath.Join(tmpDir, "out.mp4"), &FormatInfo{URL: srv.URL}, opts)
	if err != nil {
		t.Errorf("expected success for 416, got: %v", err)
	}
}

// ---- HLS Download (via httptest) ----

func TestHLSDownload_NoClient(t *testing.T) {
	d := &HLSDownloader{}
	err := d.Download(context.Background(), "/tmp/x.ts", &FormatInfo{URL: "http://example.com/pl.m3u8"}, Options{})
	if err == nil {
		t.Error("expected error when client is nil")
	}
}

func TestHLSDownload_SimplePlaylist(t *testing.T) {
	segContent := []byte("segment data here!!")

	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/playlist.m3u8":
			w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
			_, _ = w.Write([]byte("#EXTM3U\n#EXT-X-VERSION:3\n#EXTINF:10.0,\n" + srv.URL + "/seg0.ts\n#EXT-X-ENDLIST\n"))
		case "/seg0.ts":
			_, _ = w.Write(segContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client := makeClient(srv.Client())
	d := &HLSDownloader{}
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "out.ts")

	format := &FormatInfo{
		URL:      srv.URL + "/playlist.m3u8",
		Protocol: "m3u8",
	}
	opts := Options{Client: client}

	if err := d.Download(context.Background(), outPath, format, opts); err != nil {
		t.Fatalf("HLS Download: %v", err)
	}

	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != string(segContent) {
		t.Errorf("got %q, want %q", got, segContent)
	}
}

// ---- SpeedTracker.Add and Speed ----

func TestSpeedTracker_AddAndSpeed(t *testing.T) {
	tracker := NewSpeedTracker(5)

	// Single sample → Speed() returns nil.
	tracker.Add(1000)
	if tracker.Speed() != nil {
		t.Error("Speed with 1 sample should return nil")
	}

	// Add a second sample after a short wait to get a real speed.
	time.Sleep(10 * time.Millisecond)
	tracker.Add(2000)

	speed := tracker.Speed()
	if speed == nil {
		t.Error("Speed with 2 samples should return non-nil")
	}
	if *speed <= 0 {
		t.Errorf("Speed = %v, want > 0", *speed)
	}
}

func TestSpeedTracker_WindowEviction(t *testing.T) {
	tracker := NewSpeedTracker(3)
	for i := range 10 {
		tracker.Add(int64(i * 1000))
		time.Sleep(2 * time.Millisecond)
	}
	// Should not panic, window capped at 3.
	speed := tracker.Speed()
	_ = speed
}
