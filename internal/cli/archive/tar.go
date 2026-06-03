package archive

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

// maxExtractTotalBytes caps the cumulative number of bytes written during a
// single archive extraction to guard against decompression-bomb DoS
// (archive-05). 10 GiB is generous for legitimate CI/CD artifacts while still
// bounding a small malicious archive that inflates to fill the disk.
const maxExtractTotalBytes int64 = 10 << 30

// secureJoin joins name onto an absolute, cleaned destDir and guarantees the
// result stays within destDir. It rejects absolute entry names and any path
// that escapes the destination via ".." segments (tar-slip / zip-slip,
// archive-01/02/03/04). The caller passes an already-absolute, cleaned
// cleanDest so the work is done once per extraction.
func secureJoin(cleanDest, name string) (string, error) {
	if filepath.IsAbs(name) {
		return "", cmderr.Wrap(cmderr.ErrInvalidInput, "archive: entry has absolute path: "+name)
	}

	target := filepath.Clean(filepath.Join(cleanDest, name))
	if target != cleanDest && !strings.HasPrefix(target, cleanDest+string(os.PathSeparator)) {
		return "", cmderr.Wrap(cmderr.ErrInvalidInput, "archive: entry escapes destination: "+name)
	}

	return target, nil
}

func createTarArchive(w io.Writer, outFile *os.File, sources []string, opts ArchiveOptions, useGzip bool) error {
	var tw *tar.Writer

	if useGzip {
		gw := gzip.NewWriter(outFile)

		defer func() {
			_ = gw.Close()
		}()

		tw = tar.NewWriter(gw)
	} else {
		tw = tar.NewWriter(outFile)
	}

	defer func() {
		_ = tw.Close()
	}()

	baseDir := opts.Directory
	if baseDir == "" {
		baseDir = "."
	}

	for _, source := range sources {
		// Handle absolute paths - don't join with baseDir
		sourcePath := source
		if !filepath.IsAbs(source) {
			sourcePath = filepath.Join(baseDir, source)
		}

		err := filepath.WalkDir(sourcePath, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}

			info, err := d.Info()
			if err != nil {
				return err
			}

			// Create a tar header
			header, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return err
			}

			// Use relative path
			relPath, err := filepath.Rel(baseDir, path)
			if err != nil {
				relPath = path
			}

			header.Name = relPath

			// Handle symlinks
			if info.Mode()&os.ModeSymlink != 0 {
				link, err := os.Readlink(path)
				if err != nil {
					return err
				}

				header.Linkname = link
			}

			if err := tw.WriteHeader(header); err != nil {
				return err
			}

			if opts.Verbose {
				_, _ = fmt.Fprintln(w, header.Name)
			}

			// Write file content if regular file
			if info.Mode().IsRegular() {
				f, err := os.Open(path)
				if err != nil {
					return err
				}

				_, err = io.Copy(tw, f)
				_ = f.Close()

				if err != nil {
					return err
				}
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("archive: %w", err)
		}
	}

	return nil
}

