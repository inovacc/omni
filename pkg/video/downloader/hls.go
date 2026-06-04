package downloader

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/inovacc/omni/pkg/video/m3u8"
	"github.com/inovacc/omni/pkg/video/utils"
)

// maxSegmentBytes caps how many bytes are read from a single HLS segment body.
// Segment URLs come verbatim from an untrusted M3U8 manifest; without a ceiling
// a hostile or MITM'd host can stream an unbounded (or transparently gunzipped)
// body and exhaust memory. 256 MiB comfortably exceeds any legitimate media
// segment while bounding the worst case. It is a var (not a const) only so tests
// can lower it without allocating hundreds of MiB.
var maxSegmentBytes int64 = 256 << 20 // 256 MiB

// HLSDownloader downloads HLS/M3U8 streams.
type HLSDownloader struct{}

// Download downloads an HLS stream by fetching and combining all segments.
func (d *HLSDownloader) Download(ctx context.Context, path string, format *FormatInfo, opts Options) error {
	if opts.Client == nil {
		return fmt.Errorf("hls: HTTP client is required")
	}

	// Fetch the M3U8 manifest.
	// format.URL is the media playlist URL (resolved from master during extraction).
	manifestURL := format.URL

	manifestBody, err := opts.Client.GetString(ctx, manifestURL)
	if err != nil {
		return fmt.Errorf("hls: fetching manifest: %w", err)
	}

	playlist, err := m3u8.Parse(manifestBody)
	if err != nil {
		return fmt.Errorf("hls: parsing manifest: %w", err)
	}

	// If we got a master playlist, select the best matching variant.
	if playlist.Type == m3u8.PlaylistMaster {
		variantURL := selectVariant(playlist, manifestURL)
		if variantURL == "" {
			return fmt.Errorf("hls: no suitable variant found in master playlist")
		}

		// variantURL is derived from the (untrusted) master manifest; guard against SSRF.
		if err := validateFetchURL(variantURL); err != nil {
			return fmt.Errorf("hls: variant URL: %w", err)
		}

		manifestURL = variantURL

		manifestBody, err = opts.Client.GetString(ctx, manifestURL)
		if err != nil {
			return fmt.Errorf("hls: fetching media playlist: %w", err)
		}

		playlist, err = m3u8.Parse(manifestBody)
		if err != nil {
			return fmt.Errorf("hls: parsing media playlist: %w", err)
		}
	}

	if len(playlist.Segments) == 0 {
		return fmt.Errorf("hls: no segments found")
	}

	// Check for resume state.
	stateFile := path + ".omni-dl"
	startFragment := 0

	if opts.Continue {
		if state, err := LoadFragmentState(stateFile); err == nil {
			startFragment = state.LastFragment + 1
		}
	}

	// Create/open output file.
	flags := os.O_WRONLY | os.O_CREATE
	if startFragment > 0 {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}

	outFile, err := os.OpenFile(path, flags, 0o644)
	if err != nil {
		return fmt.Errorf("hls: open output: %w", err)
	}

	defer func() { _ = outFile.Close() }()

	totalSegments := len(playlist.Segments)
	tracker := NewSpeedTracker(20)
	startTime := time.Now()

	var totalDownloaded int64

	for i := startFragment; i < totalSegments; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		seg := playlist.Segments[i]
		segURL := utils.URLJoin(manifestURL, seg.URL)

		data, err := d.downloadSegment(ctx, opts, segURL, seg.Key)
		if err != nil {
			return fmt.Errorf("hls: segment %d: %w", i, err)
		}

		if _, err := outFile.Write(data); err != nil {
			return fmt.Errorf("hls: write segment %d: %w", i, err)
		}

		totalDownloaded += int64(len(data))
		tracker.Add(totalDownloaded)

		// Save state for resume.
		_ = SaveFragmentState(stateFile, &FragmentState{
			TotalFragments: totalSegments,
			LastFragment:   i,
			Filename:       path,
		})

		// Report progress.
		if opts.Progress != nil {
			fragIdx := i
			fragCount := totalSegments
			opts.Progress(ProgressInfo{
				Status:          "downloading",
				Filename:        path,
				DownloadedBytes: totalDownloaded,
				Elapsed:         time.Since(startTime).Seconds(),
				Speed:           tracker.Speed(),
				FragmentIndex:   &fragIdx,
				FragmentCount:   &fragCount,
			})
		}
	}

	// Clean up state file.
	RemoveFragmentState(stateFile)

	// Final progress.
	if opts.Progress != nil {
		opts.Progress(ProgressInfo{
			Status:          "finished",
			Filename:        path,
			DownloadedBytes: totalDownloaded,
			Elapsed:         time.Since(startTime).Seconds(),
			Speed:           tracker.Speed(),
		})
	}

	return nil
}

