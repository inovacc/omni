package video

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/inovacc/omni/pkg/userdirs"
	"github.com/inovacc/omni/pkg/video"
	"github.com/inovacc/omni/pkg/video/utils"
)

// RunChannel downloads all videos from a YouTube channel with SQLite tracking.
func RunChannel(w io.Writer, args []string, opts Options) error {
	if len(args) == 0 {
		return fmt.Errorf("video channel: URL is required")
	}

	url := normalizeVideoURL(args[0])
	ctx := context.Background()

	// Build client for extraction.
	var clientOpts []video.Option
	if opts.CookieFile != "" {
		clientOpts = append(clientOpts, video.WithCookieFile(opts.CookieFile))
	}

	if opts.Proxy != "" {
		clientOpts = append(clientOpts, video.WithProxy(opts.Proxy))
	}

	if opts.Verbose {
		clientOpts = append(clientOpts, video.WithVerbose())
	}

	if opts.CookiesFromBrowser {
		clientOpts = append(clientOpts, video.WithCookiesFromBrowser())
	}

	extractClient, err := video.New(clientOpts...)
	if err != nil {
		return fmt.Errorf("video channel: %w", err)
	}

	// Extract channel info.
	if !opts.Quiet {
		_, _ = fmt.Fprintf(w, "[channel] Extracting channel info...\n")
	}

	info, err := extractClient.Extract(ctx, url)
	if err != nil {
		return fmt.Errorf("video channel: %w", err)
	}

	if info.Type != "playlist" || len(info.Entries) == 0 {
		return fmt.Errorf("video channel: no videos found in channel")
	}

	channelName := info.Title
	if channelName == "" {
		channelName = info.ID
	}

	// Create channel folder in Downloads.
	downloadDir, err := userdirs.DownloadsDir()
	if err != nil {
		return fmt.Errorf("video channel: %w", err)
	}

	channelDir := filepath.Join(downloadDir, utils.SanitizeFilename(channelName, false))

	if err := os.MkdirAll(channelDir, 0o755); err != nil {
		return fmt.Errorf("video channel: create directory: %w", err)
	}

	// Init SQLite.
	dbPath := filepath.Join(channelDir, "channel.db")
	db, err := initChannelDB(dbPath)
	if err != nil {
		return fmt.Errorf("video channel: %w", err)
	}
	defer func() { _ = db.Close() }()

	// Upsert channel metadata.
	if err := upsertChannel(db, info); err != nil {
		return fmt.Errorf("video channel: %w", err)
	}

	// Get already-downloaded video IDs.
	seen, err := getDownloadedVideoIDs(db)
	if err != nil {
		return fmt.Errorf("video channel: %w", err)
	}

	// Filter entries.
	var newEntries []video.VideoInfo
	for _, entry := range info.Entries {
		if _, exists := seen[entry.ID]; exists {
			continue
		}

		newEntries = append(newEntries, entry)
	}

	alreadyDownloaded := len(info.Entries) - len(newEntries)
	pendingCount := len(newEntries)

	// Apply limit (-1 = all).
	if opts.Limit >= 0 && len(newEntries) > opts.Limit {
		newEntries = newEntries[:opts.Limit]
	}

	if !opts.Quiet {
		if opts.Limit >= 0 && pendingCount > opts.Limit {
			_, _ = fmt.Fprintf(w, "[channel] %s: %d total, %d pending, %d queued (limit %d), %d already downloaded\n",
				channelName, len(info.Entries), pendingCount, len(newEntries), opts.Limit, alreadyDownloaded)
		} else {
			_, _ = fmt.Fprintf(w, "[channel] %s: %d total, %d new, %d already downloaded\n",
				channelName, len(info.Entries), len(newEntries), alreadyDownloaded)
		}
	}

	if len(newEntries) == 0 {
		if !opts.Quiet {
			_, _ = fmt.Fprintf(w, "[channel] Nothing new to download.\n")
		}

		return nil
	}

	// Download each new video.
	var (
		downloaded int
		errCount   int
	)

	rateLimit, _ := utils.ParseFilesize(opts.RateLimit)

	for i, entry := range newEntries {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if !opts.Quiet {
			_, _ = fmt.Fprintf(w, "[%d/%d] Downloading: %s\n", i+1, len(newEntries), entry.Title)
		}

		// Build per-video client with output in channel dir.
		dlOpts := []video.Option{
			video.WithFormat("best"),
			video.WithOutput(filepath.Join(channelDir, "%(title)s.%(ext)s")),
			video.WithRetries(max(opts.Retries, 3)),
			video.WithWriteMarkdown(),
		}

		if opts.Quiet {
			dlOpts = append(dlOpts, video.WithQuiet())
		}

		if opts.NoProgress {
			dlOpts = append(dlOpts, video.WithNoProgress())
		}

		if rateLimit > 0 {
			dlOpts = append(dlOpts, video.WithRateLimit(rateLimit))
		}

		if opts.CookieFile != "" {
			dlOpts = append(dlOpts, video.WithCookieFile(opts.CookieFile))
		}

		if opts.Proxy != "" {
			dlOpts = append(dlOpts, video.WithProxy(opts.Proxy))
		}

		if opts.Verbose {
			dlOpts = append(dlOpts, video.WithVerbose())
		}

		if opts.CookiesFromBrowser {
			dlOpts = append(dlOpts, video.WithCookiesFromBrowser())
		}

		progressFn := MakeProgressFunc(w, opts.Quiet || opts.NoProgress)
		if progressFn != nil {
			dlOpts = append(dlOpts, video.WithProgress(progressFn))
		}

		dlClient, dlErr := video.New(dlOpts...)
		if dlErr != nil {
			_, _ = fmt.Fprintf(w, "[error] %s: %v\n", entry.Title, dlErr)
			errCount++

			continue
		}

		// Extract and download the individual video.
		videoInfo, dlErr := dlClient.Extract(ctx, entry.URL)
		if dlErr != nil {
			_, _ = fmt.Fprintf(w, "[error] %s: %v\n", entry.Title, dlErr)
			errCount++

			continue
		}

		if dlErr = dlClient.DownloadInfo(ctx, videoInfo); dlErr != nil {
			_, _ = fmt.Fprintf(w, "[error] %s: %v\n", entry.Title, dlErr)
			errCount++

			continue
		}

		// Record successful download in DB.
		if dbErr := insertVideoRecord(db, videoInfo, info.ID); dbErr != nil {
			_, _ = fmt.Fprintf(w, "[warn] DB insert failed for %s: %v\n", entry.ID, dbErr)
		}

		downloaded++
	}

	if !opts.Quiet {
		_, _ = fmt.Fprintf(w, "[channel] Done: %d downloaded, %d skipped, %d errors\n",
			downloaded, alreadyDownloaded, errCount)
	}

	return nil
}
