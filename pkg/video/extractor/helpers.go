package extractor

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/inovacc/omni/pkg/video/types"
)

// ExtractJSONFromHTML finds and parses JSON embedded in HTML (e.g., in script tags).
func ExtractJSONFromHTML(html string, patterns []string) (map[string]any, bool) {
	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}

		m := re.FindStringSubmatch(html)
		if len(m) < 2 {
			continue
		}

		var result map[string]any
		if err := json.Unmarshal([]byte(m[1]), &result); err == nil {
			return result, true
		}
	}

	return nil, false
}

// ExtractVideoTags finds <video> and <source> tags in HTML and returns URLs.
func ExtractVideoTags(html, baseURL string) []types.Format {
	formats := make([]types.Format, 0, 4)

	// Find <video src="...">
	videoSrcRe := regexp.MustCompile(`(?i)<video[^>]+src=["']([^"']+)["']`)
	for _, m := range videoSrcRe.FindAllStringSubmatch(html, -1) {
		formats = append(formats, makeDirectFormat(m[1], baseURL, len(formats)))
	}

	// Find <source src="..." type="...">
	sourceSrcRe := regexp.MustCompile(`(?i)<source[^>]+src=["']([^"']+)["'][^>]*(?:type=["']([^"']*)["'])?`)
	for _, m := range sourceSrcRe.FindAllStringSubmatch(html, -1) {
		f := makeDirectFormat(m[1], baseURL, len(formats))
		if len(m) > 2 && m[2] != "" {
			f.FormatNote = m[2]
			if strings.Contains(m[2], "webm") {
				f.Ext = "webm"
			}
		}

		formats = append(formats, f)
	}

	return formats
}

// ExtractIframeSources finds iframe src URLs that might contain embedded videos.
func ExtractIframeSources(html string) []string {
	re := regexp.MustCompile(`(?i)<iframe[^>]+src=["']([^"']+)["']`)

	var urls []string

	for _, m := range re.FindAllStringSubmatch(html, -1) {
		src := m[1]
		// Filter for common video embed patterns.
		if isVideoEmbed(src) {
			urls = append(urls, src)
		}
	}

	return urls
}

func isVideoEmbed(url string) bool {
	embedPatterns := []string{
		"youtube.com/embed",
		"player.vimeo.com",
		"dailymotion.com/embed",
		"facebook.com/plugins/video",
		"instagram.com/p/",
		"streamable.com/",
		"twitch.tv/",
	}

	lower := strings.ToLower(url)
	for _, p := range embedPatterns {
		if strings.Contains(lower, p) {
			return true
		}
	}

	return false
}

func makeDirectFormat(url, baseURL string, idx int) types.Format {
	if baseURL != "" && !strings.HasPrefix(url, "http") {
		if strings.HasPrefix(url, "//") {
			url = "https:" + url
		} else {
			// Simple join.
			if !strings.HasSuffix(baseURL, "/") && !strings.HasPrefix(url, "/") {
				baseURL += "/"
			}

			url = baseURL + url
		}
	}

	ext := guessExtFromURL(url)

	return types.Format{
		URL:      url,
		FormatID: formatID("direct", idx),
		Ext:      ext,
		Protocol: "https",
	}
}

func guessExtFromURL(url string) string {
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

func formatID(prefix string, idx int) string {
	if idx == 0 {
		return prefix + "-0"
	}
	// Simple int to string.
	digits := make([]byte, 0, 4)

	n := idx
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}

	return prefix + "-" + string(digits)
}
