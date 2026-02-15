//nolint:maintidx // Complex find command requires multiple filter conditions
package find

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/internal/cli/output"
)

// FindOptions configures the find command behavior
type FindOptions struct {
	Name         string        // -name pattern (shell glob)
	IName        string        // -iname pattern (case insensitive)
	Path         string        // -path pattern (match full path)
	IPath        string        // -ipath pattern (case insensitive path)
	Regex        string        // -regex pattern (full path regex)
	IRegex       string        // -iregex pattern (case insensitive regex)
	Type         string        // -type (f=file, d=dir, l=symlink)
	Size         string        // -size [+-]N[ckMG]
	MinDepth     int           // -mindepth N
	MaxDepth     int           // -maxdepth N (0 = unlimited)
	MTime        string        // -mtime [+-]N (days)
	MMin         string        // -mmin [+-]N (minutes)
	ATime        string        // -atime [+-]N (days)
	AMin         string        // -amin [+-]N (minutes)
	Empty        bool          // -empty (empty files/dirs)
	Executable   bool          // -executable
	Readable     bool          // -readable
	Writable     bool          // -writable
	Print0       bool          // -print0 (null separator)
	OutputFormat output.Format // output format
	// Logical operators
	Not bool // -not (negate next condition)
}

// FindResult represents a found file for JSON output
type FindResult struct {
	Path    string    `json:"path"`
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	Mode    string    `json:"mode"`
	ModTime time.Time `json:"modTime"`
	IsDir   bool      `json:"isDir"`
}

