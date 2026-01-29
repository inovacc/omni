package archive

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
