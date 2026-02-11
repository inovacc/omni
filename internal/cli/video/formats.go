package video

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/inovacc/omni/pkg/video"
	"github.com/inovacc/omni/pkg/video/format"
)

// RunListFormats extracts and displays available formats.
func RunListFormats(w io.Writer, args []string, opts Options) error {
	if len(args) == 0 {
		return fmt.Errorf("video list-formats: URL is required")
	}

	url := normalizeVideoURL(args[0])

	var clientOpts []video.Option
	if opts.CookieFile != "" {
		clientOpts = append(clientOpts, video.WithCookieFile(opts.CookieFile))
	}

	if opts.Proxy != "" {
		clientOpts = append(clientOpts, video.WithProxy(opts.Proxy))
	}

	client, err := video.New(clientOpts...)
	if err != nil {
		return fmt.Errorf("video list-formats: %w", err)
	}

	info, err := client.Extract(context.Background(), url)
	if err != nil {
		return err
	}

	format.SortFormats(info.Formats)

	if opts.JSON {
		data, err := json.MarshalIndent(info.Formats, "", "  ")
		if err != nil {
			return err
		}

		_, _ = fmt.Fprintln(w, string(data))

		return nil
	}

	_, _ = fmt.Fprintf(w, "[%s] %s: Available formats\n", info.Extractor, info.ID)
	_, _ = fmt.Fprintln(w, FormatTable(info.Formats))

	return nil
}
