package downloader

import (
	"testing"
)

func TestSelectDownloader(t *testing.T) {
	tests := []struct {
		name     string
		protocol string
		wantType string
	}{
		{"m3u8", "m3u8", "*downloader.HLSDownloader"},
		{"m3u8_native", "m3u8_native", "*downloader.HLSDownloader"},
		{"http", "http", "*downloader.HTTPDownloader"},
		{"https", "https", "*downloader.HTTPDownloader"},
		{"empty", "", "*downloader.HTTPDownloader"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := SelectDownloader(tt.protocol)

			switch tt.wantType {
			case "*downloader.HLSDownloader":
				if _, ok := d.(*HLSDownloader); !ok {
					t.Errorf("SelectDownloader(%q) returned %T, want *HLSDownloader", tt.protocol, d)
				}
			case "*downloader.HTTPDownloader":
				if _, ok := d.(*HTTPDownloader); !ok {
					t.Errorf("SelectDownloader(%q) returned %T, want *HTTPDownloader", tt.protocol, d)
				}
			}
		})
	}
}
