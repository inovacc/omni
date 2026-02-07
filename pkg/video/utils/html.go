package utils

import (
	"regexp"
	"strings"
)

var (
	htmlTagRe     = regexp.MustCompile(`<[^>]+>`)
	htmlCommentRe = regexp.MustCompile(`<!--[\s\S]*?-->`)
	ogPropertyRe  = regexp.MustCompile(`(?i)<meta[^>]+property=["']og:(\w+)["'][^>]+content=["']([^"']*)["'][^>]*/?>`)
	ogPropertyRe2 = regexp.MustCompile(`(?i)<meta[^>]+content=["']([^"']*)["'][^>]+property=["']og:(\w+)["'][^>]*/?>`)
	metaNameRe    = regexp.MustCompile(`(?i)<meta[^>]+name=["']([^"']*)["'][^>]+content=["']([^"']*)["'][^>]*/?>`)
	metaNameRe2   = regexp.MustCompile(`(?i)<meta[^>]+content=["']([^"']*)["'][^>]+name=["']([^"']*)["'][^>]*/?>`)
	htmlEntityRe  = regexp.MustCompile(`&(#?[a-zA-Z0-9]+);`)
)

// htmlEntities maps common HTML entity names to their characters.
var htmlEntities = map[string]string{
	"amp":  "&",
	"lt":   "<",
	"gt":   ">",
	"quot": "\"",
	"apos": "'",
	"nbsp": " ",
}

// CleanHTML strips HTML tags and decodes entities.
func CleanHTML(s string) string {
	s = htmlCommentRe.ReplaceAllString(s, "")
	s = htmlTagRe.ReplaceAllString(s, "")
	s = UnescapeHTML(s)

	return strings.TrimSpace(s)
}

// UnescapeHTML decodes HTML entities.
func UnescapeHTML(s string) string {
	return htmlEntityRe.ReplaceAllStringFunc(s, func(match string) string {
		entity := match[1 : len(match)-1] // Strip & and ;
		if entity[0] == '#' {
			// Numeric entity.
			var num int

			if len(entity) > 1 && (entity[1] == 'x' || entity[1] == 'X') {
				// Hex.
				for _, c := range entity[2:] {
					num *= 16

					switch {
					case c >= '0' && c <= '9':
						num += int(c - '0')
					case c >= 'a' && c <= 'f':
						num += int(c-'a') + 10
					case c >= 'A' && c <= 'F':
						num += int(c-'A') + 10
					}
				}
			} else {
				// Decimal.
				for _, c := range entity[1:] {
					if c >= '0' && c <= '9' {
						num = num*10 + int(c-'0')
					}
				}
			}

			if num > 0 && num < 0x10FFFF {
				return string(rune(num))
			}

			return match
		}

		if repl, ok := htmlEntities[entity]; ok {
			return repl
		}

		return match
	})
}

// OGSearchProperty searches for an OpenGraph meta property value in HTML.
func OGSearchProperty(html, prop string) string {
	// Try property="og:X" content="Y"
	for _, m := range ogPropertyRe.FindAllStringSubmatch(html, -1) {
		if strings.EqualFold(m[1], prop) {
			return UnescapeHTML(m[2])
		}
	}
	// Try content="Y" property="og:X"
	for _, m := range ogPropertyRe2.FindAllStringSubmatch(html, -1) {
		if strings.EqualFold(m[2], prop) {
			return UnescapeHTML(m[1])
		}
	}

	return ""
}

// HTMLSearchMeta searches for a <meta name="X"> content value.
func HTMLSearchMeta(html, name string) string {
	// Try name="X" content="Y"
	for _, m := range metaNameRe.FindAllStringSubmatch(html, -1) {
		if strings.EqualFold(m[1], name) {
			return UnescapeHTML(m[2])
		}
	}
	// Try content="Y" name="X"
	for _, m := range metaNameRe2.FindAllStringSubmatch(html, -1) {
		if strings.EqualFold(m[2], name) {
			return UnescapeHTML(m[1])
		}
	}

	return ""
}

// OGSearchTitle searches for og:title in HTML.
func OGSearchTitle(html string) string {
	return OGSearchProperty(html, "title")
}

// OGSearchDescription searches for og:description in HTML.
func OGSearchDescription(html string) string {
	return OGSearchProperty(html, "description")
}

// OGSearchVideoURL searches for og:video or og:video:url in HTML.
func OGSearchVideoURL(html string) string {
	if url := OGSearchProperty(html, "video"); url != "" {
		return url
	}

	return OGSearchProperty(html, "video:url")
}

// OGSearchThumbnail searches for og:image in HTML.
func OGSearchThumbnail(html string) string {
	return OGSearchProperty(html, "image")
}

// ExtractAttributes extracts all attributes from a single HTML tag string.
func ExtractAttributes(tag string) map[string]string {
	attrs := make(map[string]string)

	re := regexp.MustCompile(`(\w[\w-]*)=["']([^"']*)["']`)
	for _, m := range re.FindAllStringSubmatch(tag, -1) {
		attrs[strings.ToLower(m[1])] = UnescapeHTML(m[2])
	}
	// Handle unquoted attributes.
	re2 := regexp.MustCompile(`(\w[\w-]*)=(\S+)`)
	for _, m := range re2.FindAllStringSubmatch(tag, -1) {
		key := strings.ToLower(m[1])
		if _, exists := attrs[key]; !exists {
			attrs[key] = UnescapeHTML(m[2])
		}
	}

	return attrs
}

// SearchHTMLTag finds a specific HTML tag and returns its outer HTML.
func SearchHTMLTag(html, tag string) []string {
	re := regexp.MustCompile(`(?i)<` + tag + `\b[^>]*>`)
	return re.FindAllString(html, -1)
}

// HiddenInputs extracts name/value pairs from hidden input fields.
func HiddenInputs(html string) map[string]string {
	result := make(map[string]string)

	re := regexp.MustCompile(`(?i)<input[^>]+type=["']hidden["'][^>]*/?>`)
	for _, tag := range re.FindAllString(html, -1) {
		attrs := ExtractAttributes(tag)
		if name, ok := attrs["name"]; ok {
			result[name] = attrs["value"]
		}
	}

	return result
}
