package archive

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

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

		target := filepath.Join(destDir, name)

		if opts.Verbose {
			_, _ = fmt.Fprintln(w, name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, f.Mode()); err != nil {
				return err
			}

			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			_ = rc.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		_ = outFile.Close()
		_ = rc.Close()

		if err != nil {
			return err
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
