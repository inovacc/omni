package downloader

import (
	"testing"
)

// ---- Options / FormatInfo / ProgressInfo field construction ----

func TestOptions_Defaults(t *testing.T) {
	opts := Options{}
	if opts.Client != nil {
		t.Error("Client should default to nil")
	}
	if opts.RateLimit != 0 {
		t.Error("RateLimit should default to 0")
	}
	if opts.Retries != 0 {
		t.Error("Retries should default to 0")
	}
	if opts.Continue {
		t.Error("Continue should default to false")
	}
	if opts.NoPart {
		t.Error("NoPart should default to false")
	}
}

func TestFormatInfo_Construction(t *testing.T) {
	size := int64(1024)
	fi := FormatInfo{
		URL:         "https://example.com/video.mp4",
		ManifestURL: "https://example.com/manifest.m3u8",
		Ext:         "mp4",
		Protocol:    "https",
		HTTPHeaders: map[string]string{"User-Agent": "test"},
		Filesize:    &size,
	}

	if fi.URL == "" {
		t.Error("URL should be set")
	}
	if fi.Filesize == nil || *fi.Filesize != 1024 {
		t.Error("Filesize should be 1024")
	}
	if fi.HTTPHeaders["User-Agent"] != "test" {
		t.Error("HTTPHeaders should contain User-Agent")
	}
}

func TestProgressInfo_AllFields(t *testing.T) {
	total := int64(1000)
	eta := 5.0
	speed := 100.0
	fragIdx := 3
	fragCount := 10

	pi := ProgressInfo{
		Status:          "downloading",
		Filename:        "output.mp4",
		DownloadedBytes: 500,
		TotalBytes:      &total,
		Elapsed:         2.5,
		ETA:             &eta,
		Speed:           &speed,
		FragmentIndex:   &fragIdx,
		FragmentCount:   &fragCount,
	}

	if pi.Status != "downloading" {
		t.Errorf("Status = %q", pi.Status)
	}
	if *pi.TotalBytes != 1000 {
		t.Error("TotalBytes mismatch")
	}
	if *pi.FragmentIndex != 3 {
		t.Error("FragmentIndex mismatch")
	}
	if *pi.FragmentCount != 10 {
		t.Error("FragmentCount mismatch")
	}
}

// ---- SelectDownloader extended ----

func TestSelectDownloader_AllProtocols(t *testing.T) {
	tests := []struct {
		protocol string
		wantHLS  bool
	}{
		{"m3u8", true},
		{"m3u8_native", true},
		{"http", false},
		{"https", false},
		{"", false},
		{"ftp", false},
		{"rtmp", false},
	}

	for _, tt := range tests {
		t.Run(tt.protocol, func(t *testing.T) {
			d := SelectDownloader(tt.protocol)
			if d == nil {
				t.Fatal("SelectDownloader returned nil")
			}
			_, isHLS := d.(*HLSDownloader)
			_, isHTTP := d.(*HTTPDownloader)
			if !isHLS && !isHTTP {
				t.Errorf("unknown downloader type %T", d)
			}
			if tt.wantHLS && !isHLS {
				t.Errorf("protocol %q: want HLSDownloader, got %T", tt.protocol, d)
			}
			if !tt.wantHLS && !isHTTP {
				t.Errorf("protocol %q: want HTTPDownloader, got %T", tt.protocol, d)
			}
		})
	}
}

// ---- ProgressFunc type ----

func TestProgressFunc_Callable(t *testing.T) {
	called := false
	var fn ProgressFunc = func(p ProgressInfo) {
		called = true
		if p.Status != "test" {
			t.Errorf("Status = %q, want test", p.Status)
		}
	}

	fn(ProgressInfo{Status: "test"})
	if !called {
		t.Error("ProgressFunc was not called")
	}
}

// ---- ptrOr edge cases ----

func TestPtrOr_Zero(t *testing.T) {
	zero := int64(0)
	if ptrOr(&zero, 99) != 0 {
		t.Error("ptrOr with zero value pointer should return 0, not default")
	}
}

func TestPtrOr_Negative(t *testing.T) {
	neg := int64(-1)
	if ptrOr(&neg, 99) != -1 {
		t.Error("ptrOr with negative pointer should return -1")
	}
}

func TestPtrOr_LargeDefault(t *testing.T) {
	if ptrOr(nil, 9999999) != 9999999 {
		t.Error("ptrOr(nil) should return default")
	}
}

// ---- HTTPDownloader struct zero value ----

func TestHTTPDownloader_ZeroValue(t *testing.T) {
	d := &HTTPDownloader{}
	if d == nil {
		t.Error("HTTPDownloader zero value should not be nil")
	}
	// Verify it implements Downloader interface.
	var _ Downloader = d
}

// ---- HLSDownloader struct zero value ----

func TestHLSDownloader_ZeroValue(t *testing.T) {
	d := &HLSDownloader{}
	if d == nil {
		t.Error("HLSDownloader zero value should not be nil")
	}
	var _ Downloader = d
}
