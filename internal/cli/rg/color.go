package rg

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"golang.org/x/term"
)

// ColorScheme defines the colors used for output
type ColorScheme struct {
	Path      string // Color for file paths
	Line      string // Color for line numbers
	Column    string // Color for column numbers
	Match     string // Color for matched text
	MatchBg   string // Background color for matched text
	Separator string // Color for separators
	Context   string // Color for context lines
}

// ANSI color codes for terminal output.
const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Underline = "\033[4m"

	FgBlack   = "\033[30m"
	FgRed     = "\033[31m"
	FgGreen   = "\033[32m"
	FgYellow  = "\033[33m"
	FgBlue    = "\033[34m"
	FgMagenta = "\033[35m"
	FgCyan    = "\033[36m"
	FgWhite   = "\033[37m"

	FgBrightBlack   = "\033[90m"
	FgBrightRed     = "\033[91m"
	FgBrightGreen   = "\033[92m"
	FgBrightYellow  = "\033[93m"
	FgBrightBlue    = "\033[94m"
	FgBrightMagenta = "\033[95m"
	FgBrightCyan    = "\033[96m"
	FgBrightWhite   = "\033[97m"

	BgBlack   = "\033[40m"
	BgRed     = "\033[41m"
	BgGreen   = "\033[42m"
	BgYellow  = "\033[43m"
	BgBlue    = "\033[44m"
	BgMagenta = "\033[45m"
	BgCyan    = "\033[46m"
	BgWhite   = "\033[47m"
)

// DefaultScheme returns the default color scheme matching ripgrep
func DefaultScheme() ColorScheme {
	return ColorScheme{
		Path:      FgMagenta + Bold,
		Line:      FgGreen,
		Column:    FgGreen,
		Match:     FgRed + Bold,
		Separator: FgCyan,
		Context:   "", // No special color for context
	}
}

// NoColorScheme returns a scheme with no colors (for --color=never)
func NoColorScheme() ColorScheme {
	return ColorScheme{}
}

// ColorMode represents when to use colors
type ColorMode int

const (
	ColorAuto   ColorMode = iota // Use colors if terminal supports it
	ColorAlways                  // Always use colors
	ColorNever                   // Never use colors
)

// ParseColorMode parses a color mode string
func ParseColorMode(s string) ColorMode {
	switch strings.ToLower(s) {
	case "always":
		return ColorAlways
	case "never":
		return ColorNever
	default:
		return ColorAuto
	}
}

// ShouldUseColor determines if colors should be used based on mode and terminal
func ShouldUseColor(mode ColorMode) bool {
	switch mode {
	case ColorAlways:
		return true
	case ColorNever:
		return false
	case ColorAuto:
		return isTerminal()
	}

	return isTerminal()
}

// isTerminal checks if stdout is a terminal
func isTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// Colorize wraps text with color codes
func Colorize(text, color string) string {
	if color == "" {
		return text
	}

	return color + text + Reset
}

// HighlightMatches highlights all regex matches in the line
func HighlightMatches(line string, re *regexp.Regexp, scheme ColorScheme, useColor bool) string {
	if !useColor || re == nil {
		return line
	}

	// Find all match locations
	matches := re.FindAllStringIndex(line, -1)
	if len(matches) == 0 {
		return line
	}

	// Build highlighted string
	var result strings.Builder

	lastEnd := 0

	for _, loc := range matches {
		start, end := loc[0], loc[1]

		// Add non-matching part
		result.WriteString(line[lastEnd:start])

		// Add highlighted match
		result.WriteString(scheme.Match)
		result.WriteString(line[start:end])
		result.WriteString(Reset)

		lastEnd = end
	}

	// Add remaining text
	result.WriteString(line[lastEnd:])

	return result.String()
}

