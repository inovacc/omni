package printf

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// Options configures the printf command behavior
type Options struct {
	NoNewline bool // don't append newline
}

// Run executes the printf command
func Run(w io.Writer, args []string, opts Options) error {
	if len(args) == 0 {
		return fmt.Errorf("printf: missing format string")
	}

	format := args[0]
	formatArgs := args[1:]

	output, err := Format(format, formatArgs)
	if err != nil {
		return fmt.Errorf("printf: %w", err)
	}

	if opts.NoNewline {
		_, _ = fmt.Fprint(w, output)
	} else {
		_, _ = fmt.Fprintln(w, output)
	}

	return nil
}

// Format formats a string using printf-style format specifiers
func Format(format string, args []string) (string, error) {
	// Process escape sequences first
	format = processEscapes(format)

	// Find all format specifiers
	re := regexp.MustCompile(`%[-+#0 ]*(\d+)?(\.\d+)?[diouxXeEfFgGsqcb%]`)
	matches := re.FindAllStringIndex(format, -1)

	if len(matches) == 0 {
		return format, nil
	}

	var result strings.Builder

	argIndex := 0
	lastEnd := 0

	for _, match := range matches {
		start, end := match[0], match[1]

		// Add text before this format specifier
		result.WriteString(format[lastEnd:start])

		spec := format[start:end]

		// Handle %% escape
		if spec == "%%" {
			result.WriteString("%")

			lastEnd = end

			continue
		}

		// Get the argument for this specifier
		var arg string
		if argIndex < len(args) {
			arg = args[argIndex]
			argIndex++
		} else {
			arg = ""
		}

		// Format the argument based on the specifier
		formatted := formatArg(spec, arg)
		result.WriteString(formatted)

		lastEnd = end
	}

	// Add remaining text
	result.WriteString(format[lastEnd:])

	return result.String(), nil
}

// processEscapes handles escape sequences like \n, \t, etc.
func processEscapes(s string) string {
	replacements := map[string]string{
		`\\`: "\x00BACKSLASH\x00", // Temporary placeholder
		`\n`: "\n",
		`\t`: "\t",
		`\r`: "\r",
		`\a`: "\a",
		`\b`: "\b",
		`\f`: "\f",
		`\v`: "\v",
		`\0`: "\x00",
	}

	result := s
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}

	// Handle octal escapes \NNN
	octalRe := regexp.MustCompile(`\\([0-7]{1,3})`)
	result = octalRe.ReplaceAllStringFunc(result, func(match string) string {
		octal := match[1:]

		val, err := strconv.ParseInt(octal, 8, 32)
		if err != nil {
			return match
		}

		return string(rune(val))
	})

	// Handle hex escapes \xHH
	hexRe := regexp.MustCompile(`\\x([0-9a-fA-F]{1,2})`)
	result = hexRe.ReplaceAllStringFunc(result, func(match string) string {
		hex := match[2:]

		val, err := strconv.ParseInt(hex, 16, 32)
		if err != nil {
			return match
		}

		return string(rune(val))
	})

	// Restore backslashes
	result = strings.ReplaceAll(result, "\x00BACKSLASH\x00", "\\")

	return result
}

// formatArg formats a single argument based on the format specifier
func formatArg(spec, arg string) string {
	// Extract the conversion character
	conv := spec[len(spec)-1]

	// Build the Go format string
	goSpec := "%" + spec[1:len(spec)-1]

	switch conv {
	case 'd', 'i':
		// Decimal integer
		val, err := strconv.ParseInt(arg, 0, 64)
		if err != nil {
			val = 0
		}

		return fmt.Sprintf(goSpec+"d", val)

	case 'o':
		// Octal
		val, err := strconv.ParseInt(arg, 0, 64)
		if err != nil {
			val = 0
		}

		return fmt.Sprintf(goSpec+"o", val)

	case 'u':
		// Unsigned decimal
		val, err := strconv.ParseUint(arg, 0, 64)
		if err != nil {
			val = 0
		}

		return fmt.Sprintf(goSpec+"d", val)

	case 'x':
		// Lowercase hex
		val, err := strconv.ParseInt(arg, 0, 64)
		if err != nil {
			val = 0
		}

		return fmt.Sprintf(goSpec+"x", val)

	case 'X':
		// Uppercase hex
		val, err := strconv.ParseInt(arg, 0, 64)
		if err != nil {
			val = 0
		}

		return fmt.Sprintf(goSpec+"X", val)

	case 'e', 'E':
		// Scientific notation
		val, err := strconv.ParseFloat(arg, 64)
		if err != nil {
			val = 0
		}

		return fmt.Sprintf(goSpec+string(conv), val)

	case 'f', 'F':
		// Floating point
		val, err := strconv.ParseFloat(arg, 64)
		if err != nil {
			val = 0
		}

		return fmt.Sprintf(goSpec+"f", val)

	case 'g', 'G':
		// Compact float
		val, err := strconv.ParseFloat(arg, 64)
		if err != nil {
			val = 0
		}

		return fmt.Sprintf(goSpec+string(conv), val)

	case 's':
		// String
		return fmt.Sprintf(goSpec+"s", arg)

	case 'q':
		// Quoted string
		return fmt.Sprintf(goSpec+"q", arg)

	case 'c':
		// Character
		if len(arg) > 0 {
			return string(arg[0])
		}

		return ""

	case 'b':
		// Binary
		val, err := strconv.ParseInt(arg, 0, 64)
		if err != nil {
			val = 0
		}

		return fmt.Sprintf(goSpec+"b", val)

	default:
		return spec
	}
}
