package generic

import (
	"context"
	"path"
	"regexp"
	"strings"

	"github.com/inovacc/omni/pkg/video/extractor"
	"github.com/inovacc/omni/pkg/video/nethttp"
	"github.com/inovacc/omni/pkg/video/types"
	"github.com/inovacc/omni/pkg/video/utils"
)

func init() {
	extractor.Register(&GenericExtractor{})
}

var directVideoExts = regexp.MustCompile(`(?i)\.(mp4|webm|flv|mkv|avi|mov|m4v|m4a|mp3|ogg|opus|wav|aac|m3u8)(\?|$)`)

// GenericExtractor is a fallback extractor that handles direct video URLs,
// <video> tags, og:video meta tags, and common embed patterns.
type GenericExtractor struct {
	extractor.BaseExtractor
}

// Name returns the extractor name.
func (e *GenericExtractor) Name() string { return "Generic" }

// Suitable returns true for any HTTP/HTTPS URL.
func (e *GenericExtractor) Suitable(url string) bool {
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

// Extract attempts to find video content at the given URL.
func (e *GenericExtractor) Extract(ctx context.Context, url string, client *nethttp.Client) (*types.VideoInfo, error) {
	e.ExtractorName = "Generic"

	// Check if URL directly points to a video file.
	if directVideoExts.MatchString(url) {
		return e.extractDirect(url), nil
	}

	// Download the webpage.
	html, err := e.DownloadWebpage(ctx, client, url)
	if err != nil {
		return nil, err
	}

	// Try multiple extraction methods.
	info := &types.VideoInfo{
		ID:           extractIDFromURL(url),
		WebpageURL:   url,
		Extractor:    "Generic",
		ExtractorKey: "Generic",
	}

	// Extract title.
	info.Title = extractTitle(html)
	if info.Title == "" {
		info.Title = info.ID
	}

	// Extract description.
	info.Description = utils.OGSearchDescription(html)

	// Extract thumbnail.
	if thumb := utils.OGSearchThumbnail(html); thumb != "" {
		info.Thumbnails = []types.Thumbnail{{URL: utils.URLJoin(url, thumb)}}
	}

	// Try og:video.
	if videoURL := utils.OGSearchVideoURL(html); videoURL != "" {
		videoURL = utils.URLJoin(url, videoURL)
		ext := guessExt(videoURL)
		info.Formats = append(info.Formats, types.Format{
			URL:      videoURL,
			FormatID: "og-video",
			Ext:      ext,
			Protocol: guessProtocol(videoURL),
		})
	}

	// Try <video> and <source> tags.
	videoFormats := extractor.ExtractVideoTags(html, url)
	info.Formats = append(info.Formats, videoFormats...)

	// Try M3U8 URLs embedded in the page.
	m3u8Re := regexp.MustCompile(`["'](https?://[^"']+\.m3u8[^"']*)["']`)
	for _, m := range m3u8Re.FindAllStringSubmatch(html, -1) {
		m3u8URL := utils.UnescapeHTML(m[1])

		hlsFormats, err := e.ExtractM3U8Formats(ctx, client, m3u8URL, info.ID)
		if err == nil {
			info.Formats = append(info.Formats, hlsFormats...)
		}
	}

	// Try iframe embeds — return as "url" type for re-extraction.
	if len(info.Formats) == 0 {
		iframes := extractor.ExtractIframeSources(html)
		if len(iframes) > 0 {
			info.Type = "url"
			info.URL = utils.URLJoin(url, iframes[0])

			return info, nil
		}
	}

	if len(info.Formats) == 0 {
		return nil, &types.UnsupportedError{URL: url}
	}

	e.SortFormats(info.Formats)

	return info, nil
}

func (e *GenericExtractor) extractDirect(url string) *types.VideoInfo {
	ext := guessExt(url)

	return &types.VideoInfo{
		ID:           extractIDFromURL(url),
		Title:        extractIDFromURL(url),
		URL:          url,
		Ext:          ext,
		WebpageURL:   url,
		Extractor:    "Generic",
		ExtractorKey: "Generic",
		Formats: []types.Format{
			{
				URL:      url,
				FormatID: "direct",
				Ext:      ext,
				Protocol: guessProtocol(url),
			},
		},
	}
}

func extractTitle(html string) string {
	// Try og:title first.
	if title := utils.OGSearchTitle(html); title != "" {
		return title
	}
	// Try <title> tag.
	re := regexp.MustCompile(`(?i)<title[^>]*>([^<]+)</title>`)
	if m := re.FindStringSubmatch(html); m != nil {
		return utils.CleanHTML(m[1])
	}

	return ""
}

func extractIDFromURL(rawURL string) string {
	// Use the last path segment as the ID.
	rawURL = strings.TrimRight(rawURL, "/")
	if idx := strings.LastIndex(rawURL, "/"); idx >= 0 {
		id := rawURL[idx+1:]
		// Remove query string.
		if qIdx := strings.Index(id, "?"); qIdx >= 0 {
			id = id[:qIdx]
		}
		// Remove extension.
		if ext := path.Ext(id); ext != "" {
			id = id[:len(id)-len(ext)]
		}

		if id != "" {
			return id
		}
	}

	return "video"
}

func guessExt(url string) string {
	lower := strings.ToLower(url)
	switch {
	case strings.Contains(lower, ".mp4"):
		return "mp4"
	case strings.Contains(lower, ".webm"):
		return "webm"
	case strings.Contains(lower, ".m3u8"):
		return "mp4"
	case strings.Contains(lower, ".flv"):
		return "flv"
	case strings.Contains(lower, ".mkv"):
		return "mkv"
	case strings.Contains(lower, ".m4a"):
		return "m4a"
	case strings.Contains(lower, ".mp3"):
		return "mp3"
	default:
		return "mp4"
	}
}

func guessProtocol(url string) string {
	lower := strings.ToLower(url)
	if strings.Contains(lower, ".m3u8") {
		return "m3u8_native"
	}

	if strings.HasPrefix(lower, "https://") {
		return "https"
	}

	return "http"
}