// HighlightLiteralMatches highlights all literal string matches in the line
func HighlightLiteralMatches(line, pattern string, caseInsensitive bool, scheme ColorScheme, useColor bool) string {
	if !useColor || pattern == "" {
		return line
	}

	searchLine := line
	searchPattern := pattern

	if caseInsensitive {
		searchLine = strings.ToLower(line)
		searchPattern = strings.ToLower(pattern)
	}

	// Find all match locations
	var matches [][]int

	offset := 0
	for {
		idx := strings.Index(searchLine[offset:], searchPattern)
		if idx == -1 {
			break
		}

		start := offset + idx
		end := start + len(pattern)
		matches = append(matches, []int{start, end})
		offset = end
	}

	if len(matches) == 0 {
		return line
	}

	// Build highlighted string
	var result strings.Builder

	lastEnd := 0

	for _, loc := range matches {
		start, end := loc[0], loc[1]

		// Add non-matching part
		result.WriteString(line[lastEnd:start])

		// Add highlighted match
		result.WriteString(scheme.Match)
		result.WriteString(line[start:end])
		result.WriteString(Reset)

		lastEnd = end
	}

	// Add remaining text
	result.WriteString(line[lastEnd:])

	return result.String()
}

// FormatPath formats a file path with colors
func FormatPath(path string, scheme ColorScheme, useColor bool) string {
	if useColor && scheme.Path != "" {
		return scheme.Path + path + Reset
	}

	return path
}

// FormatLineNumber formats a line number with colors
func FormatLineNumber(lineNum int, scheme ColorScheme, useColor bool) string {
	s := fmt.Sprintf("%d", lineNum)
	if useColor && scheme.Line != "" {
		return scheme.Line + s + Reset
	}

	return s
}

// FormatColumn formats a column number with colors
func FormatColumn(col int, scheme ColorScheme, useColor bool) string {
	s := fmt.Sprintf("%d", col)
	if useColor && scheme.Column != "" {
		return scheme.Column + s + Reset
	}

	return s
}

// FormatSeparator formats a separator with colors
func FormatSeparator(sep string, scheme ColorScheme, useColor bool) string {
	if useColor && scheme.Separator != "" {
		return scheme.Separator + sep + Reset
	}

	return sep
}

// ParseColorSpec parses a ripgrep-style color specification like "path:fg:magenta"
func ParseColorSpec(spec string) (component, attr, value string, err error) {
	parts := strings.SplitN(spec, ":", 3)
	if len(parts) < 2 {
		return "", "", "", fmt.Errorf("invalid color spec: %s", spec)
	}

	component = parts[0]

	attr = parts[1]
	if len(parts) > 2 {
		value = parts[2]
	}

	return component, attr, value, nil
}

// ColorNameToCode converts a color name to ANSI code
func ColorNameToCode(name string, isBg bool) string {
	name = strings.ToLower(name)

	prefix := "3" // foreground
	if isBg {
		prefix = "4" // background
	}

	switch name {
	case "black":
		return "\033[" + prefix + "0m"
	case "red":
		return "\033[" + prefix + "1m"
	case "green":
		return "\033[" + prefix + "2m"
	case "yellow":
		return "\033[" + prefix + "3m"
	case "blue":
		return "\033[" + prefix + "4m"
	case "magenta":
		return "\033[" + prefix + "5m"
	case "cyan":
		return "\033[" + prefix + "6m"
	case "white":
		return "\033[" + prefix + "7m"
	default:
		return ""
	}
}

// ApplyColorSpec applies a color specification to a scheme
func ApplyColorSpec(scheme *ColorScheme, spec string) error {
	component, attr, value, err := ParseColorSpec(spec)
	if err != nil {
		return err
	}

	var code string

	switch attr {
	case "fg":
		code = ColorNameToCode(value, false)
	case "bg":
		code = ColorNameToCode(value, true)
	case "style":
		switch value {
		case "bold":
			code = Bold
		case "underline":
			code = Underline
		case "nobold", "nounderline":
			code = ""
		}
	case "none":
		code = ""
	default:
		return fmt.Errorf("unknown color attribute: %s", attr)
	}

	switch component {
	case "path":
		scheme.Path = code
	case "line":
		scheme.Line = code
	case "column":
		scheme.Column = code
	case "match":
		scheme.Match = code
	default:
		return fmt.Errorf("unknown color component: %s", component)
	}

	return nil
}
