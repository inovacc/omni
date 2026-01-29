package tail

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/inovacc/omni/internal/cli/input"
)

// TailOptions configures the tail command behavior
type TailOptions struct {
	Lines   int           // -n: number of lines to print
	Bytes   int           // -c: number of bytes to print
	Follow  bool          // -f: output appended data as file grows
	Quiet   bool          // -q: never print headers
	Verbose bool          // -v: always print headers
	Sleep   time.Duration // --sleep-interval: sleep interval for -f
	JSON    bool          // --json: output as JSON
}

// TailResult represents tail output for JSON
type TailResult struct {
	File  string   `json:"file,omitempty"`
	Lines []string `json:"lines"`
}

// DefaultTailLines is the default number of lines for tail
const DefaultTailLines = 10

// RunTail executes the tail command
// r is the default input reader (used when args is empty or contains "-")
func RunTail(w io.Writer, r io.Reader, args []string, opts TailOptions) error {
	if opts.Lines == 0 && opts.Bytes == 0 {
		opts.Lines = DefaultTailLines
	}

	if opts.Sleep == 0 {
		opts.Sleep = time.Second
	}

	sources, err := input.Open(args, r)
	if err != nil {
		return fmt.Errorf("tail: %w", err)
	}
	defer input.CloseAll(sources)

	showHeaders := len(sources) > 1 || opts.Verbose
	if opts.Quiet {
		showHeaders = false
	}

	var results []TailResult

	for i, src := range sources {
		if showHeaders {
			if i > 0 {
				_, _ = fmt.Fprintln(w)
			}

			_, _ = fmt.Fprintf(w, "==> %s <==\n", src.Name)
		}

		if opts.JSON {
			lines, err := tailLinesJSON(src.Reader, opts.Lines)
			if err != nil {
				return err
			}

			results = append(results, TailResult{File: src.Name, Lines: lines})

			continue
		}

		if opts.Bytes > 0 {
			if err := tailBytes(w, src.Reader, opts.Bytes); err != nil {
				return err
			}
		} else {
			if err := tailLines(w, src.Reader, opts.Lines); err != nil {
				return err
			}
		}

		// Handle -f (follow) mode - only works with files, not stdin
		if opts.Follow {
			if f, ok := src.Reader.(*os.File); ok {
				if err := followFile(w, f, opts.Sleep); err != nil {
					return err
				}
			}
		}
	}

	if opts.JSON {
		return json.NewEncoder(w).Encode(results)
	}

	return nil
}

func tailLinesJSON(r io.Reader, n int) ([]string, error) {
	scanner := bufio.NewScanner(r)
	lines := make([]string, 0, n)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > n {
			lines = lines[1:]
		}
	}

	return lines, scanner.Err()
}

func tailLines(w io.Writer, r io.Reader, n int) error {
	// Read all lines into a circular buffer
	scanner := bufio.NewScanner(r)
	lines := make([]string, 0, n)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > n {
			lines = lines[1:]
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	for _, line := range lines {
		_, _ = fmt.Fprintln(w, line)
	}

	return nil
}

func tailBytes(w io.Writer, r io.Reader, n int) error {
	// For seekable readers, seek to end and read backwards
	if seeker, ok := r.(io.ReadSeeker); ok {
		size, err := seeker.Seek(0, io.SeekEnd)
		if err == nil {
			start := max(size-int64(n), 0)

			_, err = seeker.Seek(start, io.SeekStart)
			if err != nil {
				return err
			}

			_, err = io.Copy(w, r)

			return err
		}
	}

	// For non-seekable readers, read all and keep last n bytes
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	start := max(len(data)-n, 0)

	_, err = w.Write(data[start:])

	return err
}

func followFile(w io.Writer, f *os.File, sleep time.Duration) error {
	reader := bufio.NewReader(f)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				time.Sleep(sleep)
				continue
			}

			return err
		}

		_, _ = fmt.Fprint(w, line)
	}
}

// Tail returns the last n lines from a slice (for compatibility)
func Tail(lines []string, n int) []string {
	if n > len(lines) {
		n = len(lines)
	}

	if n < 0 {
		n = 0
	}

	return lines[len(lines)-n:]
}
