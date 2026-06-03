package dd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

// maxDdBlockSize caps the per-block buffer allocation to guard against
// memory exhaustion from a malformed bs=/ibs=/obs= operand (1 GiB).
const maxDdBlockSize int64 = 1 << 30

// DdOptions configures the dd command behavior
type DdOptions struct {
	InputFile    string // if=FILE
	OutputFile   string // of=FILE
	BlockSize    int64  // bs=BYTES (sets both ibs and obs)
	InputBS      int64  // ibs=BYTES
	OutputBS     int64  // obs=BYTES
	Count        int64  // count=N (0 = unlimited)
	Skip         int64  // skip=N input blocks
	Seek         int64  // seek=N output blocks
	Conv         string // conv=CONVERSION[,CONVERSION]...
	Status       string // status=LEVEL (none, noxfer, progress)
	NoTrunc      bool   // don't truncate output file
	StatusWriter io.Writer
}

// DdStats holds statistics from the dd operation
type DdStats struct {
	BytesRead    int64
	BytesWritten int64
	BlocksIn     int64
	BlocksOut    int64
	PartialIn    int64
	PartialOut   int64
	StartTime    time.Time
	EndTime      time.Time
}

// RunDd copies and converts data
func RunDd(w io.Writer, opts DdOptions) error {
	// Set defaults
	if opts.InputBS == 0 {
		opts.InputBS = 512
	}

	if opts.OutputBS == 0 {
		opts.OutputBS = 512
	}

	if opts.BlockSize > 0 {
		opts.InputBS = opts.BlockSize
		opts.OutputBS = opts.BlockSize
	}

	if opts.StatusWriter == nil {
		opts.StatusWriter = os.Stderr
	}

	// Validate block sizes are strictly positive and within a sane ceiling.
	// A non-positive size would panic make([]byte, ...) / break the reslice
	// loop; an absurdly large size would attempt an unbounded allocation.
	if opts.InputBS <= 0 || opts.OutputBS <= 0 {
		return cmderr.Wrap(cmderr.ErrInvalidInput, "dd: block size must be a positive value")
	}

	if opts.InputBS > maxDdBlockSize || opts.OutputBS > maxDdBlockSize {
		return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("dd: block size exceeds maximum of %d bytes", maxDdBlockSize))
	}

	// Parse conversions
	convs := make(map[string]bool)

	if opts.Conv != "" {
		for c := range strings.SplitSeq(opts.Conv, ",") {
			convs[strings.TrimSpace(c)] = true
		}
	}

	if opts.NoTrunc {
		convs["notrunc"] = true
	}

	// Open input
	var input io.Reader

	if opts.InputFile == "" || opts.InputFile == "-" {
		input = os.Stdin

		// Skip input blocks by reading and discarding (stdin is not seekable)
		if opts.Skip > 0 {
			skipBytes := opts.Skip * opts.InputBS
			if _, err := io.CopyN(io.Discard, input, skipBytes); err != nil && err != io.EOF {
				return fmt.Errorf("dd: skip failed: %w", err)
			}
		}
	} else {
		f, err := os.Open(opts.InputFile)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("dd: failed to open %q: %s", opts.InputFile, err))
			}
			if errors.Is(err, os.ErrPermission) {
				return cmderr.Wrap(cmderr.ErrPermission, fmt.Sprintf("dd: failed to open %q: %s", opts.InputFile, err))
			}
			return fmt.Errorf("dd: failed to open %q: %w", opts.InputFile, err)
		}

		defer func() { _ = f.Close() }()

		input = f

		// Skip input blocks by seeking (file is seekable)
		if opts.Skip > 0 {
			skipBytes := opts.Skip * opts.InputBS
			if _, err := f.Seek(skipBytes, io.SeekStart); err != nil {
				return fmt.Errorf("dd: seek failed: %w", err)
			}
		}
	}

	// Open output
	var output io.Writer

	var outputFile *os.File

	if opts.OutputFile == "" || opts.OutputFile == "-" {
		output = w
	} else {
		flags := os.O_WRONLY | os.O_CREATE
		if !convs["notrunc"] {
			flags |= os.O_TRUNC
		}

		f, err := os.OpenFile(opts.OutputFile, flags, 0644)
		if err != nil {
			if errors.Is(err, os.ErrPermission) {
				return cmderr.Wrap(cmderr.ErrPermission, fmt.Sprintf("dd: failed to open %q: %s", opts.OutputFile, err))
			}
			return fmt.Errorf("dd: failed to open %q: %w", opts.OutputFile, err)
		}

		defer func() { _ = f.Close() }()

		outputFile = f
		output = f

		// Seek output blocks
		if opts.Seek > 0 {
			seekBytes := opts.Seek * opts.OutputBS
			if _, err := f.Seek(seekBytes, io.SeekStart); err != nil {
				return fmt.Errorf("dd: seek failed: %w", err)
			}
		}
	}

	stats := DdStats{StartTime: time.Now()}

	// Copy data
	buf := make([]byte, opts.InputBS)
	outBuf := make([]byte, 0, opts.OutputBS)
	blocksRead := int64(0)

	for opts.Count == 0 || blocksRead < opts.Count {
		n, err := input.Read(buf)
		if n > 0 {
			stats.BytesRead += int64(n)
			if int64(n) == opts.InputBS {
				stats.BlocksIn++
			} else {
				stats.PartialIn++
			}

			blocksRead++

			// Apply conversions
			data := buf[:n]
			data = applyDdConversions(data, convs)

			// Buffer for output block size
			outBuf = append(outBuf, data...)

			// Write complete output blocks
			for int64(len(outBuf)) >= opts.OutputBS {
				written, werr := output.Write(outBuf[:opts.OutputBS])
				if werr != nil {
					return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("dd: write error: %s", werr))
				}

				stats.BytesWritten += int64(written)

				if int64(written) == opts.OutputBS {
					stats.BlocksOut++
				} else {
					stats.PartialOut++
				}

				outBuf = outBuf[opts.OutputBS:]
			}
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("dd: read error: %s", err))
		}
	}

	// Write remaining data
	if len(outBuf) > 0 {
		written, err := output.Write(outBuf)
		if err != nil {
			return fmt.Errorf("dd: write error: %w", err)
		}

		stats.BytesWritten += int64(written)
		stats.PartialOut++
	}

	// Sync if writing to file
	if outputFile != nil && convs["fsync"] {
		if err := outputFile.Sync(); err != nil {
			return fmt.Errorf("dd: sync failed: %w", err)
		}
	}

	stats.EndTime = time.Now()

	// Print statistics
	if opts.Status != "none" {
		printDdStats(opts.StatusWriter, stats, opts.Status != "noxfer")
	}

	return nil
}

