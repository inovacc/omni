package video

import (
	"net/url"
	"strings"
)

// normalizeVideoURL removes timestamp anchors so downloads start from the beginning.
// It only normalizes HTTP(S) URLs and leaves non-URL inputs (e.g. ytsearch:) unchanged.
func normalizeVideoURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return raw
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
