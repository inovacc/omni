package video

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	pkgvideo "github.com/inovacc/omni/pkg/video"
)

const interactiveMenu = `
Interactive video menu for: %s
  1) Download (best)
  2) Download (custom format selector)
  3) List formats
  4) Show info (JSON)
  5) Nerd stats
  6) List extractors
  7) Change URL
  8) Quit
`

// RunInteractive starts an interactive menu for video operations.
func RunInteractive(out io.Writer, prompt io.Writer, in io.Reader, args []string, opts Options) error {
	reader := bufio.NewReader(in)

	url := ""
	if len(args) > 0 {
		url = strings.TrimSpace(args[0])
	}

	for {
		if url == "" {
			nextURL, err := promptLine(reader, prompt, "Video URL: ")
			if err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}

				return fmt.Errorf("video interactive: %w", err)
			}

			if nextURL == "" {
				_, _ = fmt.Fprintln(prompt, "URL is required.")
				continue
			}

			url = nextURL
		}

		_, _ = fmt.Fprintf(prompt, interactiveMenu, url)

		choice, err := promptLine(reader, prompt, "Select an option: ")
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}

			return fmt.Errorf("video interactive: %w", err)
		}

		switch normalizeChoice(choice) {
		case "1", "download", "d":
			downloadOpts := opts
			if downloadOpts.Format == "" {
				downloadOpts.Format = "best"
			}

			if err := RunDownload(out, []string{url}, downloadOpts); err != nil {
				_, _ = fmt.Fprintf(prompt, "[error] %v\n", err)
			}

		case "2", "download-format", "custom":
			selector, err := promptLine(reader, prompt, "Format selector (best/worst/ID/best[height<=720]): ")
			if err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}

				return fmt.Errorf("video interactive: %w", err)
			}

			if selector == "" {
				_, _ = fmt.Fprintln(prompt, "Format selector is required.")
				_, _ = fmt.Fprintln(prompt)
				continue
			}

			downloadOpts := opts
			downloadOpts.Format = selector

			if err := RunDownload(out, []string{url}, downloadOpts); err != nil {
				_, _ = fmt.Fprintf(prompt, "[error] %v\n", err)
			}

		case "3", "formats", "list-formats":
			listOpts := opts
			listOpts.JSON = false
			if err := RunListFormats(out, []string{url}, listOpts); err != nil {
				_, _ = fmt.Fprintf(prompt, "[error] %v\n", err)
			}

		case "4", "info", "metadata":
			if err := RunInfo(out, []string{url}, opts); err != nil {
				_, _ = fmt.Fprintf(prompt, "[error] %v\n", err)
			}

		case "5", "nerd", "nerds", "stats":
			if err := runNerdStats(out, url, opts); err != nil {
				_, _ = fmt.Fprintf(prompt, "[error] %v\n", err)
			}

		case "6", "extractors", "sites":
			names := pkgvideo.ListExtractors()
			sort.Strings(names)
			_, _ = fmt.Fprintln(out, "Supported extractors:")
			for _, name := range names {
				_, _ = fmt.Fprintf(out, "  - %s\n", name)
			}

		case "7", "change", "url":
			url = ""

		case "8", "q", "quit", "exit":
			_, _ = fmt.Fprintln(prompt, "Exiting interactive mode.")
			return nil

		default:
			_, _ = fmt.Fprintf(prompt, "Unknown option: %q\n", choice)
		}

		_, _ = fmt.Fprintln(prompt)
	}
}

