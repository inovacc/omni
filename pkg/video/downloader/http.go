package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// HTTPDownloader downloads via plain HTTP/HTTPS with resume support.
type HTTPDownloader struct{}

// Download downloads a single format via HTTP.
func (d *HTTPDownloader) Download(ctx context.Context, path string, format *FormatInfo, opts Options) error {
	if opts.Client == nil {
		return fmt.Errorf("download: HTTP client is required")
	}

	partPath := path + ".part"
	if opts.NoPart {
		partPath = path
	}

	// Check existing partial download for resume.
	var resumeOffset int64

	if opts.Continue && !opts.NoPart {
		if info, err := os.Stat(partPath); err == nil {
			resumeOffset = info.Size()
		}
	}

	retries := opts.Retries
	if retries <= 0 {
		retries = 3
	}

	var lastErr error

	for attempt := range retries {
		err := d.downloadAttempt(ctx, partPath, format, opts, resumeOffset)
		if err == nil {
			// Rename .part to final.
			if partPath != path {
				if err := os.Rename(partPath, path); err != nil {
					return fmt.Errorf("download: rename: %w", err)
				}
			}

			return nil
		}

		lastErr = err

		// Check if we got more data for resume.
		if opts.Continue && !opts.NoPart {
			if info, statErr := os.Stat(partPath); statErr == nil {
				resumeOffset = info.Size()
			}
		}

		if attempt < retries-1 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(attempt+1) * 2 * time.Second):
			}
		}
	}

	return fmt.Errorf("download: all %d retries failed: %w", retries, lastErr)
}

func (d *HTTPDownloader) downloadAttempt(ctx context.Context, path string, format *FormatInfo, opts Options, resumeOffset int64) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, format.URL, nil)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}

	// Set headers from format.
	for k, v := range format.HTTPHeaders {
		req.Header.Set(k, v)
	}

	for k, v := range opts.Headers {
		req.Header.Set(k, v)
	}

	if resumeOffset > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", resumeOffset))
	}

	resp, err := opts.Client.Do(req)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	// Handle response codes.
	switch {
	case resp.StatusCode == http.StatusRequestedRangeNotSatisfiable:
		// File already complete.
		return nil
	case resp.StatusCode == http.StatusPartialContent:
		// Resume successful.
	case resp.StatusCode == http.StatusOK:
		// Server doesn't support range, start from beginning.
		resumeOffset = 0
	case resp.StatusCode >= 400:
		return fmt.Errorf("download: HTTP %d", resp.StatusCode)
	}

	// Determine total size.
	var totalBytes *int64

	if resp.ContentLength > 0 {
		total := resp.ContentLength + resumeOffset
		totalBytes = &total
	} else if format.Filesize != nil {
		totalBytes = format.Filesize
	}

	// Open file for writing.
	flags := os.O_WRONLY | os.O_CREATE
	if resumeOffset > 0 {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}

	f, err := os.OpenFile(path, flags, 0o644)
	if err != nil {
		return fmt.Errorf("download: open file: %w", err)
	}

	defer func() { _ = f.Close() }()

	// Setup progress tracking.
	tracker := NewSpeedTracker(20)
	downloaded := resumeOffset
	startTime := time.Now()
	lastProgress := time.Now()

	// Read body with optional rate limiting.
	buf := make([]byte, 64*1024)

	var reader io.Reader = resp.Body

	if opts.RateLimit > 0 {
		reader = newRateLimitedReader(resp.Body, opts.RateLimit)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		n, readErr := reader.Read(buf)
		if n > 0 {
			if _, writeErr := f.Write(buf[:n]); writeErr != nil {
				return fmt.Errorf("download: write: %w", writeErr)
			}

			downloaded += int64(n)
			tracker.Add(downloaded)

			// Report progress at most 10 times per second.
			if opts.Progress != nil && time.Since(lastProgress) > 100*time.Millisecond {
				lastProgress = time.Now()

				opts.Progress(ProgressInfo{
					Status:          "downloading",
					Filename:        path,
					DownloadedBytes: downloaded,
					TotalBytes:      totalBytes,
					Elapsed:         time.Since(startTime).Seconds(),
					Speed:           tracker.Speed(),
					ETA:             tracker.ETA(downloaded, ptrOr(totalBytes, 0)),
				})
			}
		}

		if readErr != nil {
			if readErr == io.EOF {
				break
			}

			return fmt.Errorf("download: read: %w", readErr)
		}
	}

	// Final progress callback.
	if opts.Progress != nil {
		opts.Progress(ProgressInfo{
			Status:          "finished",
			Filename:        path,
			DownloadedBytes: downloaded,
			TotalBytes:      totalBytes,
			Elapsed:         time.Since(startTime).Seconds(),
			Speed:           tracker.Speed(),
		})
	}

	return nil
}

func ptrOr(p *int64, def int64) int64 {
	if p != nil {
		return *p
	}

	return def
}

// rateLimitedReader wraps an io.Reader with rate limiting.
type rateLimitedReader struct {
	r         io.Reader
	limit     int64 // bytes per second
	readBytes int64
	startTime time.Time
}

func newRateLimitedReader(r io.Reader, bytesPerSec int64) *rateLimitedReader {
	return &rateLimitedReader{
		r:         r,
		limit:     bytesPerSec,
		startTime: time.Now(),
	}
}

func (r *rateLimitedReader) Read(p []byte) (int, error) {
	n, err := r.r.Read(p)
	if n > 0 {
		r.readBytes += int64(n)

		elapsed := time.Since(r.startTime).Seconds()
		if elapsed > 0 {
			currentRate := float64(r.readBytes) / elapsed
			if currentRate > float64(r.limit) {
				sleepTime := float64(r.readBytes)/float64(r.limit) - elapsed
				if sleepTime > 0 {
					time.Sleep(time.Duration(sleepTime * float64(time.Second)))
				}
			}
		}
	}

	return n, err
}
