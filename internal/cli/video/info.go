package video

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/inovacc/omni/pkg/video"
)

// RunInfo extracts and displays video metadata.
func RunInfo(w io.Writer, args []string, opts Options) error {
	if len(args) == 0 {
		return fmt.Errorf("video info: URL is required")
	}

	url := normalizeVideoURL(args[0])

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

	client, err := video.New(clientOpts...)
	if err != nil {
		return fmt.Errorf("video info: %w", err)
	}

	info, err := client.Extract(context.Background(), url)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("video info: %w", err)
	}

	_, _ = fmt.Fprintln(w, string(data))

	return nil
}
