package bzip2

import (
	"compress/bzip2"
	"fmt"
	"io"
	"os"
	"strings"
)

// Bzip2Options configures the bzip2 command behavior
type Bzip2Options struct {
	Decompress bool // -d: decompress
	Keep       bool // -k: keep original files
	Force      bool // -f: force overwrite
	Stdout     bool // -c: write to stdout
	Verbose    bool // -v: verbose
}

// RunBzip2 compresses or decompresses files using bzip2
// Note: Go stdlib only supports bzip2 decompression, not compression
func RunBzip2(w io.Writer, args []string, opts Bzip2Options) error {
	if !opts.Decompress {
		return fmt.Errorf("bzip2: compression not supported (Go stdlib limitation), use -d for decompression")
	}

	// Read from stdin if no files
	if len(args) == 0 || (len(args) == 1 && args[0] == "-") {
		return bunzip2Reader(w, os.Stdin)
	}

	for _, path := range args {
		if err := bunzip2File(w, path, opts); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "bzip2: %v\n", err)
		}
	}

	return nil
}

func bunzip2Reader(w io.Writer, r io.Reader) error {
	br := bzip2.NewReader(r)
	_, err := io.Copy(w, br)

	return err
}

func bunzip2File(w io.Writer, path string, opts Bzip2Options) error {
	// Check if compressed
	if !strings.HasSuffix(path, ".bz2") {
		return fmt.Errorf("%s: unknown suffix", path)
	}

	inFile, err := os.Open(path)
	if err != nil {
		return err
	}

	defer func() { _ = inFile.Close() }()

	outPath := strings.TrimSuffix(path, ".bz2")

	if opts.Stdout {
		return bunzip2Reader(w, inFile)
	}

	// Check if output exists
	if !opts.Force {
		if _, err := os.Stat(outPath); err == nil {
			return fmt.Errorf("%s already exists; use -f to overwrite", outPath)
		}
	}

	outFile, err := os.Create(outPath)
	if err != nil {
		return err
	}

	defer func() { _ = outFile.Close() }()

	if err := bunzip2Reader(outFile, inFile); err != nil {
		return err
	}

	// Copy file mode
	if info, err := os.Stat(path); err == nil {
		_ = os.Chmod(outPath, info.Mode())
	}

	if opts.Verbose {
		_, _ = fmt.Fprintf(w, "%s:\t -- replaced with %s\n", path, outPath)
	}

	// Remove original if not keeping
	if !opts.Keep {
		return os.Remove(path)
	}

	return nil
}

// RunBunzip2 is an alias for bzip2 -d
func RunBunzip2(w io.Writer, args []string, opts Bzip2Options) error {
	opts.Decompress = true

	return RunBzip2(w, args, opts)
}

// RunBzcat is an alias for bzip2 -dc
func RunBzcat(w io.Writer, args []string) error {
	if len(args) == 0 {
		return bunzip2Reader(w, os.Stdin)
	}

	for _, path := range args {
		// Add .bz2 if not present
		if !strings.HasSuffix(path, ".bz2") {
			if _, err := os.Stat(path); os.IsNotExist(err) {
				path += ".bz2"
			}
		}

		f, err := os.Open(path)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "bzcat: %v\n", err)

			continue
		}

		err = bunzip2Reader(w, f)
		_ = f.Close()

		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "bzcat: %s: %v\n", path, err)
		}
	}

	return nil
}
