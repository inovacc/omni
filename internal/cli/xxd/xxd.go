// Package xxd provides hex dump functionality similar to the xxd command.
package xxd

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
)

// Options configures the xxd command behavior
type Options struct {
	Columns   int  // -c: bytes per line (default 16)
	Groups    int  // -g: bytes per group (default 2)
	Length    int  // -l: stop after N bytes (0 = no limit)
	Seek      int  // -s: start at offset
	Reverse   bool // -r: reverse operation (hex dump to binary)
	Plain     bool // -p: plain hex dump (no addresses/ASCII)
	Include   bool // -i: C include file style output
	Uppercase bool // -u: use uppercase hex letters
	Bits      bool // -b: binary digit dump instead of hex
}

// DefaultOptions returns the default options
func DefaultOptions() Options {
	return Options{
		Columns: 16,
		Groups:  2,
	}
}

// Run executes the xxd command
func Run(w io.Writer, r io.Reader, args []string, opts Options) error {
	// Determine input source
	var (
		input    io.Reader
		filename string
	)

	if len(args) > 0 && args[0] != "-" {
		f, err := os.Open(args[0])
		if err != nil {
			return fmt.Errorf("xxd: %w", err)
		}

		defer func() { _ = f.Close() }()

		input = f
		filename = args[0]
	} else {
		input = r
		filename = "stdin"
	}

	// Handle reverse mode
	if opts.Reverse {
		return runReverse(w, input, opts)
	}

	// Handle include mode
	if opts.Include {
		return runInclude(w, input, filename, opts)
	}

	// Handle plain hex mode
	if opts.Plain {
		return runPlain(w, input, opts)
	}

	// Handle binary mode
	if opts.Bits {
		return runBits(w, input, opts)
	}

	// Default hex dump mode
	return runDump(w, input, opts)
}

// runDump produces the standard xxd hex dump output
func runDump(w io.Writer, r io.Reader, opts Options) error {
	if opts.Columns <= 0 {
		opts.Columns = 16
	}

	if opts.Groups <= 0 {
		opts.Groups = 2
	}

	// Handle seek
	if opts.Seek > 0 {
		if seeker, ok := r.(io.Seeker); ok {
			if _, err := seeker.Seek(int64(opts.Seek), io.SeekStart); err != nil {
				return fmt.Errorf("xxd: seek failed: %w", err)
			}
		} else {
			// If can't seek, read and discard
			if _, err := io.CopyN(io.Discard, r, int64(opts.Seek)); err != nil {
				return fmt.Errorf("xxd: seek failed: %w", err)
			}
		}
	}

	buf := make([]byte, opts.Columns)
	offset := opts.Seek
	totalRead := 0

	for {
		// Check length limit
		toRead := opts.Columns
		if opts.Length > 0 && totalRead+toRead > opts.Length {
			toRead = opts.Length - totalRead
		}

		if toRead <= 0 {
			break
		}

		n, err := io.ReadFull(r, buf[:toRead])
		if n == 0 {
			if err == io.EOF {
				break
			}

			if err != nil {
				return fmt.Errorf("xxd: read error: %w", err)
			}
		}

		// Format offset
		_, _ = fmt.Fprintf(w, "%08x: ", offset)

		// Format hex bytes with grouping
		hexPart := formatHexGroups(buf[:n], opts.Columns, opts.Groups, opts.Uppercase)
		_, _ = fmt.Fprint(w, hexPart)

		// Format ASCII representation
		_, _ = fmt.Fprint(w, "  ")

		for i := range n {
			b := buf[i]
			if b >= 32 && b < 127 {
				_, _ = fmt.Fprintf(w, "%c", b)
			} else {
				_, _ = fmt.Fprint(w, ".")
			}
		}

		_, _ = fmt.Fprintln(w)

		offset += n
		totalRead += n

		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}
	}

	return nil
}

// formatHexGroups formats bytes as hex with grouping and padding
func formatHexGroups(data []byte, columns, groupSize int, uppercase bool) string {
	var sb strings.Builder

	for i := range columns {
		if i > 0 && groupSize > 0 && i%groupSize == 0 {
			sb.WriteByte(' ')
		}

		if i < len(data) {
			if uppercase {
				_, _ = fmt.Fprintf(&sb, "%02X", data[i])
			} else {
				_, _ = fmt.Fprintf(&sb, "%02x", data[i])
			}
		} else {
			sb.WriteString("  ")
		}
	}

	return sb.String()
}

