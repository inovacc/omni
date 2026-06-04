package archive

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

func TestRunArchive_NoOperation(t *testing.T) {
	var buf bytes.Buffer

	err := RunArchive(&buf, []string{}, ArchiveOptions{})
	if err == nil {
		t.Error("RunArchive() should error when no operation specified")
	}
}

func TestRunArchive_CreateTar(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "archive_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create test files
	testFile := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(testFile, []byte("hello world"), 0644)

	archivePath := filepath.Join(tmpDir, "test.tar")

	var buf bytes.Buffer

	err = RunArchive(&buf, []string{"test.txt"}, ArchiveOptions{
		Create:    true,
		File:      archivePath,
		Directory: tmpDir,
	})
	if err != nil {
		t.Fatalf("RunArchive() create tar error = %v", err)
	}

	// Verify archive exists
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		t.Error("RunArchive() should create tar file")
	}
}

func TestRunArchive_CreateTarGz(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "archive_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create test file
	testFile := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(testFile, []byte("hello world"), 0644)

	archivePath := filepath.Join(tmpDir, "test.tar.gz")

	var buf bytes.Buffer

	err = RunArchive(&buf, []string{"test.txt"}, ArchiveOptions{
		Create:    true,
		File:      archivePath,
		Directory: tmpDir,
		Gzip:      true,
	})
	if err != nil {
		t.Fatalf("RunArchive() create tar.gz error = %v", err)
	}

	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		t.Error("RunArchive() should create tar.gz file")
	}
}

func TestRunArchive_CreateZip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "archive_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create test file
	testFile := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(testFile, []byte("hello world"), 0644)

	archivePath := filepath.Join(tmpDir, "test.zip")

	var buf bytes.Buffer

	err = RunArchive(&buf, []string{"test.txt"}, ArchiveOptions{
		Create:    true,
		File:      archivePath,
		Directory: tmpDir,
	})
	if err != nil {
		t.Fatalf("RunArchive() create zip error = %v", err)
	}

	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		t.Error("RunArchive() should create zip file")
	}
}

func TestRunArchive_ExtractTar(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "archive_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a tar archive first
	testFile := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(testFile, []byte("hello world"), 0644)

	archivePath := filepath.Join(tmpDir, "test.tar")

	var buf bytes.Buffer

	_ = RunArchive(&buf, []string{"test.txt"}, ArchiveOptions{
		Create:    true,
		File:      archivePath,
		Directory: tmpDir,
	})

	// Remove original file
	_ = os.Remove(testFile)

	// Extract
	extractDir := filepath.Join(tmpDir, "extracted")
	_ = os.MkdirAll(extractDir, 0755)

	err = RunArchive(&buf, []string{}, ArchiveOptions{
		Extract:   true,
		File:      archivePath,
		Directory: extractDir,
	})
	if err != nil {
		t.Fatalf("RunArchive() extract tar error = %v", err)
	}

	// Verify extracted file
	extractedFile := filepath.Join(extractDir, "test.txt")
	if _, err := os.Stat(extractedFile); os.IsNotExist(err) {
		t.Error("RunArchive() should extract file from tar")
	}
}

func TestRunArchive_ExtractZip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "archive_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a zip archive first
	testFile := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(testFile, []byte("hello from zip"), 0644)

	archivePath := filepath.Join(tmpDir, "test.zip")

	var buf bytes.Buffer

	_ = RunArchive(&buf, []string{"test.txt"}, ArchiveOptions{
		Create:    true,
		File:      archivePath,
		Directory: tmpDir,
	})

	// Remove original file
	_ = os.Remove(testFile)

	// Extract
	extractDir := filepath.Join(tmpDir, "extracted")
	_ = os.MkdirAll(extractDir, 0755)

	err = RunArchive(&buf, []string{}, ArchiveOptions{
		Extract:   true,
		File:      archivePath,
		Directory: extractDir,
	})
	if err != nil {
		t.Fatalf("RunArchive() extract zip error = %v", err)
	}

	extractedFile := filepath.Join(extractDir, "test.txt")
	if _, err := os.Stat(extractedFile); os.IsNotExist(err) {
		t.Error("RunArchive() should extract file from zip")
	}
}