// RunFind searches for files matching criteria
func RunFind(w io.Writer, paths []string, opts FindOptions) error {
	f := output.New(w, opts.OutputFormat)
	jsonMode := f.IsJSON()

	if len(paths) == 0 {
		paths = []string{"."}
	}

	// Set default maxdepth
	if opts.MaxDepth == 0 {
		opts.MaxDepth = -1 // unlimited
	}

	// Compile patterns
	var namePattern, pathPattern *regexp.Regexp

	var err error

	if opts.Name != "" {
		namePattern, err = globToRegex(opts.Name, false)
		if err != nil {
			return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("find: invalid name pattern: %s", err))
		}
	}

	if opts.IName != "" {
		namePattern, err = globToRegex(opts.IName, true)
		if err != nil {
			return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("find: invalid iname pattern: %s", err))
		}
	}

	if opts.Path != "" {
		pathPattern, err = globToRegex(opts.Path, false)
		if err != nil {
			return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("find: invalid path pattern: %s", err))
		}
	}

	if opts.IPath != "" {
		pathPattern, err = globToRegex(opts.IPath, true)
		if err != nil {
			return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("find: invalid ipath pattern: %s", err))
		}
	}

	if opts.Regex != "" {
		pathPattern, err = regexp.Compile(opts.Regex)
		if err != nil {
			return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("find: invalid regex: %s", err))
		}
	}

	if opts.IRegex != "" {
		pathPattern, err = regexp.Compile("(?i)" + opts.IRegex)
		if err != nil {
			return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("find: invalid iregex: %s", err))
		}
	}

	// Parse size filter
	var sizeFilter func(int64) bool

	if opts.Size != "" {
		sizeFilter, err = parseSizeFilter(opts.Size)
		if err != nil {
			return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("find: invalid size: %s", err))
		}
	}

	// Parse time filters
	var mtimeFilter, atimeFilter func(time.Time) bool

	if opts.MTime != "" {
		mtimeFilter, err = parseTimeFilter(opts.MTime, 24*time.Hour)
		if err != nil {
			return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("find: invalid mtime: %s", err))
		}
	}

	if opts.MMin != "" {
		mtimeFilter, err = parseTimeFilter(opts.MMin, time.Minute)
		if err != nil {
			return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("find: invalid mmin: %s", err))
		}
	}

	if opts.ATime != "" {
		atimeFilter, err = parseTimeFilter(opts.ATime, 24*time.Hour)
		if err != nil {
			return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("find: invalid atime: %s", err))
		}
	}

	if opts.AMin != "" {
		atimeFilter, err = parseTimeFilter(opts.AMin, time.Minute)
		if err != nil {
			return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("find: invalid amin: %s", err))
		}
	}

	separator := "\n"
	if opts.Print0 {
		separator = "\x00"
	}

	var results []FindResult

	for _, root := range paths {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "find: %q: %v\n", path, err)
				return nil
			}

			// Calculate depth
			depth := 0

			if path != root {
				relPath, _ := filepath.Rel(root, path)
				depth = strings.Count(relPath, string(filepath.Separator)) + 1
			}

			// Check mindepth
			if depth < opts.MinDepth {
				return nil
			}

			// Check maxdepth
			if opts.MaxDepth >= 0 && depth > opts.MaxDepth {
				if d.IsDir() {
					return fs.SkipDir
				}

				return nil
			}

			// Apply filters
			match := true

			// Type filter
			if opts.Type != "" && match {
				match = matchType(d, opts.Type)
			}

			// Name filter
			if namePattern != nil && match {
				match = namePattern.MatchString(d.Name())
			}

			// Path filter
			if pathPattern != nil && match {
				match = pathPattern.MatchString(path)
			}

			// Size filter
			if sizeFilter != nil && match {
				if info, err := d.Info(); err == nil {
					match = sizeFilter(info.Size())
				} else {
					match = false
				}
			}

			// Time filters
			if mtimeFilter != nil && match {
				if info, err := d.Info(); err == nil {
					match = mtimeFilter(info.ModTime())
				} else {
					match = false
				}
			}

			if atimeFilter != nil && match {
				if info, err := d.Info(); err == nil {
					// Note: Go doesn't expose atime directly, using mtime as fallback
					match = atimeFilter(info.ModTime())
				} else {
					match = false
				}
			}

			// Empty filter
			if opts.Empty && match {
				match = isEmpty(path, d)
			}

			// Permission filters
			if opts.Readable && match {
				match = isReadable(path)
			}

			if opts.Writable && match {
				match = isWritable(path)
			}

			if opts.Executable && match {
				match = isExecutable(d)
			}

			// Apply NOT
			if opts.Not {
				match = !match
			}

			if match {
				if jsonMode {
					info, _ := d.Info()

					result := FindResult{
						Path:  path,
						Name:  d.Name(),
						IsDir: d.IsDir(),
					}
					if info != nil {
						result.Size = info.Size()
						result.Mode = info.Mode().String()
						result.ModTime = info.ModTime()
					}

					results = append(results, result)
				} else {
					_, _ = fmt.Fprint(w, path, separator)
				}
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("find: error walking %q: %w", root, err)
		}
	}

	if jsonMode {
		return f.Print(results)
	}

	return nil
}

func globToRegex(pattern string, caseInsensitive bool) (*regexp.Regexp, error) {
	var sb strings.Builder

	if caseInsensitive {
		sb.WriteString("(?i)")
	}

	sb.WriteString("^")

	for i := 0; i < len(pattern); i++ {
		c := pattern[i]

		switch c {
		case '*':
			sb.WriteString(".*")
		case '?':
			sb.WriteString(".")
		case '[':
			// Find closing bracket
			j := i + 1
			if j < len(pattern) && pattern[j] == '!' {
				j++
			}

			if j < len(pattern) && pattern[j] == ']' {
				j++
			}

			for j < len(pattern) && pattern[j] != ']' {
				j++
			}

			if j >= len(pattern) {
				sb.WriteString("\\[")
			} else {
				bracket := pattern[i : j+1]
				if len(bracket) > 2 && bracket[1] == '!' {
					sb.WriteString("[^")
					sb.WriteString(bracket[2:])
				} else {
					sb.WriteString(bracket)
				}

				i = j
			}
		case '.', '+', '^', '$', '(', ')', '{', '}', '|', '\\':
			sb.WriteString("\\")
			sb.WriteByte(c)
		default:
			sb.WriteByte(c)
		}
	}

	sb.WriteString("$")

	return regexp.Compile(sb.String())
}

