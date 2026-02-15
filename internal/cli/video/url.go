package video

import (
	"net/url"
	"regexp"
	"strings"
)

// youtubeIDPattern matches an 11-character YouTube video ID (alphanumeric, dash, underscore).
var youtubeIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{11}$`)

// normalizeVideoURL removes timestamp anchors so downloads start from the beginning.
// It also detects bare 11-character YouTube video IDs and expands them to full URLs.
// It only normalizes HTTP(S) URLs and leaves non-URL inputs (e.g. ytsearch:) unchanged.
func normalizeVideoURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return raw
	}

	// Detect bare YouTube video IDs (exactly 11 chars, no scheme, no slashes)
	if youtubeIDPattern.MatchString(raw) {
		return "https://www.youtube.com/watch?v=" + raw
	}

	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return raw
	}

	q := u.Query()
	q.Del("t")
	u.RawQuery = q.Encode()

	if frag := strings.TrimSpace(u.Fragment); frag != "" {
		switch {
		case strings.HasPrefix(strings.ToLower(frag), "t="):
			u.Fragment = ""
		case strings.Contains(frag, "="):
			// Handle fragments like "t=1m30s&foo=bar".
			fragValues, parseErr := url.ParseQuery(frag)
			if parseErr == nil {
				fragValues.Del("t")
				u.Fragment = fragValues.Encode()
			}
		}
	}

	return u.String()
}
