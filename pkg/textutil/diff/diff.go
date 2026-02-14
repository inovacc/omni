package diff

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
)

// Line represents a single line in a diff.
type Line struct {
	Type    rune // ' ' for context, '-' for removed, '+' for added
	Content string
}

// Hunk represents a contiguous block of changes.
type Hunk struct {
	Start1 int
	Count1 int
	Start2 int
	Count2 int
	Lines  []Line
}

// Options configures diff computation.
type Options struct {
	Context int // Lines of context (default 3)
}

// Option is a functional option for diff operations.
type Option func(*Options)

// WithContext sets the number of context lines around changes.
func WithContext(n int) Option {
	return func(o *Options) { o.Context = n }
}

// ComputeDiff computes the diff between two slices of strings using an LCS algorithm.
// Returns hunks describing the changes.
func ComputeDiff(lines1, lines2 []string, opts ...Option) []Hunk {
	cfg := Options{Context: 3}
	for _, o := range opts {
		o(&cfg)
	}

	return computeDiff(lines1, lines2, cfg)
}

// FormatUnified formats hunks as a unified diff string.
func FormatUnified(w io.Writer, file1, file2 string, hunks []Hunk) {
	if len(hunks) == 0 {
		return
	}

	_, _ = fmt.Fprintf(w, "--- %s\n", file1)
	_, _ = fmt.Fprintf(w, "+++ %s\n", file2)

	for _, hunk := range hunks {
		_, _ = fmt.Fprintf(w, "@@ -%d,%d +%d,%d @@\n", hunk.Start1, hunk.Count1, hunk.Start2, hunk.Count2)
		for _, line := range hunk.Lines {
			_, _ = fmt.Fprintf(w, "%s%s\n", string(line.Type), line.Content)
		}
	}
}

// CompareJSON compares two parsed JSON values and returns a list of human-readable differences.
func CompareJSON(v1, v2 any) []string {
	return compareJSON("", v1, v2)
}

// CompareJSONBytes parses and compares two JSON byte slices.
func CompareJSONBytes(data1, data2 []byte) ([]string, error) {
	var j1, j2 any
	if err := json.Unmarshal(data1, &j1); err != nil {
		return nil, fmt.Errorf("diff: invalid JSON in first input: %w", err)
	}

	if err := json.Unmarshal(data2, &j2); err != nil {
		return nil, fmt.Errorf("diff: invalid JSON in second input: %w", err)
	}

	if reflect.DeepEqual(j1, j2) {
		return nil, nil
	}

	return compareJSON("", j1, j2), nil
}

func computeDiff(lines1, lines2 []string, opts Options) []Hunk {
	m, n := len(lines1), len(lines2)

	// Build LCS table
	lcs := make([][]int, m+1)
	for i := range lcs {
		lcs[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if lines1[i-1] == lines2[j-1] {
				lcs[i][j] = lcs[i-1][j-1] + 1
			} else {
				lcs[i][j] = max(lcs[i-1][j], lcs[i][j-1])
			}
		}
	}

	// Backtrack to find diff
	var allLines []Line

	i, j := m, n
	for i > 0 || j > 0 {
		if i > 0 && j > 0 && lines1[i-1] == lines2[j-1] {
			allLines = append([]Line{{Type: ' ', Content: lines1[i-1]}}, allLines...)
			i--
			j--
		} else if j > 0 && (i == 0 || lcs[i][j-1] >= lcs[i-1][j]) {
			allLines = append([]Line{{Type: '+', Content: lines2[j-1]}}, allLines...)
			j--
		} else if i > 0 {
			allLines = append([]Line{{Type: '-', Content: lines1[i-1]}}, allLines...)
			i--
		}
	}

	return groupIntoHunks(allLines, opts)
}

