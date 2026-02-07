package utils

import (
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"unicode"
)

var (
	// Characters not allowed in filenames on most platforms.
	unsafeChars = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)

	// Windows reserved names.
	windowsReserved = map[string]bool{
		"CON": true, "PRN": true, "AUX": true, "NUL": true,
		"COM1": true, "COM2": true, "COM3": true, "COM4": true,
		"COM5": true, "COM6": true, "COM7": true, "COM8": true, "COM9": true,
		"LPT1": true, "LPT2": true, "LPT3": true, "LPT4": true,
		"LPT5": true, "LPT6": true, "LPT7": true, "LPT8": true, "LPT9": true,
	}

	// Multiple spaces/dashes/underscores.
	multipleSpaces = regexp.MustCompile(`\s+`)
)

// SanitizeFilename makes a string safe for use as a filename.
// It removes/replaces characters that are not allowed on the target OS.
func SanitizeFilename(name string, restrictMode bool) string {
	if name == "" {
		return "_"
	}

	// Replace path separators with dash.
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "\\", "-")

	if restrictMode {
		// ASCII-only mode: strip non-ASCII characters.
		var b strings.Builder

		for _, r := range name {
			if r <= 127 {
				b.WriteRune(r)
			}
		}

		name = b.String()
	}

	// Remove unsafe characters.
	name = unsafeChars.ReplaceAllString(name, "")

	// Collapse whitespace.
	name = multipleSpaces.ReplaceAllString(name, " ")
	name = strings.TrimSpace(name)

	// Trim trailing dots and spaces (Windows issue).
	name = strings.TrimRight(name, ". ")

	// Check Windows reserved names.
	if runtime.GOOS == "windows" || restrictMode {
		base := strings.ToUpper(name)
		// Remove extension for check.
		if idx := strings.LastIndex(base, "."); idx > 0 {
			base = base[:idx]
		}

		if windowsReserved[base] {
			name = "_" + name
		}
	}

	if name == "" {
		return "_"
	}

	// Limit filename length (255 bytes is common limit).
	if len(name) > 200 {
		ext := filepath.Ext(name)
		name = name[:200-len(ext)] + ext
	}

	return name
}

// SanitizeFilenameStrict is a stricter version that only allows alphanumeric,
// spaces, hyphens, underscores, and dots.
func SanitizeFilenameStrict(name string) string {
	var b strings.Builder

	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' || r == '-' || r == '_' || r == '.' {
			b.WriteRune(r)
		}
	}

	result := b.String()
	if result == "" {
		return "_"
	}

	return result
}