func matchType(d fs.DirEntry, typeFlag string) bool {
	switch typeFlag {
	case "f":
		return d.Type().IsRegular()
	case "d":
		return d.IsDir()
	case "l":
		return d.Type()&fs.ModeSymlink != 0
	case "p":
		return d.Type()&fs.ModeNamedPipe != 0
	case "s":
		return d.Type()&fs.ModeSocket != 0
	case "b":
		return d.Type()&fs.ModeDevice != 0
	case "c":
		return d.Type()&fs.ModeCharDevice != 0
	default:
		return true
	}
}

func parseSizeFilter(sizeStr string) (func(int64) bool, error) {
	if sizeStr == "" {
		return func(_ int64) bool { return true }, nil
	}

	var op byte = '='

	s := sizeStr
	if s[0] == '+' || s[0] == '-' {
		op = s[0]
		s = s[1:]
	}

	// Parse number and suffix
	var numStr string

	var suffix byte = 'b' // default: 512-byte blocks

	for i, c := range s {
		if c >= '0' && c <= '9' {
			continue
		}

		numStr = s[:i]

		if i < len(s) {
			suffix = s[i]
		}

		break
	}

	if numStr == "" {
		numStr = strings.TrimRight(s, "bckMGTP")
		if len(numStr) < len(s) {
			suffix = s[len(numStr)]
		}
	}

	n, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid number: %s", numStr)
	}

	// Get unit size for rounding (Linux find rounds up to units)
	var unitSize int64

	switch suffix {
	case 'b':
		unitSize = 512
	case 'c':
		unitSize = 1 // bytes - no rounding
	case 'w':
		unitSize = 2
	case 'k':
		unitSize = 1024
	case 'M':
		unitSize = 1024 * 1024
	case 'G':
		unitSize = 1024 * 1024 * 1024
	case 'T':
		unitSize = 1024 * 1024 * 1024 * 1024
	case 'P':
		unitSize = 1024 * 1024 * 1024 * 1024 * 1024
	default:
		unitSize = 512 // default to blocks
	}

	// Linux find rounds file size up to units: ceil(fileSize / unitSize)
	// Then compares the number of units
	roundUp := func(fileSize int64) int64 {
		if unitSize == 1 {
			return fileSize
		}

		return (fileSize + unitSize - 1) / unitSize
	}

	switch op {
	case '+':
		// File uses more than n units
		return func(fileSize int64) bool { return roundUp(fileSize) > n }, nil
	case '-':
		// File uses less than n units
		return func(fileSize int64) bool { return roundUp(fileSize) < n }, nil
	default:
		// File uses exactly n units
		return func(fileSize int64) bool {
			return roundUp(fileSize) == n
		}, nil
	}
}

func parseTimeFilter(timeStr string, unit time.Duration) (func(time.Time) bool, error) {
	if timeStr == "" {
		return func(_ time.Time) bool { return true }, nil
	}

	var op byte = '='

	s := timeStr
	if s[0] == '+' || s[0] == '-' {
		op = s[0]
		s = s[1:]
	}

	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid time: %s", timeStr)
	}

	now := time.Now()
	threshold := now.Add(-time.Duration(n) * unit)
	nextThreshold := now.Add(-time.Duration(n+1) * unit)

	switch op {
	case '+':
		// More than N units ago
		return func(t time.Time) bool { return t.Before(nextThreshold) }, nil
	case '-':
		// Less than N units ago
		return func(t time.Time) bool { return t.After(threshold) }, nil
	default:
		// Exactly N units ago (within the unit)
		return func(t time.Time) bool {
			return t.After(nextThreshold) && t.Before(threshold)
		}, nil
	}
}

func isEmpty(path string, d fs.DirEntry) bool {
	if d.IsDir() {
		entries, err := os.ReadDir(path)
		return err == nil && len(entries) == 0
	}

	if info, err := d.Info(); err == nil {
		return info.Size() == 0
	}

	return false
}

func isReadable(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}

	_ = f.Close()

	return true
}

func isWritable(path string) bool {
	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return false
	}

	_ = f.Close()

	return true
}

func isExecutable(d fs.DirEntry) bool {
	if d.IsDir() {
		return true // directories are "executable" if traversable
	}

	info, err := d.Info()
	if err != nil {
		return false
	}

	return info.Mode()&0111 != 0
}
