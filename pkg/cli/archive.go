package cli

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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
	return fmt.Errorf("archive: must specify -c, -x, or -t")
}

func createArchive(w io.Writer, sources []string, opts ArchiveOptions) error {
	if opts.File == "" {
		return fmt.Errorf("archive: no output file specified (-f)")
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

			// Create tar header
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

func extractArchive(w io.Writer, opts ArchiveOptions) error {
	if opts.File == "" {
		return fmt.Errorf("archive: no input file specified (-f)")
	}

	isZip := strings.HasSuffix(opts.File, ".zip")

	if isZip {
		return extractZipArchive(w, opts)
	}
	return extractTarArchive(w, opts)
}

func extractTarArchive(w io.Writer, opts ArchiveOptions) error {
	f, err := os.Open(opts.File)
	if err != nil {
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

func listArchive(w io.Writer, opts ArchiveOptions) error {
	if opts.File == "" {
		return fmt.Errorf("archive: no input file specified (-f)")
	}

	isZip := strings.HasSuffix(opts.File, ".zip")

	if isZip {
		return listZipArchive(w, opts)
	}
	return listTarArchive(w, opts)
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

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("archive: %w", err)
		}

		if opts.Verbose {
			_, _ = fmt.Fprintf(w, "%s %8d %s %s\n",
				os.FileMode(header.Mode).String(),
				header.Size,
				header.ModTime.Format("2006-01-02 15:04"),
				header.Name)
		} else {
			_, _ = fmt.Fprintln(w, header.Name)
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

	for _, f := range r.File {
		if opts.Verbose {
			_, _ = fmt.Fprintf(w, "%s %8d %s %s\n",
				f.Mode().String(),
				f.UncompressedSize64,
				f.Modified.Format("2006-01-02 15:04"),
				f.Name)
		} else {
			_, _ = fmt.Fprintln(w, f.Name)
		}
	}

	return nil
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
