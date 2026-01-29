package shuf

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

// ShufOptions configures the shuf command behavior
type ShufOptions struct {
	Echo       bool   // -e: treat args as input lines
	InputRange string // -i LO-HI: treat each number LO through HI as input line
	HeadCount  int    // -n: output at most COUNT lines
	Repeat     bool   // -r: output lines can be repeated
	ZeroTerm   bool   // -z: line delimiter is NUL
	JSON       bool   // --json: output as JSON
}

// ShufResult represents shuf output for JSON
type ShufResult struct {
	Lines []string `json:"lines"`
	Count int      `json:"count"`
}

// RunShuf shuffles input lines randomly
func RunShuf(w io.Writer, args []string, opts ShufOptions) error {
	var lines []string

	if opts.InputRange != "" {
		// Generate range
		parts := strings.Split(opts.InputRange, "-")
		if len(parts) != 2 {
			return fmt.Errorf("shuf: invalid input range %q", opts.InputRange)
		}

		lo, err := strconv.Atoi(parts[0])
		if err != nil {
			return fmt.Errorf("shuf: invalid input range %q", opts.InputRange)
		}

		hi, err := strconv.Atoi(parts[1])
		if err != nil {
			return fmt.Errorf("shuf: invalid input range %q", opts.InputRange)
		}

		if lo > hi {
			return fmt.Errorf("shuf: invalid input range %q", opts.InputRange)
		}

		for i := lo; i <= hi; i++ {
			lines = append(lines, strconv.Itoa(i))
		}
	} else if opts.Echo {
		// Use args as input lines
		lines = args
	} else {
		// Read from file or stdin
		var reader io.Reader

		if len(args) == 0 || args[0] == "-" {
			reader = os.Stdin
		} else {
			f, err := os.Open(args[0])
			if err != nil {
				return fmt.Errorf("shuf: %w", err)
			}

			defer func() { _ = f.Close() }()

			reader = f
		}

		delimiter := byte('\n')
		if opts.ZeroTerm {
			delimiter = 0
		}

		scanner := bufio.NewScanner(reader)
		scanner.Split(splitFunc(delimiter))

		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("shuf: %w", err)
		}
	}

	// Shuffle using Fisher-Yates
	rand.Shuffle(len(lines), func(i, j int) {
		lines[i], lines[j] = lines[j], lines[i]
	})

	outputDelim := "\n"
	if opts.ZeroTerm {
		outputDelim = "\x00"
	}

	count := len(lines)
	if opts.HeadCount > 0 && opts.HeadCount < count {
		count = opts.HeadCount
	}

	if opts.JSON {
		var output []string
		if opts.Repeat {
			if opts.HeadCount == 0 {
				return fmt.Errorf("shuf: --repeat requires --head-count")
			}
			for i := 0; i < opts.HeadCount; i++ {
				idx := rand.Intn(len(lines))
				output = append(output, lines[idx])
			}
		} else {
			output = lines[:count]
		}
		return json.NewEncoder(w).Encode(ShufResult{Lines: output, Count: len(output)})
	}

	if opts.Repeat {
		// Output with repetition allowed
		if opts.HeadCount == 0 {
			return fmt.Errorf("shuf: --repeat requires --head-count")
		}

		for i := 0; i < opts.HeadCount; i++ {
			idx := rand.Intn(len(lines))
			_, _ = fmt.Fprint(w, lines[idx], outputDelim)
		}
	} else {
		for i := 0; i < count; i++ {
			_, _ = fmt.Fprint(w, lines[i], outputDelim)
		}
	}

	return nil
}

func splitFunc(delim byte) bufio.SplitFunc {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		if i := indexOf(data, delim); i >= 0 {
			return i + 1, data[0:i], nil
		}

		if atEOF {
			return len(data), data, nil
		}

		return 0, nil, nil
	}
}

func indexOf(data []byte, b byte) int {
	for i, c := range data {
		if c == b {
			return i
		}
	}

	return -1
}
