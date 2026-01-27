package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// LsOptions configures the ls command behavior
type LsOptions struct {
	All          bool // -a: show hidden files
	AlmostAll    bool // -A: show hidden except . and ..
	LongFormat   bool // -l: long listing format
	HumanReadble bool // -h: human readable sizes
	OnePerLine   bool // -1: one entry per line
	Recursive    bool // -R: recursive listing
	Reverse      bool // -r: reverse order
	SortByTime   bool // -t: sort by modification time
	SortBySize   bool // -S: sort by size
	NoSort       bool // -U: do not sort
	Directory    bool // -d: list directories themselves, not contents
	Classify     bool // -F: append indicator (*/=>@|)
	Inode        bool // -i: show inode numbers
	JSON         bool // --json: output in JSON format
}

// FileEntry represents a file entry for ls output
type FileEntry struct {
	Name    string      `json:"name"`
	Size    int64       `json:"size"`
	Mode    string      `json:"mode"`
	ModTime time.Time   `json:"modTime"`
	IsDir   bool        `json:"isDir"`
	IsLink  bool        `json:"isLink"`
	Link    string      `json:"link,omitempty"`
	Inode   uint64      `json:"inode,omitempty"`
	NLink   uint64      `json:"nlink,omitempty"`
	UID     uint32      `json:"uid,omitempty"`
	GID     uint32      `json:"gid,omitempty"`
	perm    fs.FileMode // internal use for sorting
}

// RunLs executes the ls command with the given options
func RunLs(w io.Writer, args []string, opts LsOptions) error {
	paths := args
	if len(paths) == 0 {
		paths = []string{"."}
	}

	var (
		allEntries []FileEntry
		errors     []error
	)

	for i, path := range paths {
		entries, err := listPath(path, opts)
		if err != nil {
			errors = append(errors, fmt.Errorf("ls: cannot access '%s': %w", path, err))
			continue
		}

		if opts.JSON {
			allEntries = append(allEntries, entries...)
		} else {
			// Print header for multiple paths
			if len(paths) > 1 {
				if i > 0 {
					_, _ = fmt.Fprintln(w)
				}

				_, _ = fmt.Fprintf(w, "%s:\n", path)
			}

			printEntries(w, entries, opts)
		}
	}

	if opts.JSON {
		return json.NewEncoder(w).Encode(allEntries)
	}

	if len(errors) > 0 {
		return errors[0]
	}

	return nil
}

func listPath(path string, opts LsOptions) ([]FileEntry, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}

	// If it's a file or -d flag, just return the single entry
	if !info.IsDir() || opts.Directory {
		entry := fileInfoToEntry(path, info)
		return []FileEntry{entry}, nil
	}

	// Read directory contents
	dirEntries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var entries []FileEntry

	for _, de := range dirEntries {
		name := de.Name()

		// Filter hidden files
		if !opts.All && !opts.AlmostAll && strings.HasPrefix(name, ".") {
			continue
		}

		if opts.AlmostAll && (name == "." || name == "..") {
			continue
		}

		fullPath := filepath.Join(path, name)

		info, err := de.Info()
		if err != nil {
			continue
		}

		entry := fileInfoToEntry(fullPath, info)
		entry.Name = name // Use relative name for display
		entries = append(entries, entry)
	}

	// Add . and .. for -a flag
	if opts.All {
		if dotInfo, err := os.Lstat(path); err == nil {
			dot := fileInfoToEntry(path, dotInfo)
			dot.Name = "."
			entries = append([]FileEntry{dot}, entries...)
		}

		parentPath := filepath.Dir(path)
		if parentInfo, err := os.Lstat(parentPath); err == nil {
			dotdot := fileInfoToEntry(parentPath, parentInfo)
			dotdot.Name = ".."
			entries = append([]FileEntry{entries[0], dotdot}, entries[1:]...)
		}
	}

	// Sort entries
	sortEntries(entries, opts)

	// Handle recursive listing
	if opts.Recursive {
		var recursiveEntries []FileEntry

		recursiveEntries = append(recursiveEntries, entries...)
		for _, entry := range entries {
			if entry.IsDir && entry.Name != "." && entry.Name != ".." {
				subPath := filepath.Join(path, entry.Name)

				subEntries, err := listPath(subPath, opts)
				if err == nil {
					recursiveEntries = append(recursiveEntries, subEntries...)
				}
			}
		}

		return recursiveEntries, nil
	}

	return entries, nil
}