func extractTarArchive(w io.Writer, opts ArchiveOptions) error {
	f, err := os.Open(opts.File)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("archive: %s", err))
		}
		return fmt.Errorf("archive: %w", err)
	}

	defer func() {
		_ = f.Close()
	}()

	var tr *tar.Reader

	// Check for gzip
	isTarGz := strings.HasSuffix(opts.File, ".tar.gz") || strings.HasSuffix(opts.File, ".tgz") || opts.Gzip
	if isTarGz {
		gr, err := gzip.NewReader(f)
		if err != nil {
			return fmt.Errorf("archive: %w", err)
		}

		defer func() {
			_ = gr.Close()
		}()

		tr = tar.NewReader(gr)
	} else {
		tr = tar.NewReader(f)
	}

	destDir := opts.Directory
	if destDir == "" {
		destDir = "."
	}

	// Resolve the destination once so every entry can be containment-checked
	// against it (tar-slip / symlink / hardlink escape: archive-01/03/04).
	cleanDest, err := filepath.Abs(filepath.Clean(destDir))
	if err != nil {
		return fmt.Errorf("archive: %w", err)
	}

	var totalWritten int64

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("archive: %w", err)
		}

		// Strip leading components if requested
		name := header.Name
		if opts.StripComponents > 0 {
			parts := strings.Split(name, "/")
			if len(parts) > opts.StripComponents {
				name = strings.Join(parts[opts.StripComponents:], "/")
			} else {
				continue
			}
		}

		// Containment check: reject absolute names and ".." escapes before any
		// filesystem write (archive-01).
		target, err := secureJoin(cleanDest, name)
		if err != nil {
			return err
		}

		if opts.Verbose {
			_, _ = fmt.Fprintln(w, name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}

			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			// Bound the copy so a decompression bomb cannot fill the disk
			// (archive-05). LimitReader caps the per-call read; the cumulative
			// counter caps the whole extraction.
			remaining := maxExtractTotalBytes - totalWritten
			n, err := io.Copy(outFile, io.LimitReader(tr, remaining+1))
			_ = outFile.Close()

			if err != nil {
				return err
			}

			totalWritten += n
			if totalWritten > maxExtractTotalBytes {
				return cmderr.Wrap(cmderr.ErrInvalidInput, "archive: extraction exceeds maximum allowed size")
			}
		case tar.TypeSymlink:
			// Validate the symlink destination stays within destDir so a later
			// entry cannot be written through an escaping symlink
			// (archive-03). Reject absolute Linkname outright; resolve a
			// relative Linkname against the link's own parent dir, then require
			// the result to remain contained.
			if filepath.IsAbs(header.Linkname) {
				return cmderr.Wrap(cmderr.ErrInvalidInput, "archive: symlink target is absolute: "+header.Linkname)
			}

			resolved := filepath.Clean(filepath.Join(filepath.Dir(target), header.Linkname))
			if resolved != cleanDest && !strings.HasPrefix(resolved, cleanDest+string(os.PathSeparator)) {
				return cmderr.Wrap(cmderr.ErrInvalidInput, "archive: symlink target escapes destination: "+header.Linkname)
			}

			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}

			if err := os.Symlink(header.Linkname, target); err != nil {
				return err
			}
		case tar.TypeLink:
			// Both endpoints must be inside destDir; Linkname is attacker
			// controlled and may contain ".." segments (archive-04).
			linkTarget, err := secureJoin(cleanDest, header.Linkname)
			if err != nil {
				return cmderr.Wrap(cmderr.ErrInvalidInput, "archive: hardlink target escapes destination: "+header.Linkname)
			}

			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}

			if err := os.Link(linkTarget, target); err != nil {
				return err
			}
		}
	}

	return nil
}

func listTarArchive(w io.Writer, opts ArchiveOptions) error {
	f, err := os.Open(opts.File)
	if err != nil {
		return fmt.Errorf("archive: %w", err)
	}

	defer func() {
		_ = f.Close()
	}()

	var tr *tar.Reader

	isTarGz := strings.HasSuffix(opts.File, ".tar.gz") || strings.HasSuffix(opts.File, ".tgz") || opts.Gzip
	if isTarGz {
		gr, err := gzip.NewReader(f)
		if err != nil {
			return fmt.Errorf("archive: %w", err)
		}

		defer func() {
			_ = gr.Close()
		}()

		tr = tar.NewReader(gr)
	} else {
		tr = tar.NewReader(f)
	}

	var entries []ArchiveEntry

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("archive: %w", err)
		}

		if opts.JSON {
			entryType := "file"
			isDir := false

			switch header.Typeflag {
			case tar.TypeDir:
				entryType = "dir"
				isDir = true
			case tar.TypeSymlink:
				entryType = "symlink"
			case tar.TypeLink:
				entryType = "link"
			}

			entries = append(entries, ArchiveEntry{
				Name:    header.Name,
				Size:    header.Size,
				Mode:    os.FileMode(header.Mode).String(),
				ModTime: header.ModTime,
				IsDir:   isDir,
				Type:    entryType,
			})
		} else if opts.Verbose {
			_, _ = fmt.Fprintf(w, "%s %8d %s %s\n",
				os.FileMode(header.Mode).String(),
				header.Size,
				header.ModTime.Format("2006-01-02 15:04"),
				header.Name)
		} else {
			_, _ = fmt.Fprintln(w, header.Name)
		}
	}

	if opts.JSON {
		return json.NewEncoder(w).Encode(ArchiveListResult{
			Archive: opts.File,
			Entries: entries,
			Count:   len(entries),
		})
	}

	return nil
}