func groupIntoHunks(allLines []Line, opts Options) []Hunk {
	context := opts.Context
	if context == 0 {
		context = 3
	}

	var (
		hunks       []Hunk
		currentHunk *Hunk
	)

	contextBuffer := make([]Line, 0, context)
	pos1, pos2 := 0, 0

	for _, line := range allLines {
		isChange := line.Type != ' '

		if isChange {
			if currentHunk == nil {
				start1 := pos1 - len(contextBuffer)
				start2 := pos2 - len(contextBuffer)

				if start1 < 0 {
					start1 = 0
				}

				if start2 < 0 {
					start2 = 0
				}

				currentHunk = &Hunk{
					Start1: start1 + 1,
					Start2: start2 + 1,
					Lines:  append([]Line{}, contextBuffer...),
				}
			}

			currentHunk.Lines = append(currentHunk.Lines, line)
		} else {
			if currentHunk != nil {
				currentHunk.Lines = append(currentHunk.Lines, line)
				trailingContext := 0

				for i := len(currentHunk.Lines) - 1; i >= 0; i-- {
					if currentHunk.Lines[i].Type == ' ' {
						trailingContext++
					} else {
						break
					}
				}

				if trailingContext >= context*2 {
					currentHunk.Lines = currentHunk.Lines[:len(currentHunk.Lines)-context]
					currentHunk.Count1, currentHunk.Count2 = countLines(currentHunk.Lines)
					hunks = append(hunks, *currentHunk)
					currentHunk = nil
					contextBuffer = contextBuffer[:0]
				}
			} else {
				contextBuffer = append(contextBuffer, line)
				if len(contextBuffer) > context {
					contextBuffer = contextBuffer[1:]
				}
			}
		}

		switch line.Type {
		case ' ':
			pos1++
			pos2++
		case '-':
			pos1++
		case '+':
			pos2++
		}
	}

	if currentHunk != nil {
		trailingContext := 0

		for i := len(currentHunk.Lines) - 1; i >= 0; i-- {
			if currentHunk.Lines[i].Type == ' ' {
				trailingContext++
			} else {
				break
			}
		}

		if trailingContext > context {
			currentHunk.Lines = currentHunk.Lines[:len(currentHunk.Lines)-(trailingContext-context)]
		}

		currentHunk.Count1, currentHunk.Count2 = countLines(currentHunk.Lines)
		hunks = append(hunks, *currentHunk)
	}

	return hunks
}

func countLines(lines []Line) (count1, count2 int) {
	for _, line := range lines {
		switch line.Type {
		case ' ':
			count1++
			count2++
		case '-':
			count1++
		case '+':
			count2++
		}
	}

	return
}

func compareJSON(path string, v1, v2 any) []string {
	var diffs []string

	if reflect.TypeOf(v1) != reflect.TypeOf(v2) {
		diffs = append(diffs, fmt.Sprintf("%s: type mismatch: %T vs %T", pathOrRoot(path), v1, v2))
		return diffs
	}

	switch val1 := v1.(type) {
	case map[string]any:
		val2 := v2.(map[string]any)

		allKeys := make(map[string]bool)
		for k := range val1 {
			allKeys[k] = true
		}

		for k := range val2 {
			allKeys[k] = true
		}

		for k := range allKeys {
			p := path + "." + k
			if path == "" {
				p = k
			}

			_, in1 := val1[k]
			_, in2 := val2[k]

			if !in1 {
				diffs = append(diffs, fmt.Sprintf("+ %s: %v", p, val2[k]))
			} else if !in2 {
				diffs = append(diffs, fmt.Sprintf("- %s: %v", p, val1[k]))
			} else {
				diffs = append(diffs, compareJSON(p, val1[k], val2[k])...)
			}
		}
	case []any:
		val2 := v2.([]any)
		maxLen := max(len(val2), len(val1))

		for i := range maxLen {
			p := fmt.Sprintf("%s[%d]", path, i)
			if path == "" {
				p = fmt.Sprintf("[%d]", i)
			}

			if i >= len(val1) {
				diffs = append(diffs, fmt.Sprintf("+ %s: %v", p, val2[i]))
			} else if i >= len(val2) {
				diffs = append(diffs, fmt.Sprintf("- %s: %v", p, val1[i]))
			} else {
				diffs = append(diffs, compareJSON(p, val1[i], val2[i])...)
			}
		}
	default:
		if !reflect.DeepEqual(v1, v2) {
			diffs = append(diffs, fmt.Sprintf("~ %s: %v -> %v", pathOrRoot(path), v1, v2))
		}
	}

	return diffs
}

func pathOrRoot(path string) string {
	if path == "" {
		return "(root)"
	}

	return path
}

// TruncateOrPad truncates a string to width or pads it with spaces.
func TruncateOrPad(s string, width int) string {
	if len(s) > width {
		return s[:width-1] + ">"
	}

	return s + strings.Repeat(" ", width-len(s))
}