func TestRunArchive_ListTar(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "archive_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create test files
	_ = os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content1"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("content2"), 0644)

	archivePath := filepath.Join(tmpDir, "test.tar")

	var buf bytes.Buffer

	_ = RunArchive(&buf, []string{"file1.txt", "file2.txt"}, ArchiveOptions{
		Create:    true,
		File:      archivePath,
		Directory: tmpDir,
	})

	// List contents
	buf.Reset()

	err = RunArchive(&buf, []string{}, ArchiveOptions{
		List: true,
		File: archivePath,
	})
	if err != nil {
		t.Fatalf("RunArchive() list tar error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "file1.txt") || !strings.Contains(output, "file2.txt") {
		t.Errorf("RunArchive() list should show files: %s", output)
	}
}

func TestRunArchive_ListZip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "archive_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create test file
	_ = os.WriteFile(filepath.Join(tmpDir, "zipfile.txt"), []byte("zip content"), 0644)

	archivePath := filepath.Join(tmpDir, "test.zip")

	var buf bytes.Buffer

	_ = RunArchive(&buf, []string{"zipfile.txt"}, ArchiveOptions{
		Create:    true,
		File:      archivePath,
		Directory: tmpDir,
	})

	// List contents
	buf.Reset()

	err = RunArchive(&buf, []string{}, ArchiveOptions{
		List: true,
		File: archivePath,
	})
	if err != nil {
		t.Fatalf("RunArchive() list zip error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "zipfile.txt") {
		t.Errorf("RunArchive() list should show files: %s", output)
	}
}

func TestRunArchive_Verbose(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "archive_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "verbose.txt")
	_ = os.WriteFile(testFile, []byte("verbose test"), 0644)

	archivePath := filepath.Join(tmpDir, "test.tar")

	var buf bytes.Buffer

	err = RunArchive(&buf, []string{"verbose.txt"}, ArchiveOptions{
		Create:    true,
		File:      archivePath,
		Directory: tmpDir,
		Verbose:   true,
	})
	if err != nil {
		t.Fatalf("RunArchive() verbose error = %v", err)
	}

	if !strings.Contains(buf.String(), "verbose.txt") {
		t.Error("RunArchive() verbose should print file names")
	}
}

func TestRunArchive_NoFile(t *testing.T) {
	var buf bytes.Buffer

	err := RunArchive(&buf, []string{}, ArchiveOptions{Create: true})
	if err == nil {
		t.Error("RunArchive() should error when no file specified")
	}
}

func TestRunTar(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tar_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "tar.txt")
	_ = os.WriteFile(testFile, []byte("tar content"), 0644)

	archivePath := filepath.Join(tmpDir, "output.tar")

	var buf bytes.Buffer

	err = RunTar(&buf, []string{"tar.txt"}, ArchiveOptions{
		Create:    true,
		File:      archivePath,
		Directory: tmpDir,
	})
	if err != nil {
		t.Fatalf("RunTar() error = %v", err)
	}
}

func TestRunZip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zip_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "zip.txt")
	_ = os.WriteFile(testFile, []byte("zip content"), 0644)

	archivePath := filepath.Join(tmpDir, "output.zip")

	var buf bytes.Buffer

	err = RunZip(&buf, []string{archivePath, "zip.txt"}, ArchiveOptions{
		Directory: tmpDir,
	})
	if err != nil {
		t.Fatalf("RunZip() error = %v", err)
	}
}

func TestRunUnzip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "unzip_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create zip first
	testFile := filepath.Join(tmpDir, "unzip.txt")
	_ = os.WriteFile(testFile, []byte("unzip content"), 0644)

	archivePath := filepath.Join(tmpDir, "test.zip")

	var buf bytes.Buffer

	_ = RunZip(&buf, []string{archivePath, "unzip.txt"}, ArchiveOptions{
		Directory: tmpDir,
	})

	// Remove original
	_ = os.Remove(testFile)

	// Extract
	err = RunUnzip(&buf, []string{archivePath}, ArchiveOptions{
		Directory: tmpDir,
	})
	if err != nil {
		t.Fatalf("RunUnzip() error = %v", err)
	}
}

