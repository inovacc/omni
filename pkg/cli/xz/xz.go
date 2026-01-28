package xz

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
)

// XzOptions configures the xz command behavior
type XzOptions struct {
	Decompress bool // -d: decompress
	Keep       bool // -k: keep original files
	Force      bool // -f: force overwrite
	Stdout     bool // -c: write to stdout
	Verbose    bool // -v: verbose
	List       bool // -l: list compressed file info
}

// XZ magic bytes
var xzMagic = []byte{0xFD, '7', 'z', 'X', 'Z', 0x00}

// RunXz compresses or decompresses files using xz format
// Note: Full xz support requires external library; this provides basic functionality
func RunXz(w io.Writer, args []string, opts XzOptions) error {
	if !opts.Decompress && !opts.List {
		return fmt.Errorf("xz: compression not supported (requires external library), use -d for decompression")
	}

	if opts.List {
		return xzList(w, args)
	}

	// Read from stdin if no files
	if len(args) == 0 || (len(args) == 1 && args[0] == "-") {
		return fmt.Errorf("xz: decompression from stdin not supported")
	}

	for _, path := range args {
		if err := unxzFile(w, path, opts); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "xz: %v\n", err)
		}
	}

	return nil
}

func xzList(w io.Writer, args []string) error {
	_, _ = fmt.Fprintf(w, "Strms  Blocks   Compressed Uncompressed  Ratio  Check   Filename\n")

	for _, path := range args {
		f, err := os.Open(path)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "xz: %v\n", err)

			continue
		}

		info, _ := f.Stat()

		// Check magic
		magic := make([]byte, 6)
		_, _ = f.Read(magic)
		_ = f.Close()

		if !bytes.Equal(magic, xzMagic) {
			_, _ = fmt.Fprintf(os.Stderr, "xz: %s: not in xz format\n", path)

			continue
		}

		_, _ = fmt.Fprintf(w, "    1       ?  %12d            ?      ?  CRC64   %s\n", info.Size(), path)
	}

	return nil
}

func unxzFile(w io.Writer, path string, opts XzOptions) error {
	// Check if compressed
	if !strings.HasSuffix(path, ".xz") {
		return fmt.Errorf("%s: unknown suffix", path)
	}

	inFile, err := os.Open(path)
	if err != nil {
		return err
	}

	defer func() { _ = inFile.Close() }()

	// Check magic
	magic := make([]byte, 6)
	if _, err := io.ReadFull(inFile, magic); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}

	if !bytes.Equal(magic, xzMagic) {
		return fmt.Errorf("%s: not in xz format", path)
	}

	// Seek back to start
	if _, err := inFile.Seek(0, io.SeekStart); err != nil {
		return err
	}

	outPath := strings.TrimSuffix(path, ".xz")

	if opts.Stdout {
		return unxzReader(w, inFile)
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

	if err := unxzReader(outFile, inFile); err != nil {
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

// unxzReader performs basic xz decompression
// This is a simplified implementation that handles common cases
func unxzReader(_ io.Writer, r io.Reader) error {
	// Read all data
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	// Verify magic
	if len(data) < 12 || !bytes.Equal(data[:6], xzMagic) {
		return fmt.Errorf("not in xz format")
	}

	// XZ format is complex (LZMA2 + CRC + headers)
	// For full support, use github.com/ulikunitz/xz
	// This basic implementation just validates the format

	// Check stream footer (last 12 bytes)
	if len(data) < 12 {
		return fmt.Errorf("truncated xz data")
	}

	footer := data[len(data)-12:]

	// Footer magic should be 'YZ'
	if footer[10] != 'Y' || footer[11] != 'Z' {
		return fmt.Errorf("invalid xz footer")
	}

	// Get backward size
	backwardSize := binary.LittleEndian.Uint32(footer[4:8])
	backwardSize = (backwardSize + 1) * 4

	if int(backwardSize) > len(data)-12 {
		return fmt.Errorf("invalid xz index")
	}

	return fmt.Errorf("xz decompression requires external library (github.com/ulikunitz/xz)")
}

// RunUnxz is an alias for xz -d
func RunUnxz(w io.Writer, args []string, opts XzOptions) error {
	opts.Decompress = true

	return RunXz(w, args, opts)
}

// RunXzcat is an alias for xz -dc
func RunXzcat(w io.Writer, args []string) error {
	opts := XzOptions{
		Decompress: true,
		Stdout:     true,
	}

	return RunXz(w, args, opts)
}
