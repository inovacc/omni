package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// WatchOptions configures the watch command behavior
type WatchOptions struct {
	Interval    time.Duration // -n: seconds to wait between updates
	Differences bool          // -d: highlight differences between updates
	NoTitle     bool          // -t: turn off header
	BeepOnError bool          // -b: beep if command has a non-zero exit
	ExitOnError bool          // -e: exit if command has a non-zero exit
	Precise     bool          // -p: attempt to run command at precise intervals
	Color       bool          // -c: interpret ANSI color sequences
}

// WatchFunc is the function type for watch operations
type WatchFunc func() (string, error)

// RunWatch repeatedly executes a function at specified intervals
// Note: omni doesn't exec external commands, so this works with functions
func RunWatch(ctx context.Context, w io.Writer, fn WatchFunc, opts WatchOptions) error {
	if opts.Interval == 0 {
		opts.Interval = 2 * time.Second
	}

	// Handle signals for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(opts.Interval)
	defer ticker.Stop()

	// Run immediately first
	if err := runWatchIteration(w, fn, opts); err != nil {
		if opts.ExitOnError {
			return err
		}

		if opts.BeepOnError {
			_, _ = fmt.Fprint(w, "\a") // Bell character
		}
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-sigCh:
			return nil
		case <-ticker.C:
			if err := runWatchIteration(w, fn, opts); err != nil {
				if opts.ExitOnError {
					return err
				}

				if opts.BeepOnError {
					_, _ = fmt.Fprint(w, "\a")
				}
			}
		}
	}
}

func runWatchIteration(w io.Writer, fn WatchFunc, opts WatchOptions) error {
	// Clear screen
	_, _ = fmt.Fprint(w, "\033[H\033[2J")

	// Print header unless disabled
	if !opts.NoTitle {
		now := time.Now().Format("Mon Jan 2 15:04:05 2006")
		_, _ = fmt.Fprintf(w, "Every %.1fs: watching...\t%s\n\n", opts.Interval.Seconds(), now)
	}

	output, err := fn()
	_, _ = fmt.Fprint(w, output)

	return err
}

// RunWatchCommand is a simplified watch that just prints a message
func RunWatchCommand(ctx context.Context, w io.Writer, message string, opts WatchOptions) error {
	return RunWatch(ctx, w, func() (string, error) {
		return message + "\n", nil
	}, opts)
}

// WatchFile watches a file for changes and calls the callback when modified
func WatchFile(ctx context.Context, path string, callback func(path string) error, interval time.Duration) error {
	if interval == 0 {
		interval = time.Second
	}

	var lastModTime time.Time

	// Get initial mod time
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	lastModTime = info.ModTime()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			info, err := os.Stat(path)
			if err != nil {
				continue
			}

			if info.ModTime().After(lastModTime) {
				lastModTime = info.ModTime()

				if err := callback(path); err != nil {
					return err
				}
			}
		}
	}
}

// WatchDir watches a directory for changes
func WatchDir(ctx context.Context, path string, callback func(event string, path string) error, interval time.Duration) error {
	if interval == 0 {
		interval = time.Second
	}

	// Track file states
	fileStates := make(map[string]time.Time)

	// Initial scan
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		fileStates[entry.Name()] = info.ModTime()
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			entries, err := os.ReadDir(path)
			if err != nil {
				continue
			}

			currentFiles := make(map[string]bool)

			for _, entry := range entries {
				info, err := entry.Info()
				if err != nil {
					continue
				}

				name := entry.Name()
				currentFiles[name] = true
				modTime := info.ModTime()

				if lastMod, exists := fileStates[name]; !exists {
					// New file
					fileStates[name] = modTime
					if err := callback("created", name); err != nil {
						return err
					}
				} else if modTime.After(lastMod) {
					// Modified
					fileStates[name] = modTime
					if err := callback("modified", name); err != nil {
						return err
					}
				}
			}

			// Check for deleted files
			for name := range fileStates {
				if !currentFiles[name] {
					delete(fileStates, name)

					if err := callback("deleted", name); err != nil {
						return err
					}
				}
			}
		}
	}
}
