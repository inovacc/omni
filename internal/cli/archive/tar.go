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

		target := filepath.Join(destDir, name)

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

			_, err = io.Copy(outFile, tr)
			_ = outFile.Close()

			if err != nil {
				return err
			}
		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}

			if err := os.Symlink(header.Linkname, target); err != nil {
				return err
			}
		case tar.TypeLink:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}

			linkTarget := filepath.Join(destDir, header.Linkname)
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
