package utils

import (
	"net/url"
	"strings"
)

// URLJoin joins a base URL and a relative path, handling edge cases.
func URLJoin(base, path string) string {
	if path == "" {
		return base
	}
	// Already absolute.
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	// Protocol-relative.
	if strings.HasPrefix(path, "//") {
		if strings.HasPrefix(base, "https://") {
			return "https:" + path
		}

		return "http:" + path
	}

	baseURL, err := url.Parse(base)
	if err != nil {
		return path
	}

	ref, err := url.Parse(path)
	if err != nil {
		return path
	}

	return baseURL.ResolveReference(ref).String()
}

// UpdateURLQuery adds or updates query parameters on a URL.
func UpdateURLQuery(rawURL string, params map[string]string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}

	u.RawQuery = q.Encode()

	return u.String()
}

// SanitizeURL removes tracking parameters and normalizes the URL.
func SanitizeURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	u.Fragment = ""

	return u.String()
}

// ParseQueryString parses a URL query string into a map.
func ParseQueryString(query string) map[string]string {
	result := make(map[string]string)

	values, err := url.ParseQuery(query)
	if err != nil {
		return result
	}

	for k, v := range values {
		if len(v) > 0 {
			result[k] = v[0]
		}
	}

	return result
}

// ExtractQueryParam extracts a single query parameter from a URL.
func ExtractQueryParam(rawURL, param string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	return u.Query().Get(param)
}