// runPlain produces plain hexadecimal output
func runPlain(w io.Writer, r io.Reader, opts Options) error {
	// Handle seek
	if opts.Seek > 0 {
		if seeker, ok := r.(io.Seeker); ok {
			if _, err := seeker.Seek(int64(opts.Seek), io.SeekStart); err != nil {
				return fmt.Errorf("xxd: seek failed: %w", err)
			}
		} else {
			if _, err := io.CopyN(io.Discard, r, int64(opts.Seek)); err != nil {
				return fmt.Errorf("xxd: seek failed: %w", err)
			}
		}
	}

	buf := make([]byte, 4096)
	totalRead := 0
	lineLen := 0
	maxLineLen := 60 // Characters per line in plain mode

	for {
		toRead := len(buf)
		if opts.Length > 0 && totalRead+toRead > opts.Length {
			toRead = opts.Length - totalRead
		}

		if toRead <= 0 {
			break
		}

		n, err := r.Read(buf[:toRead])
		if n > 0 {
			for i := range n {
				if opts.Uppercase {
					_, _ = fmt.Fprintf(w, "%02X", buf[i])
				} else {
					_, _ = fmt.Fprintf(w, "%02x", buf[i])
				}

				lineLen += 2

				if lineLen >= maxLineLen {
					_, _ = fmt.Fprintln(w)
					lineLen = 0
				}
			}

			totalRead += n
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("xxd: read error: %w", err)
		}
	}

	if lineLen > 0 {
		_, _ = fmt.Fprintln(w)
	}

	return nil
}

// runBits produces binary digit dump output
func runBits(w io.Writer, r io.Reader, opts Options) error {
	if opts.Columns <= 0 {
		opts.Columns = 6 // Default 6 bytes per line in binary mode
	}

	// Handle seek
	if opts.Seek > 0 {
		if seeker, ok := r.(io.Seeker); ok {
			if _, err := seeker.Seek(int64(opts.Seek), io.SeekStart); err != nil {
				return fmt.Errorf("xxd: seek failed: %w", err)
			}
		} else {
			if _, err := io.CopyN(io.Discard, r, int64(opts.Seek)); err != nil {
				return fmt.Errorf("xxd: seek failed: %w", err)
			}
		}
	}

	buf := make([]byte, opts.Columns)
	offset := opts.Seek
	totalRead := 0

	for {
		toRead := opts.Columns
		if opts.Length > 0 && totalRead+toRead > opts.Length {
			toRead = opts.Length - totalRead
		}

		if toRead <= 0 {
			break
		}

		n, err := io.ReadFull(r, buf[:toRead])
		if n == 0 {
			if err == io.EOF {
				break
			}

			if err != nil {
				return fmt.Errorf("xxd: read error: %w", err)
			}
		}

		// Format offset
		_, _ = fmt.Fprintf(w, "%08x:", offset)

		// Format binary bytes
		for i := range n {
			_, _ = fmt.Fprintf(w, " %08b", buf[i])
		}

		// Pad if needed
		for i := n; i < opts.Columns; i++ {
			_, _ = fmt.Fprint(w, "         ")
		}

		// Format ASCII representation
		_, _ = fmt.Fprint(w, "  ")

		for i := range n {
			b := buf[i]
			if b >= 32 && b < 127 {
				_, _ = fmt.Fprintf(w, "%c", b)
			} else {
				_, _ = fmt.Fprint(w, ".")
			}
		}

		_, _ = fmt.Fprintln(w)

		offset += n
		totalRead += n

		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}
	}

	return nil
}