// writeTarGz builds a .tar.gz archive at path from the supplied headers and
// (for regular files) their contents. data may be nil for non-regular entries.
func writeTarGz(t *testing.T, path string, headers []*tar.Header, data [][]byte) {
	t.Helper()

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)

	for i, h := range headers {
		if err := tw.WriteHeader(h); err != nil {
			t.Fatal(err)
		}
		if h.Typeflag == tar.TypeReg && i < len(data) && data[i] != nil {
			if _, err := tw.Write(data[i]); err != nil {
				t.Fatal(err)
			}
		}
	}

	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestExtractTar_RejectsPathTraversal(t *testing.T) {
	tmpDir := t.TempDir()

	archivePath := filepath.Join(tmpDir, "evil.tar.gz")
	writeTarGz(t, archivePath, []*tar.Header{
		{Name: "../../../../etc/cron.d/evil", Mode: 0644, Size: 4, Typeflag: tar.TypeReg},
	}, [][]byte{[]byte("boom")})

	extractDir := filepath.Join(tmpDir, "out")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err := RunArchive(&buf, []string{}, ArchiveOptions{
		Extract:   true,
		File:      archivePath,
		Directory: extractDir,
	})
	if err == nil {
		t.Fatal("expected path-traversal entry to be rejected")
	}
	if !errors.Is(err, cmderr.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestExtractTar_RejectsAbsolutePath(t *testing.T) {
	tmpDir := t.TempDir()

	archivePath := filepath.Join(tmpDir, "abs.tar.gz")
	absName := "/abs-escape"
	if os.PathSeparator == '\\' {
		absName = "C:/abs-escape"
	}
	writeTarGz(t, archivePath, []*tar.Header{
		{Name: absName, Mode: 0644, Size: 2, Typeflag: tar.TypeReg},
	}, [][]byte{[]byte("hi")})

	var buf bytes.Buffer
	err := RunArchive(&buf, []string{}, ArchiveOptions{
		Extract:   true,
		File:      archivePath,
		Directory: filepath.Join(tmpDir, "out"),
	})
	if err == nil {
		t.Fatal("expected absolute entry name to be rejected")
	}
	if !errors.Is(err, cmderr.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestExtractTar_RejectsEscapingSymlink(t *testing.T) {
	tmpDir := t.TempDir()

	archivePath := filepath.Join(tmpDir, "symlink.tar.gz")
	writeTarGz(t, archivePath, []*tar.Header{
		{Name: "link", Linkname: "../../../../etc", Typeflag: tar.TypeSymlink, Mode: 0777},
	}, nil)

	var buf bytes.Buffer
	err := RunArchive(&buf, []string{}, ArchiveOptions{
		Extract:   true,
		File:      archivePath,
		Directory: filepath.Join(tmpDir, "out"),
	})
	if err == nil {
		t.Fatal("expected escaping symlink to be rejected")
	}
	if !errors.Is(err, cmderr.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestExtractTar_RejectsEscapingHardlink(t *testing.T) {
	tmpDir := t.TempDir()

	archivePath := filepath.Join(tmpDir, "hardlink.tar.gz")
	writeTarGz(t, archivePath, []*tar.Header{
		{Name: "hl", Linkname: "../../../../etc/passwd", Typeflag: tar.TypeLink, Mode: 0644},
	}, nil)

	var buf bytes.Buffer
	err := RunArchive(&buf, []string{}, ArchiveOptions{
		Extract:   true,
		File:      archivePath,
		Directory: filepath.Join(tmpDir, "out"),
	})
	if err == nil {
		t.Fatal("expected escaping hardlink to be rejected")
	}
	if !errors.Is(err, cmderr.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestExtractTar_AllowsContainedRelativeSymlink(t *testing.T) {
	tmpDir := t.TempDir()

	archivePath := filepath.Join(tmpDir, "good.tar.gz")
	writeTarGz(t, archivePath, []*tar.Header{
		{Name: "real.txt", Mode: 0644, Size: 5, Typeflag: tar.TypeReg},
		{Name: "alias.txt", Linkname: "real.txt", Typeflag: tar.TypeSymlink, Mode: 0777},
	}, [][]byte{[]byte("hello"), nil})

	extractDir := filepath.Join(tmpDir, "out")

	var buf bytes.Buffer
	err := RunArchive(&buf, []string{}, ArchiveOptions{
		Extract:   true,
		File:      archivePath,
		Directory: extractDir,
	})
	if err != nil {
		// os.Symlink requires privilege/developer-mode on Windows; skip when
		// the OS refuses to create the link (our containment check already
		// passed, the failure is from os.Symlink itself).
		if errors.Is(err, os.ErrPermission) || strings.Contains(err.Error(), "symlink") {
			t.Skipf("symlink creation unsupported in this environment: %v", err)
		}
		t.Fatalf("contained relative symlink should extract cleanly: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(extractDir, "alias.txt")); err != nil {
		t.Errorf("expected contained symlink to be created: %v", err)
	}
}

// writeZip builds a .zip archive at path with a single entry whose stored
// (uncompressed) content is body. The Store method keeps the on-disk archive
// small relative to the declared uncompressed size when body is highly
// compressible, but here we use it to deterministically control the number of
// bytes extractZipArchive will write.
func writeZip(t *testing.T, path, entryName string, body []byte) {
	t.Helper()

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	zw := zip.NewWriter(f)
	hdr := &zip.FileHeader{Name: entryName, Method: zip.Deflate}
	hdr.SetMode(0644)

	wr, err := zw.CreateHeader(hdr)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := wr.Write(body); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestExtractZip_RejectsDecompressionBomb(t *testing.T) {
	// Temporarily shrink the extraction cap so the test stays fast and does
	// not need a multi-GiB archive. The production default remains
	// maxExtractTotalBytes (10 GiB).
	orig := extractByteCap
	extractByteCap = 16
	defer func() { extractByteCap = orig }()

	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "bomb.zip")

	// A single entry whose extracted size (64 bytes) exceeds the lowered cap.
	writeZip(t, archivePath, "big.txt", bytes.Repeat([]byte("A"), 64))

	extractDir := filepath.Join(tmpDir, "out")

	var buf bytes.Buffer
	err := RunArchive(&buf, []string{}, ArchiveOptions{
		Extract:   true,
		File:      archivePath,
		Directory: extractDir,
	})
	if err == nil {
		t.Fatal("expected decompression-bomb zip to be rejected")
	}
	if !errors.Is(err, cmderr.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
	if !strings.Contains(err.Error(), "exceeds maximum allowed size") {
		t.Errorf("expected size-cap message, got %v", err)
	}
}

func TestExtractZip_AllowsUnderCap(t *testing.T) {
	// An archive whose extracted size is within the cap must extract cleanly.
	orig := extractByteCap
	extractByteCap = 1 << 20 // 1 MiB
	defer func() { extractByteCap = orig }()

	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "ok.zip")
	writeZip(t, archivePath, "ok.txt", []byte("small body"))

	extractDir := filepath.Join(tmpDir, "out")

	var buf bytes.Buffer
	err := RunArchive(&buf, []string{}, ArchiveOptions{
		Extract:   true,
		File:      archivePath,
		Directory: extractDir,
	})
	if err != nil {
		t.Fatalf("under-cap zip should extract cleanly: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(extractDir, "ok.txt"))
	if err != nil {
		t.Fatalf("expected extracted file: %v", err)
	}
	if string(got) != "small body" {
		t.Errorf("extracted content = %q, want %q", got, "small body")
	}
}

// TestExtractTar_RefusesWriteThroughPreplantedSymlink defends against CWE-59
// write-through escape: a symlink that already exists on disk inside a reused
// destDir is not visible to the lexical secureJoin containment check, so an
// entry named "<symlink>/file" would otherwise be written through the symlink
// to a location outside destDir. The extractor must refuse before writing.
func TestExtractTar_RefusesWriteThroughPreplantedSymlink(t *testing.T) {
	if os.PathSeparator == '\\' {
		t.Skip("symlink-through escape test requires POSIX symlink semantics")
	}

	tmpDir := t.TempDir()

	// The attacker-controlled escape destination, outside extractDir.
	escapeDir := filepath.Join(tmpDir, "escape")
	if err := os.MkdirAll(escapeDir, 0755); err != nil {
		t.Fatal(err)
	}

	extractDir := filepath.Join(tmpDir, "out")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Pre-plant a symlink inside the (reused) extraction dir that points out
	// of the destination. A lexical containment check on "sub/evil.txt" passes
	// because the literal path stays under extractDir.
	preplanted := filepath.Join(extractDir, "sub")
	if err := os.Symlink(escapeDir, preplanted); err != nil {
		if errors.Is(err, os.ErrPermission) {
			t.Skipf("symlink creation unsupported in this environment: %v", err)
		}
		t.Fatal(err)
	}

	archivePath := filepath.Join(tmpDir, "evil.tar.gz")
	writeTarGz(t, archivePath, []*tar.Header{
		{Name: "sub/evil.txt", Mode: 0644, Size: 4, Typeflag: tar.TypeReg},
	}, [][]byte{[]byte("boom")})

	var buf bytes.Buffer
	err := RunArchive(&buf, []string{}, ArchiveOptions{
		Extract:   true,
		File:      archivePath,
		Directory: extractDir,
	})
	if err == nil {
		t.Fatal("expected write-through pre-planted symlink to be refused")
	}
	if !errors.Is(err, cmderr.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}

	// The escape file must NOT have been written through the symlink.
	if _, statErr := os.Stat(filepath.Join(escapeDir, "evil.txt")); statErr == nil {
		t.Error("file was written through the pre-planted symlink (escape succeeded)")
	}
}

// TestExtractZip_RefusesWriteThroughPreplantedSymlink mirrors the tar case for
// the zip extractor.
func TestExtractZip_RefusesWriteThroughPreplantedSymlink(t *testing.T) {
	if os.PathSeparator == '\\' {
		t.Skip("symlink-through escape test requires POSIX symlink semantics")
	}

	tmpDir := t.TempDir()

	escapeDir := filepath.Join(tmpDir, "escape")
	if err := os.MkdirAll(escapeDir, 0755); err != nil {
		t.Fatal(err)
	}

	extractDir := filepath.Join(tmpDir, "out")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		t.Fatal(err)
	}

	preplanted := filepath.Join(extractDir, "sub")
	if err := os.Symlink(escapeDir, preplanted); err != nil {
		if errors.Is(err, os.ErrPermission) {
			t.Skipf("symlink creation unsupported in this environment: %v", err)
		}
		t.Fatal(err)
	}

	archivePath := filepath.Join(tmpDir, "evil.zip")
	writeZip(t, archivePath, "sub/evil.txt", []byte("boom"))

	var buf bytes.Buffer
	err := RunArchive(&buf, []string{}, ArchiveOptions{
		Extract:   true,
		File:      archivePath,
		Directory: extractDir,
	})
	if err == nil {
		t.Fatal("expected write-through pre-planted symlink to be refused")
	}
	if !errors.Is(err, cmderr.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}

	if _, statErr := os.Stat(filepath.Join(escapeDir, "evil.txt")); statErr == nil {
		t.Error("file was written through the pre-planted symlink (escape succeeded)")
	}
}

func TestRunArchive_Directory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "archive_dir_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create directory with files
	subDir := filepath.Join(tmpDir, "subdir")
	_ = os.MkdirAll(subDir, 0755)
	_ = os.WriteFile(filepath.Join(subDir, "file.txt"), []byte("in subdir"), 0644)

	archivePath := filepath.Join(tmpDir, "dir.tar")

	var buf bytes.Buffer

	err = RunArchive(&buf, []string{"subdir"}, ArchiveOptions{
		Create:    true,
		File:      archivePath,
		Directory: tmpDir,
	})
	if err != nil {
		t.Fatalf("RunArchive() directory error = %v", err)
	}

	// List to verify
	buf.Reset()
	_ = RunArchive(&buf, []string{}, ArchiveOptions{
		List: true,
		File: archivePath,
	})

	output := buf.String()
	if !strings.Contains(output, "subdir") {
		t.Errorf("RunArchive() should include directory: %s", output)
	}
}