func (d *HLSDownloader) downloadSegment(ctx context.Context, opts Options, segURL string, key *m3u8.Key) ([]byte, error) {
	// segURL is derived from the (untrusted) manifest; guard against SSRF.
	if err := validateFetchURL(segURL); err != nil {
		return nil, fmt.Errorf("segment URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, segURL, nil)
	if err != nil {
		return nil, err
	}

	for k, v := range opts.Headers {
		req.Header.Set(k, v)
	}

	resp, err := opts.Client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// Bound the per-segment read: segURL is attacker-controllable (it comes
	// verbatim from the untrusted M3U8), so a bare io.ReadAll would let a hostile
	// or MITM'd host stream an unbounded (or transparently gunzipped) body and
	// exhaust memory. Read one byte past the cap to detect overflow and fail
	// (rather than silently truncating).
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxSegmentBytes+1))
	if err != nil {
		return nil, err
	}

	if int64(len(data)) > maxSegmentBytes {
		return nil, fmt.Errorf("hls: segment body exceeds %d byte limit", maxSegmentBytes)
	}

	// Decrypt if AES-128.
	if key != nil && key.Method == "AES-128" {
		data, err = d.decryptAES128(ctx, opts, data, key)
		if err != nil {
			return nil, fmt.Errorf("decrypt: %w", err)
		}
	}

	return data, nil
}

func (d *HLSDownloader) decryptAES128(ctx context.Context, opts Options, data []byte, key *m3u8.Key) ([]byte, error) {
	// key.URI comes verbatim from the (untrusted) EXT-X-KEY tag; guard against SSRF.
	if err := validateFetchURL(key.URI); err != nil {
		return nil, fmt.Errorf("key URI: %w", err)
	}

	// Fetch the key.
	keyData, err := opts.Client.GetJSON(ctx, key.URI)
	if err != nil {
		return nil, fmt.Errorf("fetching key: %w", err)
	}

	if len(keyData) != 16 {
		return nil, fmt.Errorf("invalid key length: %d", len(keyData))
	}

	block, err := aes.NewCipher(keyData)
	if err != nil {
		return nil, err
	}

	// Parse IV or use default (segment sequence number as big-endian).
	iv := make([]byte, aes.BlockSize)

	if key.IV != "" {
		ivStr := key.IV
		if len(ivStr) > 2 && ivStr[:2] == "0x" {
			ivStr = ivStr[2:]
		}

		decoded, err := hex.DecodeString(ivStr)
		if err != nil {
			return nil, fmt.Errorf("invalid IV: %w", err)
		}

		copy(iv, decoded)
	}

	// CBC decrypt.
	if len(data)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("ciphertext not multiple of block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(data, data)

	// PKCS7 unpadding.
	data = pkcs7Unpad(data)

	return data, nil
}

// selectVariant picks the best variant from a master playlist.
// It selects the variant with the highest bandwidth.
func selectVariant(playlist *m3u8.Playlist, baseURL string) string {
	if len(playlist.Variants) == 0 {
		return ""
	}

	best := playlist.Variants[0]
	for _, v := range playlist.Variants[1:] {
		if v.Bandwidth > best.Bandwidth {
			best = v
		}
	}

	return utils.URLJoin(baseURL, best.URL)
}

// allowLoopbackFetch, when true, permits fetching loopback addresses
// (127.0.0.0/8, ::1). It defaults to false so production callers are protected
// against SSRF. It exists so in-process tests backed by httptest servers (which
// bind to loopback) can exercise the download paths. Non-loopback private,
// link-local and metadata ranges remain blocked regardless of this toggle.
var allowLoopbackFetch = false

// validateFetchURL guards against SSRF for URLs taken from untrusted HLS
// manifests (variant, segment and key URIs). It requires an http/https scheme
// and rejects any host that resolves exclusively to private, loopback,
// link-local, or otherwise non-public address ranges (e.g. the cloud metadata
// endpoint 169.254.169.254). The default behavior of fetching ordinary public
// URLs is unchanged.
func validateFetchURL(raw string) error {
	if raw == "" {
		return fmt.Errorf("empty URL")
	}

	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("unsupported scheme %q (only http/https allowed)", u.Scheme)
	}

	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("missing host")
	}

	// If the host is a literal IP, validate it directly.
	if ip := net.ParseIP(host); ip != nil {
		if !isPublicIP(ip) {
			return fmt.Errorf("host %q resolves to a non-public address", host)
		}

		return nil
	}

	// Otherwise resolve the hostname and reject if it maps only to
	// non-public addresses.
	addrs, err := net.LookupIP(host)
	if err != nil {
		return fmt.Errorf("resolving host %q: %w", host, err)
	}

	if len(addrs) == 0 {
		return fmt.Errorf("host %q resolved to no addresses", host)
	}

	for _, ip := range addrs {
		if !isPublicIP(ip) {
			return fmt.Errorf("host %q resolves to a non-public address", host)
		}
	}

	return nil
}

// isPublicIP reports whether ip is a routable, public address. It rejects
// loopback, link-local (including the 169.254.0.0/16 cloud metadata range),
// private, unspecified, and multicast addresses.
func isPublicIP(ip net.IP) bool {
	if ip.IsLoopback() && allowLoopbackFetch {
		return true
	}

	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsPrivate() || ip.IsUnspecified() || ip.IsMulticast() || ip.IsInterfaceLocalMulticast() {
		return false
	}

	// Reject IPv4-mapped/embedded representations of the above.
	if v4 := ip.To4(); v4 != nil {
		// 100.64.0.0/10 (carrier-grade NAT / shared address space).
		if v4[0] == 100 && v4[1]&0xc0 == 64 {
			return false
		}
	}

	return true
}

func pkcs7Unpad(data []byte) []byte {
	if len(data) == 0 {
		return data
	}

	padLen := int(data[len(data)-1])
	if padLen > len(data) || padLen > aes.BlockSize {
		return data
	}

	for i := len(data) - padLen; i < len(data); i++ {
		if data[i] != byte(padLen) {
			return data
		}
	}

	return data[:len(data)-padLen]
}
