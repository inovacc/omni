package cmp

import (
	"fmt"
	"io"
	"os"

	"github.com/inovacc/omni/internal/cli/output"
)

// CmpOptions configures the cmp command behavior
type CmpOptions struct {
	Silent       bool          // -s: suppress all output
	Verbose      bool          // -l: output byte numbers and values
	PrintBytes   bool          // -b: print differing bytes
	SkipBytes1   int64         // -i SKIP1: skip first SKIP1 bytes of FILE1
	SkipBytes2   int64         // -i SKIP2: skip first SKIP2 bytes of FILE2
	MaxBytes     int64         // -n LIMIT: compare at most LIMIT bytes
	OutputFormat output.Format // output format
}

// CmpJSONResult represents the JSON output for cmp
type CmpJSONResult struct {
	File1     string `json:"file1"`
	File2     string `json:"file2"`
	Identical bool   `json:"identical"`
	DiffByte  int64  `json:"diffByte,omitempty"`
	DiffLine  int64  `json:"diffLine,omitempty"`
	Byte1     byte   `json:"byte1,omitempty"`
	Byte2     byte   `json:"byte2,omitempty"`
	EOF       string `json:"eof,omitempty"`
}

// CmpResult represents the result of comparison
type CmpResult int

const (
	CmpEqual  CmpResult = 0
	CmpDiffer CmpResult = 1
	CmpError  CmpResult = 2
)

// RunCmp compares two files byte by byte
func RunCmp(w io.Writer, args []string, opts CmpOptions) (CmpResult, error) {
	if len(args) < 2 {
		return CmpError, fmt.Errorf("cmp: missing operand")
	}

	f := output.New(w, opts.OutputFormat)
	jsonMode := f.IsJSON()

	file1, file2 := args[0], args[1]

	// Open files
	var r1, r2 io.Reader

	if file1 == "-" {
		r1 = os.Stdin
	} else {
		f, err := os.Open(file1)
		if err != nil {
			return CmpError, fmt.Errorf("cmp: %s: %w", file1, err)
		}

		defer func() { _ = f.Close() }()

		r1 = f

		// Skip bytes if requested
		if opts.SkipBytes1 > 0 {
			if _, err := f.Seek(opts.SkipBytes1, io.SeekStart); err != nil {
				return CmpError, fmt.Errorf("cmp: %s: %w", file1, err)
			}
		}
	}

	if file2 == "-" {
		if file1 == "-" {
			return CmpError, fmt.Errorf("cmp: both files cannot be stdin")
		}

		r2 = os.Stdin
	} else {
		f, err := os.Open(file2)
		if err != nil {
			return CmpError, fmt.Errorf("cmp: %s: %w", file2, err)
		}

		defer func() { _ = f.Close() }()

		r2 = f

		// Skip bytes if requested
		if opts.SkipBytes2 > 0 {
			if _, err := f.Seek(opts.SkipBytes2, io.SeekStart); err != nil {
				return CmpError, fmt.Errorf("cmp: %s: %w", file2, err)
			}
		}
	}

	// Compare byte by byte
	buf1 := make([]byte, 4096)
	buf2 := make([]byte, 4096)
	byteNum := int64(1)
	lineNum := int64(1)
	totalRead := int64(0)

	for opts.MaxBytes <= 0 || totalRead < opts.MaxBytes {
		toRead := len(buf1)
		if opts.MaxBytes > 0 && totalRead+int64(toRead) > opts.MaxBytes {
			toRead = int(opts.MaxBytes - totalRead)
		}

		n1, err1 := r1.Read(buf1[:toRead])
		n2, err2 := r2.Read(buf2[:toRead])

		// Compare bytes
		minN := min(n2, n1)

		for i := range minN {
			if buf1[i] != buf2[i] {
				if jsonMode {
					result := CmpJSONResult{
						File1:     file1,
						File2:     file2,
						Identical: false,
						DiffByte:  byteNum,
						DiffLine:  lineNum,
						Byte1:     buf1[i],
						Byte2:     buf2[i],
					}

					if err := f.Print(result); err != nil {
						return CmpError, fmt.Errorf("cmp: json encode: %w", err)
					}

					return CmpDiffer, nil
				}

				if !opts.Silent {
					if opts.Verbose {
						_, _ = fmt.Fprintf(w, "%d %o %o\n", byteNum, buf1[i], buf2[i])
					} else if opts.PrintBytes {
						_, _ = fmt.Fprintf(w, "%s %s differ: byte %d, line %d is %3o %c %3o %c\n",
							file1, file2, byteNum, lineNum, buf1[i], printableChar(buf1[i]), buf2[i], printableChar(buf2[i]))

						return CmpDiffer, nil
					} else {
						_, _ = fmt.Fprintf(w, "%s %s differ: byte %d, line %d\n", file1, file2, byteNum, lineNum)

						return CmpDiffer, nil
					}
				} else {
					return CmpDiffer, nil
				}
			}

			if buf1[i] == '\n' {
				lineNum++
			}

			byteNum++
		}

		totalRead += int64(minN)

		// Check for EOF differences
		if n1 != n2 {
			if jsonMode {
				eofFile := file1
				if n1 > n2 {
					eofFile = file2
				}

				result := CmpJSONResult{
					File1:     file1,
					File2:     file2,
					Identical: false,
					DiffByte:  byteNum - 1,
					EOF:       eofFile,
				}

				if err := f.Print(result); err != nil {
					return CmpError, fmt.Errorf("cmp: json encode: %w", err)
				}

				return CmpDiffer, nil
			}

			if !opts.Silent {
				if n1 < n2 {
					_, _ = fmt.Fprintf(w, "cmp: EOF on %s after byte %d\n", file1, byteNum-1)
				} else {
					_, _ = fmt.Fprintf(w, "cmp: EOF on %s after byte %d\n", file2, byteNum-1)
				}
			}

			return CmpDiffer, nil
		}

		if err1 == io.EOF && err2 == io.EOF {
			break
		}

		if err1 != nil && err1 != io.EOF {
			return CmpError, fmt.Errorf("cmp: %s: %w", file1, err1)
		}

		if err2 != nil && err2 != io.EOF {
			return CmpError, fmt.Errorf("cmp: %s: %w", file2, err2)
		}
	}

	if jsonMode {
		result := CmpJSONResult{
			File1:     file1,
			File2:     file2,
			Identical: true,
		}

		if err := f.Print(result); err != nil {
			return CmpError, fmt.Errorf("cmp: json encode: %w", err)
		}
	}

	return CmpEqual, nil
}

func printableChar(b byte) byte {
	if b >= 32 && b < 127 {
		return b
	}

	return ' '
}