func fileInfoToEntry(path string, info fs.FileInfo) FileEntry {
	entry := FileEntry{
		Name:    info.Name(),
		Size:    info.Size(),
		Mode:    info.Mode().String(),
		ModTime: info.ModTime(),
		IsDir:   info.IsDir(),
		IsLink:  info.Mode()&os.ModeSymlink != 0,
		perm:    info.Mode(),
	}

	// Resolve symlink target
	if entry.IsLink {
		if target, err := os.Readlink(path); err == nil {
			entry.Link = target
		}
	}

	// Get inode and other Unix-specific info (platform-specific)
	fillUnixInfo(&entry, info)

	return entry
}

func sortEntries(entries []FileEntry, opts LsOptions) {
	if opts.NoSort {
		return
	}

	sort.Slice(entries, func(i, j int) bool {
		// Keep . and .. at the top
		if entries[i].Name == "." {
			return true
		}

		if entries[j].Name == "." {
			return false
		}

		if entries[i].Name == ".." {
			return true
		}

		if entries[j].Name == ".." {
			return false
		}

		var result bool
		if opts.SortByTime {
			result = entries[i].ModTime.After(entries[j].ModTime)
		} else if opts.SortBySize {
			result = entries[i].Size > entries[j].Size
		} else {
			result = strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
		}

		if opts.Reverse {
			return !result
		}

		return result
	})
}

func printEntries(w io.Writer, entries []FileEntry, opts LsOptions) {
	if opts.LongFormat {
		printLongFormat(w, entries, opts)
	} else if opts.OnePerLine {
		printOnePerLine(w, entries, opts)
	} else {
		printSimple(w, entries, opts)
	}
}

func printLongFormat(w io.Writer, entries []FileEntry, opts LsOptions) {
	for _, e := range entries {
		name := e.Name
		if opts.Classify {
			name += classifyIndicator(e)
		}

		if e.IsLink && e.Link != "" {
			name = fmt.Sprintf("%s -> %s", name, e.Link)
		}

		size := fmt.Sprintf("%d", e.Size)
		if opts.HumanReadble {
			size = humanReadableSize(e.Size)
		}

		timeStr := e.ModTime.Format("Jan _2 15:04")
		if e.ModTime.Year() != time.Now().Year() {
			timeStr = e.ModTime.Format("Jan _2  2006")
		}

		if opts.Inode {
			_, _ = fmt.Fprintf(w, "%8d ", e.Inode)
		}

		_, _ = fmt.Fprintf(w, "%s %8s %s %s\n", e.Mode, size, timeStr, name)
	}
}

func printOnePerLine(w io.Writer, entries []FileEntry, opts LsOptions) {
	for _, e := range entries {
		name := e.Name
		if opts.Classify {
			name += classifyIndicator(e)
		}

		_, _ = fmt.Fprintln(w, name)
	}
}

func printSimple(w io.Writer, entries []FileEntry, opts LsOptions) {
	names := make([]string, 0, len(entries))

	for _, e := range entries {
		name := e.Name
		if opts.Classify {
			name += classifyIndicator(e)
		}

		names = append(names, name)
	}

	_, _ = fmt.Fprintln(w, strings.Join(names, "  "))
}

func classifyIndicator(e FileEntry) string {
	if e.IsDir {
		return "/"
	}

	if e.IsLink {
		return "@"
	}

	if e.perm&0111 != 0 {
		return "*"
	}

	return ""
}

func humanReadableSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d", size)
	}

	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f%c", float64(size)/float64(div), "KMGTPE"[exp])
}

// Ls returns a list of file names in the given path (simple version for compatibility)
func Ls(path string) ([]string, error) {
	if path == "" {
		path = "."
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	out := make([]string, 0, len(entries))
	for _, e := range entries {
		out = append(out, e.Name())
	}

	return out, nil
}
