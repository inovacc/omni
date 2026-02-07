package extractor

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/inovacc/omni/pkg/video/format"
	"github.com/inovacc/omni/pkg/video/nethttp"
	"github.com/inovacc/omni/pkg/video/types"
	"github.com/inovacc/omni/pkg/video/utils"
)

// Extractor is the interface all site extractors must implement.
type Extractor interface {
	// Name returns the extractor name (e.g., "YouTube", "Generic").
	Name() string
	// Suitable returns true if this extractor can handle the given URL.
	Suitable(url string) bool
	// Extract fetches video metadata and formats from the URL.
	Extract(ctx context.Context, url string, client *nethttp.Client) (*types.VideoInfo, error)
}

// BaseExtractor provides common helpers for all extractors.
type BaseExtractor struct {
	ExtractorName string
}

// DownloadWebpage fetches a webpage and returns its HTML content.
func (b *BaseExtractor) DownloadWebpage(ctx context.Context, client *nethttp.Client, url string) (string, error) {
	body, err := client.GetString(ctx, url)
	if err != nil {
		return "", &types.ExtractorError{
			Extractor: b.ExtractorName,
			Message:   "failed to download webpage",
			Cause:     err,
		}
	}

	return body, nil
}

// DownloadJSON fetches JSON from a URL and unmarshals it.
func (b *BaseExtractor) DownloadJSON(ctx context.Context, client *nethttp.Client, url string, dst any) error {
	data, err := client.GetJSON(ctx, url)
	if err != nil {
		return &types.ExtractorError{
			Extractor: b.ExtractorName,
			Message:   "failed to download JSON",
			Cause:     err,
		}
	}

	if err := json.Unmarshal(data, dst); err != nil {
		return &types.ExtractorError{
			Extractor: b.ExtractorName,
			Message:   "failed to parse JSON",
			Cause:     err,
		}
	}

	return nil
}

// SearchRegex searches HTML content for a regex pattern and returns the first
// capture group. Returns empty string if not found and fatal is false.
func (b *BaseExtractor) SearchRegex(patterns []string, content, name string, fatal bool) (string, error) {
	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}

		m := re.FindStringSubmatch(content)
		if len(m) > 1 {
			return m[1], nil
		}
	}

	if fatal {
		return "", &types.ExtractorError{
			Extractor: b.ExtractorName,
			Message:   fmt.Sprintf("unable to extract %s", name),
		}
	}

	return "", nil
}

// HTMLSearchMeta searches for a <meta> tag value.
func (b *BaseExtractor) HTMLSearchMeta(html, name string) string {
	return utils.HTMLSearchMeta(html, name)
}

// OGSearchTitle returns the og:title value.
func (b *BaseExtractor) OGSearchTitle(html string) string {
	return utils.OGSearchTitle(html)
}

// OGSearchDescription returns the og:description value.
func (b *BaseExtractor) OGSearchDescription(html string) string {
	return utils.OGSearchDescription(html)
}

// OGSearchVideoURL returns the og:video URL.
func (b *BaseExtractor) OGSearchVideoURL(html string) string {
	return utils.OGSearchVideoURL(html)
}

// OGSearchThumbnail returns the og:image URL.
func (b *BaseExtractor) OGSearchThumbnail(html string) string {
	return utils.OGSearchThumbnail(html)
}

// ParseJSON parses a JSON string into a map.
func (b *BaseExtractor) ParseJSON(jsonStr string) (map[string]any, error) {
	var result map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("%s: invalid JSON: %w", b.ExtractorName, err)
	}

	return result, nil
}

// ExtractM3U8Formats extracts format information from an M3U8 manifest URL.
func (b *BaseExtractor) ExtractM3U8Formats(ctx context.Context, client *nethttp.Client, m3u8URL string, videoID string) ([]types.Format, error) {
	body, err := client.GetString(ctx, m3u8URL)
	if err != nil {
		return nil, fmt.Errorf("%s: fetching m3u8: %w", b.ExtractorName, err)
	}

	return ParseM3U8Formats(body, m3u8URL, videoID), nil
}

// SortFormats sorts formats by quality.
func (b *BaseExtractor) SortFormats(formats []types.Format) {
	format.SortFormats(formats)
}

// HiddenInputs extracts hidden form fields from HTML.
func (b *BaseExtractor) HiddenInputs(html string) map[string]string {
	return utils.HiddenInputs(html)
}

// ParseM3U8Formats parses M3U8 content into video formats.
func ParseM3U8Formats(body, manifestURL, videoID string) []types.Format {
	lines := strings.Split(strings.TrimSpace(body), "\n")

	var formats []types.Format

	formatIdx := 0

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		if attrStr, ok := strings.CutPrefix(line, "#EXT-X-STREAM-INF:"); ok {
			attrs := parseM3U8Attrs(attrStr)

			// Find URL on next line.
			var streamURL string

			for i++; i < len(lines); i++ {
				next := strings.TrimSpace(lines[i])
				if next != "" && !strings.HasPrefix(next, "#") {
					streamURL = utils.URLJoin(manifestURL, next)
					break
				}
			}

			if streamURL == "" {
				continue
			}

			f := types.Format{
				URL:         streamURL,
				ManifestURL: manifestURL,
				FormatID:    fmt.Sprintf("hls-%d", formatIdx),
				Ext:         "mp4",
				Protocol:    "m3u8_native",
				VCodec:      "avc1",
				ACodec:      "mp4a",
			}

			if res, ok := attrs["RESOLUTION"]; ok {
				f.Resolution = res

				parts := strings.SplitN(res, "x", 2)
				if len(parts) == 2 {
					w, h := parseInt(parts[0]), parseInt(parts[1])
					f.Width = &w
					f.Height = &h
				}
			}

			if bw, ok := attrs["BANDWIDTH"]; ok {
				tbr := float64(parseIntStr(bw)) / 1000.0
				f.TBR = &tbr
			}

			if codecs, ok := attrs["CODECS"]; ok {
				parseCodecs(codecs, &f)
			}

			formats = append(formats, f)
			formatIdx++
		}
	}

	return formats
}

func parseM3U8Attrs(s string) map[string]string {
	attrs := make(map[string]string)

	re := regexp.MustCompile(`([A-Z0-9-]+)=(?:"([^"]*)"|([^,]*))`)
	for _, m := range re.FindAllStringSubmatch(s, -1) {
		key := m[1]

		val := m[2]
		if val == "" {
			val = m[3]
		}

		attrs[key] = val
	}

	return attrs
}

func parseCodecs(codecs string, f *types.Format) {
	for c := range strings.SplitSeq(codecs, ",") {
		c = strings.TrimSpace(c)
		switch {
		case strings.HasPrefix(c, "avc1"), strings.HasPrefix(c, "hvc1"),
			strings.HasPrefix(c, "hev1"), strings.HasPrefix(c, "vp9"),
			strings.HasPrefix(c, "vp09"), strings.HasPrefix(c, "av01"):
			f.VCodec = c
		case strings.HasPrefix(c, "mp4a"), strings.HasPrefix(c, "opus"),
			strings.HasPrefix(c, "vorbis"), strings.HasPrefix(c, "ac-3"),
			strings.HasPrefix(c, "ec-3"), strings.HasPrefix(c, "flac"):
			f.ACodec = c
		}
	}
}

func parseInt(s string) int {
	n := 0

	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}

	return n
}

func parseIntStr(s string) int {
	return parseInt(s)
}
