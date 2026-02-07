package video

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"github.com/inovacc/omni/pkg/video/downloader"
	"github.com/inovacc/omni/pkg/video/extractor"
	_ "github.com/inovacc/omni/pkg/video/extractor/all" // Register all extractors.
	"github.com/inovacc/omni/pkg/video/format"
	"github.com/inovacc/omni/pkg/video/nethttp"
	"github.com/inovacc/omni/pkg/video/utils"
)

// Client is the main video download orchestrator.
type Client struct {
	opts   Options
	client *nethttp.Client
}

// New creates a new video download client.
func New(opts ...Option) (*Client, error) {
	o := applyOptions(opts)

	httpOpts := nethttp.ClientOptions{
		Proxy:      o.Proxy,
		CookieFile: o.CookieFile,
		UserAgent:  o.UserAgent,
		Headers:    o.Headers,
		Retries:    o.Retries,
	}

	httpClient, err := nethttp.NewClient(httpOpts)
	if err != nil {
		return nil, fmt.Errorf("video: %w", err)
	}

	return &Client{
		opts:   o,
		client: httpClient,
	}, nil
}

// Extract retrieves video metadata from the given URL.
func (c *Client) Extract(ctx context.Context, url string) (*VideoInfo, error) {
	ext, ok := extractor.Match(url)
	if !ok {
		return nil, &UnsupportedError{URL: url}
	}

	info, err := ext.Extract(ctx, url, c.client)
	if err != nil {
		return nil, err
	}

	return c.processResult(ctx, info)
}

// Download extracts and downloads video from the given URL.
func (c *Client) Download(ctx context.Context, url string) error {
	info, err := c.Extract(ctx, url)
	if err != nil {
		return err
	}

	return c.DownloadInfo(ctx, info)
}

// DownloadInfo downloads a video from previously extracted info.
func (c *Client) DownloadInfo(ctx context.Context, info *VideoInfo) error {
	// Handle playlist.
	if info.Type == "playlist" && len(info.Entries) > 0 {
		return c.downloadPlaylist(ctx, info)
	}

	// Select format.
	if len(info.Formats) == 0 {
		return fmt.Errorf("video: no formats available for %s", info.ID)
	}

	selector := format.NewSelector(c.opts.Format)

	selected, err := selector.Select(info.Formats)
	if err != nil {
		return err
	}

	f := &selected[0]

	// Build filename.
	filename := c.buildFilename(info, f)

	// Write info JSON if requested.
	if c.opts.WriteInfo {
		if err := c.writeInfoJSON(info, filename); err != nil {
			return err
		}
	}

	// Download.
	dl := downloader.SelectDownloader(f.Protocol)

	// Bridge progress callback from video.ProgressFunc to downloader.ProgressFunc.
	var dlProgress downloader.ProgressFunc
	if c.opts.Progress != nil {
		dlProgress = func(p downloader.ProgressInfo) {
			c.opts.Progress(ProgressInfo{
				Status:          p.Status,
				Filename:        p.Filename,
				DownloadedBytes: p.DownloadedBytes,
				TotalBytes:      p.TotalBytes,
				Elapsed:         p.Elapsed,
				ETA:             p.ETA,
				Speed:           p.Speed,
				FragmentIndex:   p.FragmentIndex,
				FragmentCount:   p.FragmentCount,
			})
		}
	}

	dlOpts := downloader.Options{
		Client:    c.client,
		Progress:  dlProgress,
		RateLimit: c.opts.RateLimit,
		Retries:   c.opts.Retries,
		Continue:  c.opts.Continue,
		NoPart:    c.opts.NoPart,
		Headers:   c.mergeHeaders(f.HTTPHeaders),
	}

	// Convert video.Format to downloader.FormatInfo.
	dlFormat := &downloader.FormatInfo{
		URL:         f.URL,
		ManifestURL: f.ManifestURL,
		Ext:         f.Ext,
		Protocol:    f.Protocol,
		HTTPHeaders: f.HTTPHeaders,
		Filesize:    f.Filesize,
	}

	return dl.Download(ctx, filename, dlFormat, dlOpts)
}

