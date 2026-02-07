package downloader

import (
	"context"

	"github.com/inovacc/omni/pkg/video/nethttp"
)

// FormatInfo contains the information needed to download a single format.
// This duplicates only the fields from video.Format needed by downloaders,
// avoiding an import cycle with the parent video package.
type FormatInfo struct {
	URL         string
	ManifestURL string
	Ext         string
	Protocol    string
	HTTPHeaders map[string]string
	Filesize    *int64
}

// ProgressInfo reports download progress.
type ProgressInfo struct {
	Status          string
	Filename        string
	DownloadedBytes int64
	TotalBytes      *int64
	Elapsed         float64
	ETA             *float64
	Speed           *float64
	FragmentIndex   *int
	FragmentCount   *int
}

// ProgressFunc is called during download to report progress.
type ProgressFunc func(ProgressInfo)

// Downloader downloads a single format to disk.
type Downloader interface {
	Download(ctx context.Context, path string, format *FormatInfo, opts Options) error
}

// Options configures download behavior.
type Options struct {
	Client    *nethttp.Client
	Progress  ProgressFunc
	RateLimit int64 // bytes per second (0 = unlimited)
	Retries   int
	Continue  bool // resume partial downloads
	NoPart    bool // don't use .part suffix
	Headers   map[string]string
}

// SelectDownloader returns the appropriate downloader for a format's protocol.
func SelectDownloader(protocol string) Downloader {
	switch protocol {
	case "m3u8", "m3u8_native":
		return &HLSDownloader{}
	default:
		return &HTTPDownloader{}
	}
}