func applyDdConversions(data []byte, convs map[string]bool) []byte {
	result := make([]byte, len(data))
	copy(result, data)

	if convs["lcase"] {
		for i, b := range result {
			if b >= 'A' && b <= 'Z' {
				result[i] = b + 32
			}
		}
	}

	if convs["ucase"] {
		for i, b := range result {
			if b >= 'a' && b <= 'z' {
				result[i] = b - 32
			}
		}
	}

	if convs["swab"] {
		for i := 0; i+1 < len(result); i += 2 {
			result[i], result[i+1] = result[i+1], result[i]
		}
	}

	return result
}

func printDdStats(w io.Writer, stats DdStats, showTransfer bool) {
	_, _ = fmt.Fprintf(w, "%d+%d records in\n", stats.BlocksIn, stats.PartialIn)
	_, _ = fmt.Fprintf(w, "%d+%d records out\n", stats.BlocksOut, stats.PartialOut)

	if showTransfer {
		duration := stats.EndTime.Sub(stats.StartTime)
		_, _ = fmt.Fprintf(w, "%d bytes transferred in %.6f secs (%s/sec)\n",
			stats.BytesWritten,
			duration.Seconds(),
			formatDdBytes(float64(stats.BytesWritten)/duration.Seconds()),
		)
	}
}

func formatDdBytes(b float64) string {
	const unit = 1024

	if b < unit {
		return fmt.Sprintf("%.0f B", b)
	}

	div, exp := float64(unit), 0

	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", b/div, "KMGTPE"[exp])
}

// ParseDdSize parses size specifications like "1K", "1M", "1G"
func ParseDdSize(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, cmderr.Wrap(cmderr.ErrInvalidInput, "dd: empty size")
	}

	// Find where the number ends
	i := 0
	for i < len(s) && (unicode.IsDigit(rune(s[i])) || s[i] == '.') {
		i++
	}

	numStr := s[:i]
	suffix := strings.ToUpper(s[i:])

	num, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		return 0, cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("dd: invalid number: %s", numStr))
	}

	if num < 0 {
		return 0, cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("dd: negative size: %s", numStr))
	}

	var multiplier int64

	switch suffix {
	case "", "B":
		multiplier = 1
	case "K", "KB":
		multiplier = 1024
	case "M", "MB":
		multiplier = 1024 * 1024
	case "G", "GB":
		multiplier = 1024 * 1024 * 1024
	case "T", "TB":
		multiplier = 1024 * 1024 * 1024 * 1024
	case "KIB":
		multiplier = 1024
	case "MIB":
		multiplier = 1024 * 1024
	case "GIB":
		multiplier = 1024 * 1024 * 1024
	default:
		return 0, cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("dd: unknown suffix: %s", suffix))
	}

	result := num * multiplier
	// Detect signed-integer multiplication overflow: a non-zero num that does
	// not divide back out indicates the product wrapped around int64.
	if num != 0 && result/num != multiplier {
		return 0, cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("dd: size too large: %s", s))
	}

	return result, nil
}
