package archive

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

// containedTarget validates that joining name onto destDir does not escape
// destDir, defending against zip-slip path traversal (CWE-22). Entry names
// from an untrusted archive may contain "../" segments or absolute paths that
// filepath.Join/Clean would resolve outside the destination directory.
func containedTarget(destDir, name string) (string, error) {
	cleanDest, err := filepath.Abs(filepath.Clean(destDir))
	if err != nil {
		return "", cmderr.Wrap(cmderr.ErrIO, "archive: resolve destination: "+err.Error())
	}

	target := filepath.Join(cleanDest, name)
	cleanTarget := filepath.Clean(target)

	if cleanTarget != cleanDest && !strings.HasPrefix(cleanTarget, cleanDest+string(os.PathSeparator)) {
		return "", cmderr.Wrap(cmderr.ErrInvalidInput, "archive: entry escapes destination: "+name)
	}

	return target, nil
}

func createZipArchive(w io.Writer, outFile *os.File, sources []string, opts ArchiveOptions) error {
	zw := zip.NewWriter(outFile)

	defer func() {
		_ = zw.Close()
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

			// Create zip header
			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}

			// Use relative path
			relPath, err := filepath.Rel(baseDir, path)
			if err != nil {
				relPath = path
			}

			header.Name = relPath

			if info.IsDir() {
				header.Name += "/"
			} else {
				header.Method = zip.Deflate
			}

			writer, err := zw.CreateHeader(header)
			if err != nil {
				return err
			}

			if opts.Verbose {
				_, _ = fmt.Fprintln(w, header.Name)
			}

			if info.Mode().IsRegular() {
				f, err := os.Open(path)
				if err != nil {
					return err
				}

				_, err = io.Copy(writer, f)
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

func extractZipArchive(w io.Writer, opts ArchiveOptions) error {
	r, err := zip.OpenReader(opts.File)
	if err != nil {
		return fmt.Errorf("archive: %w", err)
	}

	defer func() {
		_ = r.Close()
	}()

	destDir := opts.Directory
	if destDir == "" {
		destDir = "."
	}

	// Resolve the destination once so the write-through-symlink walk
	// (refuseWriteThroughSymlink) can reason about the same absolute root that
	// containedTarget uses for its lexical check.
	cleanDest, err := filepath.Abs(filepath.Clean(destDir))
	if err != nil {
		return cmderr.Wrap(cmderr.ErrIO, "archive: resolve destination: "+err.Error())
	}

	var totalWritten int64

	for _, f := range r.File {
		name := f.Name
		if opts.StripComponents > 0 {
			parts := strings.Split(name, "/")
			if len(parts) > opts.StripComponents {
				name = strings.Join(parts[opts.StripComponents:], "/")
			} else {
				continue
			}
		}

		target, err := containedTarget(destDir, name)
		if err != nil {
			return err
		}

		if opts.Verbose {
			_, _ = fmt.Fprintln(w, name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, f.Mode()); err != nil {
				return err
			}

			continue
		}

		// Refuse to write through any pre-existing on-disk symlink in the path
		// (CWE-59); containedTarget above only checks the path string.
		if err := refuseWriteThroughSymlink(cleanDest, target); err != nil {
			return err
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC|oNoFollow, f.Mode())
		if err != nil {
			_ = rc.Close()
			return err
		}

		// Bound the copy so a decompression bomb cannot fill the disk
		// (archive-05). LimitReader caps the per-call read; the cumulative
		// counter caps the whole extraction. Mirrors extractTarArchive.
		remaining := extractByteCap - totalWritten
		n, err := io.Copy(outFile, io.LimitReader(rc, remaining+1))
		_ = outFile.Close()
		_ = rc.Close()

		if err != nil {
			return err
		}

		totalWritten += n
		if totalWritten > extractByteCap {
			return cmderr.Wrap(cmderr.ErrInvalidInput, "archive: extraction exceeds maximum allowed size")
		}
	}

	return nil
}

func listZipArchive(w io.Writer, opts ArchiveOptions) error {
	r, err := zip.OpenReader(opts.File)
	if err != nil {
		return fmt.Errorf("archive: %w", err)
	}

	defer func() {
		_ = r.Close()
	}()

	var entries []ArchiveEntry

	for _, f := range r.File {
		if opts.JSON {
			entryType := "file"

			isDir := f.FileInfo().IsDir()
			if isDir {
				entryType = "dir"
			}

			entries = append(entries, ArchiveEntry{
				Name:    f.Name,
				Size:    int64(f.UncompressedSize64),
				Mode:    f.Mode().String(),
				ModTime: f.Modified,
				IsDir:   isDir,
				Type:    entryType,
			})
		} else if opts.Verbose {
			_, _ = fmt.Fprintf(w, "%s %8d %s %s\n",
				f.Mode().String(),
				f.UncompressedSize64,
				f.Modified.Format("2006-01-02 15:04"),
				f.Name)
		} else {
			_, _ = fmt.Fprintln(w, f.Name)
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
