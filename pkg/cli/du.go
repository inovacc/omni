package cli

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// DUOptions configures the du command behavior
type DUOptions struct {
	All            bool  // -a: write counts for all files, not just directories
	SummarizeOnly  bool  // -s: display only a total for each argument
	HumanReadable  bool  // -h: print sizes in human readable format
	ByteCount      bool  // -b: equivalent to --apparent-size --block-size=1
	BlockSize      int64 // --block-size: scale sizes by SIZE
	Total          bool  // -c: produce a grand total
	MaxDepth       int   // --max-depth: print total for directory only if it is N or fewer levels
	OneFileSystem  bool  // -x: skip directories on different file systems
	ApparentSize   bool  // --apparent-size: print apparent sizes rather than disk usage
	NullTerminator bool  // -0: end each output line with NUL, not newline
}

// DUResult represents the result of a du operation
type DUResult struct {
	Path string `json:"path"`
	Size int64  `json:"size"`
}

// RunDU executes the du command
func RunDU(w io.Writer, args []string, opts DUOptions) error {
	if opts.BlockSize == 0 {
		opts.BlockSize = 1024 // Default 1K blocks
	}

	if opts.ByteCount {
		opts.BlockSize = 1
		opts.ApparentSize = true
	}

	paths := args
	if len(paths) == 0 {
		paths = []string{"."}
	}

	var grandTotal int64
	terminator := "\n"
	if opts.NullTerminator {
		terminator = "\x00"
	}

	for _, path := range paths {
		total, err := duPath(w, path, opts, 0, terminator)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "du: %s: %v\n", path, err)
			continue
		}
		grandTotal += total
	}

	if opts.Total && len(paths) > 1 {
		printDUSize(w, grandTotal, "total", opts, terminator)
	}

	return nil
}

func duPath(w io.Writer, path string, opts DUOptions, depth int, terminator string) (int64, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return 0, err
	}

	// If it's a file, just return its size
	if !info.IsDir() {
		size := info.Size()
		if opts.All || opts.SummarizeOnly {
			printDUSize(w, size, path, opts, terminator)
		}
		return size, nil
	}

	// It's a directory - walk it
	var totalSize int64
	entries := make(map[string]int64)

	err = filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		fileInfo, err := d.Info()
		if err != nil {
			return nil
		}

		size := fileInfo.Size()
		totalSize += size

		// Track directory sizes for non-summarize mode
		if d.IsDir() && p != path {
			rel, _ := filepath.Rel(path, p)
			relDepth := len(filepath.SplitList(rel))
			if opts.MaxDepth == 0 || relDepth <= opts.MaxDepth {
				entries[p] = 0 // Will be calculated
			}
		}

		if opts.All && !d.IsDir() {
			rel, _ := filepath.Rel(path, p)
			relDepth := len(filepath.SplitList(rel))
			if opts.MaxDepth == 0 || relDepth <= opts.MaxDepth {
				printDUSize(w, size, p, opts, terminator)
			}
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	// Calculate and print directory sizes if not summarize-only
	if !opts.SummarizeOnly {
		// Get sorted directory list for consistent output
		var dirs []string
		for dir := range entries {
			dirs = append(dirs, dir)
		}
		sort.Strings(dirs)

		for _, dir := range dirs {
			dirSize := calculateDirSize(dir)
			rel, _ := filepath.Rel(path, dir)
			relDepth := len(filepath.SplitList(rel))
			if opts.MaxDepth == 0 || relDepth <= opts.MaxDepth {
				printDUSize(w, dirSize, dir, opts, terminator)
			}
		}
	}

	// Always print the root path
	printDUSize(w, totalSize, path, opts, terminator)

	return totalSize, nil
}

func calculateDirSize(path string) int64 {
	var size int64
	_ = filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if info, err := d.Info(); err == nil {
			size += info.Size()
		}
		return nil
	})
	return size
}

func printDUSize(w io.Writer, size int64, path string, opts DUOptions, terminator string) {
	var sizeStr string

	if opts.HumanReadable {
		sizeStr = formatHumanSize(size)
	} else {
		blocks := (size + opts.BlockSize - 1) / opts.BlockSize
		sizeStr = fmt.Sprintf("%d", blocks)
	}

	_, _ = fmt.Fprintf(w, "%s\t%s%s", sizeStr, path, terminator)
}

// formatHumanSize formats bytes into human readable form
func formatHumanSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%c", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// DiskUsage returns the total size of a path
func DiskUsage(path string) (int64, error) {
	var size int64
	err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if info, err := d.Info(); err == nil && !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}
