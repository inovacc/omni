package video

import (
	"context"
	"fmt"
	"io"

	"github.com/inovacc/omni/pkg/video"
	"github.com/inovacc/omni/pkg/video/utils"
)

// RunDownload downloads a video from the given URL.
func RunDownload(w io.Writer, args []string, opts Options) error {
	if len(args) == 0 {
		return fmt.Errorf("video download: URL is required")
	}

	url := normalizeVideoURL(args[0])

	rateLimit, _ := utils.ParseFilesize(opts.RateLimit)

	// --complete forces best format and writes a markdown sidecar.
	if opts.Complete {
		opts.Format = "best"
	}

	clientOpts := []video.Option{
		video.WithFormat(orDefault(opts.Format, "best")),
		video.WithRetries(max(opts.Retries, 3)),
	}

	if opts.Output != "" {
		clientOpts = append(clientOpts, video.WithOutput(opts.Output))
	}

	if opts.Quiet {
		clientOpts = append(clientOpts, video.WithQuiet())
	}

	if opts.NoProgress {
		clientOpts = append(clientOpts, video.WithNoProgress())
	}

	if rateLimit > 0 {
		clientOpts = append(clientOpts, video.WithRateLimit(rateLimit))
	}

	if opts.Continue {
		clientOpts = append(clientOpts, video.WithContinue())
	}

	if opts.NoPart {
		clientOpts = append(clientOpts, video.WithNoPart())
	}

	if opts.CookieFile != "" {
		clientOpts = append(clientOpts, video.WithCookieFile(opts.CookieFile))
	}

	if opts.Proxy != "" {
		clientOpts = append(clientOpts, video.WithProxy(opts.Proxy))
	}

	if opts.WriteInfoJSON {
		clientOpts = append(clientOpts, video.WithWriteInfo())
	}

	if opts.WriteSubs {
		clientOpts = append(clientOpts, video.WithWriteSubs())
	}

	if opts.Complete {
		clientOpts = append(clientOpts, video.WithWriteMarkdown())
	}

	if opts.NoPlaylist {
		clientOpts = append(clientOpts, video.WithNoPlaylist())
	}

	if opts.Verbose {
		clientOpts = append(clientOpts, video.WithVerbose())
	}

	if opts.CookiesFromBrowser {
		clientOpts = append(clientOpts, video.WithCookiesFromBrowser())
	}

	// Add progress callback.
	progressFn := MakeProgressFunc(w, opts.Quiet || opts.NoProgress)
	if progressFn != nil {
		clientOpts = append(clientOpts, video.WithProgress(progressFn))
	}

	client, err := video.New(clientOpts...)
	if err != nil {
		return fmt.Errorf("video download: %w", err)
	}

	ctx := context.Background()

	// Extract info first for display.
	info, err := client.Extract(ctx, url)
	if err != nil {
		return err
	}

	if !opts.Quiet {
		_, _ = fmt.Fprintf(w, "[%s] %s: Downloading video\n", info.Extractor, info.ID)
		_, _ = fmt.Fprintf(w, "[download] %s\n", info.Title)
	}

	return client.DownloadInfo(ctx, info)
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}

	return s
}