// runInclude produces C include file style output
func runInclude(w io.Writer, r io.Reader, filename string, opts Options) error {
	// Sanitize filename for C variable name
	varName := sanitizeVarName(filename)

	// Handle seek
	if opts.Seek > 0 {
		if seeker, ok := r.(io.Seeker); ok {
			if _, err := seeker.Seek(int64(opts.Seek), io.SeekStart); err != nil {
				return fmt.Errorf("xxd: seek failed: %w", err)
			}
		} else {
			if _, err := io.CopyN(io.Discard, r, int64(opts.Seek)); err != nil {
				return fmt.Errorf("xxd: seek failed: %w", err)
			}
		}
	}

	_, _ = fmt.Fprintf(w, "unsigned char %s[] = {\n", varName)

	buf := make([]byte, 4096)
	totalRead := 0
	lineBytes := 0
	first := true

	for {
		toRead := len(buf)
		if opts.Length > 0 && totalRead+toRead > opts.Length {
			toRead = opts.Length - totalRead
		}

		if toRead <= 0 {
			break
		}

		n, err := r.Read(buf[:toRead])
		if n > 0 {
			for i := range n {
				if !first {
					_, _ = fmt.Fprint(w, ",")
				}

				if lineBytes == 0 {
					_, _ = fmt.Fprint(w, "\n  ")
				}

				if opts.Uppercase {
					_, _ = fmt.Fprintf(w, " 0x%02X", buf[i])
				} else {
					_, _ = fmt.Fprintf(w, " 0x%02x", buf[i])
				}

				first = false
				lineBytes++

				if lineBytes >= 12 {
					lineBytes = 0
				}
			}

			totalRead += n
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("xxd: read error: %w", err)
		}
	}

	_, _ = fmt.Fprintln(w, "\n};")
	_, _ = fmt.Fprintf(w, "unsigned int %s_len = %d;\n", varName, totalRead)

	return nil
}

// sanitizeVarName converts a filename to a valid C variable name
func sanitizeVarName(filename string) string {
	var sb strings.Builder

	// Remove directory path
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '/' || filename[i] == '\\' {
			filename = filename[i+1:]

			break
		}
	}

	for _, r := range filename {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			sb.WriteRune(r)
		} else {
			sb.WriteRune('_')
		}
	}

	result := sb.String()
	if len(result) == 0 || unicode.IsDigit(rune(result[0])) {
		result = "_" + result
	}

	return result
}

// runReverse converts hex dump back to binary
func runReverse(w io.Writer, r io.Reader, opts Options) error {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()

		if opts.Plain {
			// Plain mode: just hex characters
			cleaned := strings.Map(func(r rune) rune {
				if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') {
					return r
				}

				return -1
			}, line)

			data, err := hex.DecodeString(cleaned)
			if err != nil {
				return fmt.Errorf("xxd: invalid hex: %w", err)
			}

			if _, err := w.Write(data); err != nil {
				return fmt.Errorf("xxd: write error: %w", err)
			}
		} else {
			// Standard xxd format: skip address, parse hex part
			data, err := parseXxdLine(line)
			if err != nil {
				// Skip invalid lines (might be comments or empty)
				continue
			}

			if _, err := w.Write(data); err != nil {
				return fmt.Errorf("xxd: write error: %w", err)
			}
		}
	}

	return scanner.Err()
}

// parseXxdLine parses a standard xxd output line
func parseXxdLine(line string) ([]byte, error) {
	// Skip empty lines
	line = strings.TrimSpace(line)
	if len(line) == 0 {
		return nil, fmt.Errorf("empty line")
	}

	// Find the colon (end of address) using strings.Cut
	_, remaining, found := strings.Cut(line, ":")
	if !found {
		return nil, fmt.Errorf("no colon found")
	}

	// Find where ASCII part starts (two or more spaces followed by printable chars)
	hexPart := remaining

	for i := 0; i < len(remaining)-1; i++ {
		if remaining[i] == ' ' && remaining[i+1] == ' ' {
			// Check if this is the start of ASCII section
			rest := strings.TrimLeft(remaining[i:], " ")
			if len(rest) > 0 && !strings.ContainsAny(rest[:1], "0123456789abcdefABCDEF") {
				hexPart = remaining[:i]

				break
			}
		}
	}

	// Extract hex characters
	cleaned := strings.Map(func(r rune) rune {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') {
			return r
		}

		return -1
	}, hexPart)

	if len(cleaned) == 0 {
		return nil, fmt.Errorf("no hex data")
	}

	return hex.DecodeString(cleaned)
}

// Dump is a convenience function to dump bytes to a writer in xxd format
func Dump(w io.Writer, _ []byte) error {
	return Run(w, nil, nil, DefaultOptions())
}

// DumpString returns the xxd dump as a string
func DumpString(data []byte) (string, error) {
	var sb strings.Builder

	r := strings.NewReader(string(data))
	if err := Run(&sb, r, nil, DefaultOptions()); err != nil {
		return "", err
	}

	return sb.String(), nil
}
