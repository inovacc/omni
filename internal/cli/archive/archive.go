package archive

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

// ArchiveOptions configures the archive command behavior
type ArchiveOptions struct {
	Create          bool   // -c: create archive
	Extract         bool   // -x: extract archive
	List            bool   // -t: list contents
	Verbose         bool   // -v: verbose output
	File            string // -f: archive file name
	Directory       string // -C: change to directory before operation
	Gzip            bool   // -z: use gzip compression
	StripComponents int    // --strip-components: strip N leading path components
	JSON            bool   // --json: output as JSON (for list mode)
}

// ArchiveEntry represents a file entry in an archive
type ArchiveEntry struct {
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	Mode    string    `json:"mode"`
	ModTime time.Time `json:"modTime"`
	IsDir   bool      `json:"isDir"`
	Type    string    `json:"type"` // "file", "dir", "symlink", "link"
}

// ArchiveListResult represents the JSON output for archive listing
type ArchiveListResult struct {
	Archive string         `json:"archive"`
	Entries []ArchiveEntry `json:"entries"`
	Count   int            `json:"count"`
}

// RunArchive handles archive operations (tar-like interface)
func RunArchive(w io.Writer, args []string, opts ArchiveOptions) error {
	if opts.Create {
		return createArchive(w, args, opts)
	}

	if opts.Extract {
		return extractArchive(w, opts)
	}

	if opts.List {
		return listArchive(w, opts)
	}

	return cmderr.Wrap(cmderr.ErrInvalidInput, "archive: must specify -c, -x, or -t")
}

func createArchive(w io.Writer, sources []string, opts ArchiveOptions) error {
	if opts.File == "" {
		return cmderr.Wrap(cmderr.ErrInvalidInput, "archive: no output file specified (-f)")
	}

	// Determine format from extension
	isZip := strings.HasSuffix(opts.File, ".zip")
	isTarGz := strings.HasSuffix(opts.File, ".tar.gz") || strings.HasSuffix(opts.File, ".tgz") || opts.Gzip

	outFile, err := os.Create(opts.File)
	if err != nil {
		return fmt.Errorf("archive: %w", err)
	}

	defer func() {
		_ = outFile.Close()
	}()

	if isZip {
		return createZipArchive(w, outFile, sources, opts)
	}

	return createTarArchive(w, outFile, sources, opts, isTarGz)
}

func extractArchive(w io.Writer, opts ArchiveOptions) error {
	if opts.File == "" {
		return cmderr.Wrap(cmderr.ErrInvalidInput, "archive: no input file specified (-f)")
	}

	isZip := strings.HasSuffix(opts.File, ".zip")

	if isZip {
		return extractZipArchive(w, opts)
	}

	return extractTarArchive(w, opts)
}

func listArchive(w io.Writer, opts ArchiveOptions) error {
	if opts.File == "" {
		return cmderr.Wrap(cmderr.ErrInvalidInput, "archive: no input file specified (-f)")
	}

	isZip := strings.HasSuffix(opts.File, ".zip")

	if isZip {
		return listZipArchive(w, opts)
	}

	return listTarArchive(w, opts)
}

// RunTar provides tar command compatibility
func RunTar(w io.Writer, args []string, opts ArchiveOptions) error {
	return RunArchive(w, args, opts)
}

// RunZip provides zip command compatibility
func RunZip(w io.Writer, args []string, opts ArchiveOptions) error {
	if opts.File == "" && len(args) > 0 {
		opts.File = args[0]
		args = args[1:]
	}

	opts.Create = true

	return RunArchive(w, args, opts)
}

// RunUnzip provides unzip command compatibility
func RunUnzip(w io.Writer, args []string, opts ArchiveOptions) error {
	if opts.File == "" && len(args) > 0 {
		opts.File = args[0]
	}

	opts.Extract = true

	return RunArchive(w, args, opts)
}
