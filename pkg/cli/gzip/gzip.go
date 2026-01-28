package gzip

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// GzipOptions configures the gzip command behavior
type GzipOptions struct {
	Decompress bool // -d: decompress
	Keep       bool // -k: keep original files
	Force      bool // -f: force overwrite
	Stdout     bool // -c: write to stdout
	Verbose    bool // -v: verbose
	Level      int  // -1 to -9: compression level
}

// RunGzip compresses or decompresses files
func RunGzip(w io.Writer, args []string, opts GzipOptions) error {
	if opts.Level == 0 {
		opts.Level = gzip.DefaultCompression
	}

	// Read from stdin if no files
	if len(args) == 0 || (len(args) == 1 && args[0] == "-") {
		if opts.Decompress {
			return gunzipReader(w, os.Stdin)
		}

		return gzipReader(w, os.Stdin, opts.Level)
	}

	for _, path := range args {
		if opts.Decompress {
			if err := gunzipFile(w, path, opts); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "gzip: %v\n", err)
			}
		} else {
			if err := gzipFile(w, path, opts); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "gzip: %v\n", err)
			}
		}
	}

	return nil
}

func gzipReader(w io.Writer, r io.Reader, level int) error {
	gw, err := gzip.NewWriterLevel(w, level)
	if err != nil {
		return err
	}

	_, err = io.Copy(gw, r)
	if err != nil {
		_ = gw.Close()
		return err
	}

	return gw.Close()
}

func gunzipReader(w io.Writer, r io.Reader) error {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}

	defer func() { _ = gr.Close() }()

	_, err = io.Copy(w, gr)

	return err
}

func gzipFile(w io.Writer, path string, opts GzipOptions) error {
	// Check if already compressed
	if strings.HasSuffix(path, ".gz") {
		return fmt.Errorf("%s already has .gz suffix", path)
	}

	inFile, err := os.Open(path)
	if err != nil {
		return err
	}

	defer func() { _ = inFile.Close() }()

	outPath := path + ".gz"

	if opts.Stdout {
		if opts.Verbose {
			_, _ = fmt.Fprintf(os.Stderr, "%s:\t", path)
		}

		return gzipReader(w, inFile, opts.Level)
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

	if err := gzipReader(outFile, inFile, opts.Level); err != nil {
		return err
	}

	// Copy file mode
	if info, err := os.Stat(path); err == nil {
		_ = os.Chmod(outPath, info.Mode())
	}

	if opts.Verbose {
		inInfo, _ := os.Stat(path)
		outInfo, _ := os.Stat(outPath)

		if inInfo != nil && outInfo != nil {
			ratio := 100.0 - (float64(outInfo.Size())/float64(inInfo.Size()))*100
			_, _ = fmt.Fprintf(w, "%s:\t%.1f%% -- replaced with %s\n", path, ratio, outPath)
		}
	}

	// Remove original if not keeping
	if !opts.Keep {
		return os.Remove(path)
	}

	return nil
}

func gunzipFile(w io.Writer, path string, opts GzipOptions) error {
	// Check if compressed
	if !strings.HasSuffix(path, ".gz") {
		return fmt.Errorf("%s: unknown suffix", path)
	}

	inFile, err := os.Open(path)
	if err != nil {
		return err
	}

	defer func() { _ = inFile.Close() }()

	outPath := strings.TrimSuffix(path, ".gz")

	if opts.Stdout {
		return gunzipReader(w, inFile)
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

	if err := gunzipReader(outFile, inFile); err != nil {
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

// RunGunzip is an alias for gzip -d
func RunGunzip(w io.Writer, args []string, opts GzipOptions) error {
	opts.Decompress = true

	return RunGzip(w, args, opts)
}

// RunZcat is an alias for gzip -dc
func RunZcat(w io.Writer, args []string) error {
	if len(args) == 0 {
		return gunzipReader(w, os.Stdin)
	}

	for _, path := range args {
		// Add .gz if not present
		if !strings.HasSuffix(path, ".gz") {
			if _, err := os.Stat(path); os.IsNotExist(err) {
				path += ".gz"
			}
		}

		f, err := os.Open(path)
		if err != nil {
			// Try without .gz
			path = strings.TrimSuffix(path, ".gz")

			f, err = os.Open(path)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "zcat: %v\n", err)

				continue
			}
		}

		// Check if it's gzipped by trying to read
		if filepath.Ext(path) == ".gz" {
			err = gunzipReader(w, f)
		} else {
			_, err = io.Copy(w, f)
		}

		_ = f.Close()

		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "zcat: %s: %v\n", path, err)
		}
	}

	return nil
}