func runNerdStats(out io.Writer, url string, opts Options) error {
	url = normalizeVideoURL(url)

	var clientOpts []pkgvideo.Option
	if opts.CookieFile != "" {
		clientOpts = append(clientOpts, pkgvideo.WithCookieFile(opts.CookieFile))
	}

	if opts.Proxy != "" {
		clientOpts = append(clientOpts, pkgvideo.WithProxy(opts.Proxy))
	}

	if opts.Verbose {
		clientOpts = append(clientOpts, pkgvideo.WithVerbose())
	}

	client, err := pkgvideo.New(clientOpts...)
	if err != nil {
		return fmt.Errorf("video nerd stats: %w", err)
	}

	info, err := client.Extract(context.Background(), url)
	if err != nil {
		return err
	}

	videoOnly := 0
	audioOnly := 0
	muxed := 0
	unknown := 0

	protocolCounts := map[string]int{}
	extensionCounts := map[string]int{}
	containerCounts := map[string]int{}

	for _, f := range info.Formats {
		hasVideo := f.HasVideo()
		hasAudio := f.HasAudio()

		switch {
		case hasVideo && hasAudio:
			muxed++
		case hasVideo:
			videoOnly++
		case hasAudio:
			audioOnly++
		default:
			unknown++
		}

		protocol := f.Protocol
		if protocol == "" {
			protocol = "unknown"
		}
		protocolCounts[protocol]++

		ext := f.Ext
		if ext == "" {
			ext = "unknown"
		}
		extensionCounts[ext]++

		container := f.Container
		if container == "" {
			container = "unknown"
		}
		containerCounts[container]++
	}

	subtitleLangs := make([]string, 0, len(info.Subtitles))
	for lang := range info.Subtitles {
		subtitleLangs = append(subtitleLangs, lang)
	}
	sort.Strings(subtitleLangs)

	videoType := info.Type
	if videoType == "" {
		videoType = "video"
	}

	_, _ = fmt.Fprintf(out, "[%s] %s: nerd stats\n", info.Extractor, info.ID)
	_, _ = fmt.Fprintf(out, "  title: %s\n", info.Title)
	_, _ = fmt.Fprintf(out, "  type: %s\n", videoType)
	_, _ = fmt.Fprintf(out, "  extractor_key: %s\n", orUnknown(info.ExtractorKey))
	_, _ = fmt.Fprintf(out, "  duration_seconds: %.0f\n", info.Duration)
	_, _ = fmt.Fprintf(out, "  formats_total: %d\n", len(info.Formats))
	_, _ = fmt.Fprintf(out, "  streams: muxed=%d video_only=%d audio_only=%d unknown=%d\n", muxed, videoOnly, audioOnly, unknown)
	_, _ = fmt.Fprintf(out, "  playlist_entries: %d\n", len(info.Entries))
	_, _ = fmt.Fprintf(out, "  thumbnails: %d\n", len(info.Thumbnails))
	_, _ = fmt.Fprintf(out, "  chapters: %d\n", len(info.Chapters))
	_, _ = fmt.Fprintf(out, "  subtitle_langs: %d\n", len(subtitleLangs))
	if len(subtitleLangs) > 0 {
		_, _ = fmt.Fprintf(out, "  subtitle_codes: %s\n", strings.Join(subtitleLangs, ", "))
	}

	_, _ = fmt.Fprintf(out, "  protocols: %s\n", renderCounts(protocolCounts))
	_, _ = fmt.Fprintf(out, "  extensions: %s\n", renderCounts(extensionCounts))
	_, _ = fmt.Fprintf(out, "  containers: %s\n", renderCounts(containerCounts))

	limit := len(info.Formats)
	if limit > 5 {
		limit = 5
	}

	if limit > 0 {
		_, _ = fmt.Fprintln(out, "  sample_formats:")
		for i := 0; i < limit; i++ {
			f := info.Formats[i]
			_, _ = fmt.Fprintf(out, "    - %s | %s | v=%s a=%s | %s\n",
				f.FormatID,
				f.FormatResolution(),
				orUnknown(f.VCodec),
				orUnknown(f.ACodec),
				orUnknown(f.Protocol),
			)
		}
	}

	return nil
}

func renderCounts(counts map[string]int) string {
	if len(counts) == 0 {
		return "none"
	}

	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%d", key, counts[key]))
	}

	return strings.Join(parts, ", ")
}

func normalizeChoice(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func orUnknown(s string) string {
	if s == "" {
		return "unknown"
	}

	return s
}

func promptLine(reader *bufio.Reader, w io.Writer, prompt string) (string, error) {
	_, _ = fmt.Fprint(w, prompt)

	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}

	line = strings.TrimSpace(line)

	if errors.Is(err, io.EOF) && line == "" {
		return "", io.EOF
	}

	return line, nil
}
