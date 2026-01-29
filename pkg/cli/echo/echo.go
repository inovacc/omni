package echo

import (
	"fmt"
	"io"
	"strings"
)

// EchoOptions holds the options for the echo command.
type EchoOptions struct {
	NoNewline      bool // -n: do not output trailing newline
	EnableEscapes  bool // -e: enable interpretation of backslash escapes
	DisableEscapes bool // -E: disable interpretation of backslash escapes (default)
}

// RunEcho writes the arguments to the writer.
func RunEcho(w io.Writer, args []string, opts EchoOptions) error {
	output := strings.Join(args, " ")

	if opts.EnableEscapes && !opts.DisableEscapes {
		output = interpretEscapes(output)
	}

	if opts.NoNewline {
		_, err := fmt.Fprint(w, output)
		return err
	}

	_, err := fmt.Fprintln(w, output)

	return err
}

// interpretEscapes interprets backslash escape sequences.
func interpretEscapes(s string) string {
	var result strings.Builder
	result.Grow(len(s))

	i := 0
	for i < len(s) {
		if s[i] == '\\' && i+1 < len(s) {
			switch s[i+1] {
			case '\\':
				result.WriteByte('\\')

				i += 2
			case 'a':
				result.WriteByte('\a')

				i += 2
			case 'b':
				result.WriteByte('\b')

				i += 2
			case 'c':
				// \c: produce no further output
				return result.String()
			case 'e':
				result.WriteByte('\x1b')

				i += 2
			case 'f':
				result.WriteByte('\f')

				i += 2
			case 'n':
				result.WriteByte('\n')

				i += 2
			case 'r':
				result.WriteByte('\r')

				i += 2
			case 't':
				result.WriteByte('\t')

				i += 2
			case 'v':
				result.WriteByte('\v')

				i += 2
			case '0':
				// \0nnn: octal value (up to 3 digits)
				val, consumed := parseOctal(s[i+2:], 3)
				result.WriteByte(byte(val))

				i += 2 + consumed
			case 'x':
				// \xHH: hex value (up to 2 digits)
				if i+2 < len(s) {
					val, consumed := parseHex(s[i+2:], 2)
					if consumed > 0 {
						result.WriteByte(byte(val))

						i += 2 + consumed
					} else {
						result.WriteByte(s[i])
						i++
					}
				} else {
					result.WriteByte(s[i])
					i++
				}
			default:
				result.WriteByte(s[i])
				i++
			}
		} else {
			result.WriteByte(s[i])
			i++
		}
	}

	return result.String()
}

// parseOctal parses up to maxDigits octal digits from s.
func parseOctal(s string, maxDigits int) (int, int) {
	val := 0
	consumed := 0

	for i := 0; i < len(s) && i < maxDigits; i++ {
		if s[i] >= '0' && s[i] <= '7' {
			val = val*8 + int(s[i]-'0')
			consumed++
		} else {
			break
		}
	}

	return val, consumed
}

// parseHex parses up to maxDigits hex digits from s.
func parseHex(s string, maxDigits int) (int, int) {
	val := 0
	consumed := 0

	for i := 0; i < len(s) && i < maxDigits; i++ {
		c := s[i]

		var digit int

		switch {
		case c >= '0' && c <= '9':
			digit = int(c - '0')
		case c >= 'a' && c <= 'f':
			digit = int(c-'a') + 10
		case c >= 'A' && c <= 'F':
			digit = int(c-'A') + 10
		default:
			return val, consumed
		}

		val = val*16 + digit
		consumed++
	}

	return val, consumed
}