// ListExtractors returns the names of all registered extractors.
func ListExtractors() []string {
	return extractor.Names()
}

func (c *Client) processResult(ctx context.Context, info *VideoInfo) (*VideoInfo, error) {
	if info == nil {
		return nil, fmt.Errorf("video: nil result")
	}

	// Handle URL redirection (extractor returned a URL to re-extract).
	if info.Type == "url" && info.URL != "" {
		return c.Extract(ctx, info.URL)
	}

	// Handle transparent URL (merge with parent info).
	if info.Type == "url_transparent" && info.URL != "" {
		child, err := c.Extract(ctx, info.URL)
		if err != nil {
			return nil, err
		}
		// Merge parent info into child.
		if child.Title == "" {
			child.Title = info.Title
		}

		return child, nil
	}

	// Handle playlist.
	if info.Type == "playlist" && c.opts.NoPlaylist && len(info.Entries) > 0 {
		// Return first entry only.
		first := info.Entries[0]
		if first.Type == "url" {
			return c.Extract(ctx, first.URL)
		}

		return &first, nil
	}

	return info, nil
}

func (c *Client) downloadPlaylist(ctx context.Context, info *VideoInfo) error {
	entries := info.Entries

	// Apply playlist range.
	start := 0
	if c.opts.PlaylistStart > 0 {
		start = c.opts.PlaylistStart - 1
	}

	end := len(entries)
	if c.opts.PlaylistEnd > 0 && c.opts.PlaylistEnd < end {
		end = c.opts.PlaylistEnd
	}

	if start >= len(entries) {
		return nil
	}

	entries = entries[start:end]

	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if entry.Type == "url" && entry.URL != "" {
			if err := c.Download(ctx, entry.URL); err != nil {
				return fmt.Errorf("playlist entry %s: %w", entry.ID, err)
			}
		}
	}

	return nil
}

func (c *Client) buildFilename(info *VideoInfo, f *Format) string {
	if c.opts.Output != "" {
		return expandOutputTemplate(c.opts.Output, info, f)
	}

	title := info.Title
	if title == "" {
		title = info.ID
	}

	title = utils.SanitizeFilename(title, false)

	ext := f.Ext
	if ext == "" {
		ext = "mp4"
	}

	return title + "." + ext
}

func expandOutputTemplate(tmpl string, info *VideoInfo, f *Format) string {
	replacer := strings.NewReplacer(
		"%(id)s", info.ID,
		"%(title)s", utils.SanitizeFilename(info.Title, false),
		"%(ext)s", orDefault(f.Ext, "mp4"),
		"%(uploader)s", utils.SanitizeFilename(info.Uploader, false),
		"%(upload_date)s", info.UploadDate,
		"%(channel)s", utils.SanitizeFilename(info.Channel, false),
		"%(format_id)s", f.FormatID,
		"%(resolution)s", f.FormatResolution(),
	)
	result := replacer.Replace(tmpl)

	// Ensure extension.
	if !strings.Contains(result, ".") {
		ext := f.Ext
		if ext == "" {
			ext = "mp4"
		}

		result += "." + ext
	}

	return result
}

func (c *Client) writeInfoJSON(info *VideoInfo, videoPath string) error {
	jsonPath := strings.TrimSuffix(videoPath, filepath.Ext(videoPath)) + ".info.json"

	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("video: marshal info: %w", err)
	}

	return os.WriteFile(jsonPath, data, 0o644)
}

func (c *Client) mergeHeaders(formatHeaders map[string]string) map[string]string {
	merged := make(map[string]string)
	maps.Copy(merged, c.opts.Headers)
	maps.Copy(merged, formatHeaders)

	return merged
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}

	return s
}
