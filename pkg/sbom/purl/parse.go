package purl

import "strings"

// Parse decomposes a Go package-URL ("pkg:golang/<module>[@<version>]") into its
// module path and version. ok is false for any purl whose type is not "golang".
// version is the substring after the last '@' (empty if absent). It is the
// inverse of ForModule for canonical inputs.
func Parse(s string) (modulePath, version string, ok bool) {
	const prefix = "pkg:golang/"
	if !strings.HasPrefix(s, prefix) {
		return "", "", false
	}
	body := s[len(prefix):]
	if at := strings.LastIndex(body, "@"); at >= 0 {
		return body[:at], body[at+1:], true
	}
	return body, "", true
}
