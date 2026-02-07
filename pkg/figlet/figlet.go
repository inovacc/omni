package figlet

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Font represents a parsed FIGlet font.
type Font struct {
	Hardblank    rune
	Height       int
	Baseline     int
	MaxLength    int
	SmushMode    int
	CommentLines int
	Characters   map[rune][]string
}

// config holds rendering options.
type config struct {
	fontName string
	font     *Font
	width    int
}

// Option configures the renderer.
type Option func(*config)

// WithFont sets the font by name (from embedded fonts).
func WithFont(name string) Option {
	return func(c *config) { c.fontName = name }
}

// WithLoadedFont uses a pre-loaded font directly.
func WithLoadedFont(f *Font) Option {
	return func(c *config) { c.font = f }
}

// WithWidth sets the maximum output width. 0 means unlimited.
func WithWidth(w int) Option {
	return func(c *config) { c.width = w }
}

// Render renders text as ASCII art and returns it as a single string.
func Render(text string, opts ...Option) (string, error) {
	lines, err := RenderLines(text, opts...)
	if err != nil {
		return "", err
	}
	return strings.Join(lines, "\n"), nil
}

// RenderLines renders text as ASCII art and returns the individual lines.
func RenderLines(text string, opts ...Option) ([]string, error) {
	cfg := config{fontName: "standard"}
	for _, o := range opts {
		o(&cfg)
	}

	f := cfg.font
	if f == nil {
		var err error
		f, err = LoadEmbedded(cfg.fontName)
		if err != nil {
			return nil, fmt.Errorf("figlet: load font %q: %w", cfg.fontName, err)
		}
	}

	return renderText(f, text, cfg.width), nil
}

// LoadFont parses FIGlet font data from raw bytes.
func LoadFont(data []byte) (*Font, error) {
	return parseFont(string(data))
}

// LoadFontFile reads and parses a FIGlet font from a file path.
func LoadFontFile(path string) (*Font, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("figlet: read font file: %w", err)
	}
	return LoadFont(data)
}

// parseFont parses the FIGlet font format.
func parseFont(data string) (*Font, error) {
	lines := strings.Split(data, "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("figlet: empty font data")
	}

	f, err := parseHeader(lines[0])
	if err != nil {
		return nil, err
	}

	// Skip header + comment lines
	charStart := 1 + f.CommentLines
	if charStart >= len(lines) {
		return nil, fmt.Errorf("figlet: font has no character data")
	}

	f.Characters = make(map[rune][]string)

	// Parse required ASCII characters (32-126)
	pos := charStart
	for ch := rune(32); ch <= 126; ch++ {
		charLines, newPos, err := readCharacter(lines, pos, f.Height)
		if err != nil {
			return nil, fmt.Errorf("figlet: char %d (%c): %w", ch, ch, err)
		}
		f.Characters[ch] = charLines
		pos = newPos
	}

	// Parse optional extended characters (code-tagged)
	for pos < len(lines) {
		line := strings.TrimSpace(lines[pos])
		if line == "" {
			pos++
			continue
		}

		// Extended characters have a code tag line before the character data
		code, err := parseCodeTag(line)
		if err != nil {
			break // No more valid code tags
		}

		pos++
		charLines, newPos, err := readCharacter(lines, pos, f.Height)
		if err != nil {
			break
		}
		f.Characters[rune(code)] = charLines
		pos = newPos
	}

	return f, nil
}

// parseHeader parses the FIGlet header line.
// Format: "flf2a<hardblank> height baseline maxlen smushmode commentlines direction fullwidth"
func parseHeader(line string) (*Font, error) {
	if !strings.HasPrefix(line, "flf2") {
		return nil, fmt.Errorf("figlet: not a FIGlet font (missing flf2 header)")
	}

	// The header format is "flf2a<hardblank> <params...>"
	// After "flf2a" the next character is the hardblank
	if len(line) < 6 {
		return nil, fmt.Errorf("figlet: header too short")
	}

	hardblank := rune(line[5])

	// Split rest of header by spaces
	parts := strings.Fields(line[6:])
	if len(parts) < 5 {
		return nil, fmt.Errorf("figlet: header has too few parameters (got %d, need at least 5)", len(parts))
	}

	height, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("figlet: invalid height: %w", err)
	}

	baseline, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("figlet: invalid baseline: %w", err)
	}

	maxLen, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("figlet: invalid max length: %w", err)
	}

	smushMode, err := strconv.Atoi(parts[3])
	if err != nil {
		return nil, fmt.Errorf("figlet: invalid smush mode: %w", err)
	}

	commentLines, err := strconv.Atoi(parts[4])
	if err != nil {
		return nil, fmt.Errorf("figlet: invalid comment lines: %w", err)
	}

	return &Font{
		Hardblank:    hardblank,
		Height:       height,
		Baseline:     baseline,
		MaxLength:    maxLen,
		SmushMode:    smushMode,
		CommentLines: commentLines,
	}, nil
}

// readCharacter reads one FIGcharacter (Height lines) from the font data.
// Each line ends with '@' (continuation) or '@@' (last line of character).
func readCharacter(lines []string, start, height int) ([]string, int, error) {
	result := make([]string, 0, height)
	pos := start

	for i := 0; i < height; i++ {
		if pos >= len(lines) {
			return nil, pos, fmt.Errorf("unexpected end of font data")
		}

		line := lines[pos]
		pos++

		// Strip end markers: @@ or @
		line = strings.TrimRight(line, "\r")
		if strings.HasSuffix(line, "@@") {
			line = line[:len(line)-2]
		} else if strings.HasSuffix(line, "@") {
			line = line[:len(line)-1]
		}

		result = append(result, line)
	}

	return result, pos, nil
}

// parseCodeTag extracts the character code from an extended character tag line.
// Formats: "196" (decimal), "0x00C4" (hex), "0304" (octal)
func parseCodeTag(line string) (int, error) {
	// The code tag can have a trailing comment after whitespace
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return 0, fmt.Errorf("empty code tag")
	}

	tag := parts[0]

	if strings.HasPrefix(tag, "0x") || strings.HasPrefix(tag, "0X") {
		val, err := strconv.ParseInt(tag[2:], 16, 32)
		return int(val), err
	}

	if strings.HasPrefix(tag, "0") && len(tag) > 1 {
		val, err := strconv.ParseInt(tag[1:], 8, 32)
		return int(val), err
	}

	val, err := strconv.Atoi(tag)
	if err != nil {
		return 0, err
	}

	// Negative values signal end of font
	if val < 0 {
		return 0, fmt.Errorf("negative code tag: %d", val)
	}

	return val, nil
}

// renderText renders a string using the given font.
func renderText(f *Font, text string, maxWidth int) []string {
	if len(text) == 0 {
		return nil
	}

	result := make([]string, f.Height)

	for _, ch := range text {
		charLines, ok := f.Characters[ch]
		if !ok {
			// Use space for unknown characters
			charLines = f.Characters[' ']
			if charLines == nil {
				// Fallback: empty character of font height
				charLines = make([]string, f.Height)
			}
		}

		for row := 0; row < f.Height; row++ {
			line := ""
			if row < len(charLines) {
				line = charLines[row]
			}
			// Replace hardblank with space
			line = strings.ReplaceAll(line, string(f.Hardblank), " ")
			result[row] += line
		}
	}

	// Apply width limit if set
	if maxWidth > 0 {
		for i, line := range result {
			if len(line) > maxWidth {
				result[i] = line[:maxWidth]
			}
		}
	}

	// Trim trailing whitespace from each line
	for i, line := range result {
		result[i] = strings.TrimRight(line, " ")
	}

	return result
}
