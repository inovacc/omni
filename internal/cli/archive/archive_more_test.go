package archive

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestRoundTrip_TarGz creates a gzip-compressed tar from a directory tree,
// lists it, extracts it to a fresh dir, and compares contents. This drives
// createTarArchive (gzip branch), listTarArchive, and extractTarArchive (gzip
// branch) end to end.
func TestRoundTrip_TarGz(t *testing.T) {
	srcDir := t.TempDir()

	if err := os.MkdirAll(filepath.Join(srcDir, "tree", "nested"), 0o755); err != nil {
		t.Fatal(err)
	}

	files := map[string]string{
		"tree/a.txt":           "alpha",
		"tree/b.txt":           "bravo",
		"tree/nested/deep.txt": "charlie",
	}
	for rel, content := range files {
		if err := os.WriteFile(filepath.Join(srcDir, rel), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	archivePath := filepath.Join(srcDir, "out.tar.gz")

	var buf bytes.Buffer
	if err := RunArchive(&buf, []string{"tree"}, ArchiveOptions{
		Create:    true,
		File:      archivePath,
		Directory: srcDir,
		Gzip:      true,
	}); err != nil {
		t.Fatalf("create tar.gz error = %v", err)
	}

	// List (plain).
	buf.Reset()
	if err := RunArchive(&buf, nil, ArchiveOptions{List: true, File: archivePath}); err != nil {
		t.Fatalf("list tar.gz error = %v", err)
	}

	// Tar entry names are stored using the host path separator by
	// createTarArchive (it writes filepath.Rel verbatim), so normalize both
	// sides to forward slashes before comparing for cross-platform stability.
	listOut := filepath.ToSlash(buf.String())
	for rel := range files {
		if !strings.Contains(listOut, rel) {
			t.Errorf("list output missing %q:\n%s", rel, listOut)
		}
	}

	// Extract to a different dir and compare.
	dstDir := t.TempDir()
	buf.Reset()
	if err := RunArchive(&buf, nil, ArchiveOptions{
		Extract:   true,
		File:      archivePath,
		Directory: dstDir,
	}); err != nil {
		t.Fatalf("extract tar.gz error = %v", err)
	}

	for rel, want := range files {
		got, err := os.ReadFile(filepath.Join(dstDir, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatalf("extracted file %q missing: %v", rel, err)
		}

		if string(got) != want {
			t.Errorf("extracted %q = %q, want %q", rel, got, want)
		}
	}
}

// TestRoundTrip_Zip mirrors the tar.gz round-trip for the zip path.
func TestRoundTrip_Zip(t *testing.T) {
	srcDir := t.TempDir()

	if err := os.MkdirAll(filepath.Join(srcDir, "data"), 0o755); err != nil {
		t.Fatal(err)
	}

	files := map[string]string{
		"data/one.txt": "111",
		"data/two.txt": "222",
	}
	for rel, content := range files {
		if err := os.WriteFile(filepath.Join(srcDir, rel), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	archivePath := filepath.Join(srcDir, "out.zip")

	var buf bytes.Buffer
	if err := RunArchive(&buf, []string{"data"}, ArchiveOptions{
		Create:    true,
		File:      archivePath,
		Directory: srcDir,
	}); err != nil {
		t.Fatalf("create zip error = %v", err)
	}

	dstDir := t.TempDir()
	buf.Reset()
	if err := RunArchive(&buf, nil, ArchiveOptions{
		Extract:   true,
		File:      archivePath,
		Directory: dstDir,
	}); err != nil {
		t.Fatalf("extract zip error = %v", err)
	}

	for rel, want := range files {
		got, err := os.ReadFile(filepath.Join(dstDir, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatalf("extracted file %q missing: %v", rel, err)
		}

		if string(got) != want {
			t.Errorf("extracted %q = %q, want %q", rel, got, want)
		}
	}
}

// TestListTar_Verbose exercises the verbose (long-listing) branch of
// listTarArchive.
func TestListTar_Verbose(t *testing.T) {
	srcDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(srcDir, "v.txt"), []byte("verbose body"), 0o644); err != nil {
		t.Fatal(err)
	}

	archivePath := filepath.Join(srcDir, "v.tar")

	var buf bytes.Buffer
	if err := RunArchive(&buf, []string{"v.txt"}, ArchiveOptions{
		Create:    true,
		File:      archivePath,
		Directory: srcDir,
	}); err != nil {
		t.Fatalf("create error = %v", err)
	}

	buf.Reset()
	if err := RunArchive(&buf, nil, ArchiveOptions{
		List:    true,
		File:    archivePath,
		Verbose: true,
	}); err != nil {
		t.Fatalf("list verbose error = %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "v.txt") {
		t.Errorf("verbose list missing file name: %q", out)
	}

	// Verbose listing prints the size column; "verbose body" is 12 bytes.
	if !strings.Contains(out, "12") {
		t.Errorf("verbose list missing size column: %q", out)
	}
}

// TestListTarGz_JSON drives the JSON branch of listTarArchive over a gzip
// archive, validating the structured output shape.
func TestListTarGz_JSON(t *testing.T) {
	srcDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(srcDir, "j.txt"), []byte("json body"), 0o644); err != nil {
		t.Fatal(err)
	}

	archivePath := filepath.Join(srcDir, "j.tar.gz")

	var buf bytes.Buffer
	if err := RunArchive(&buf, []string{"j.txt"}, ArchiveOptions{
		Create:    true,
		File:      archivePath,
		Directory: srcDir,
		Gzip:      true,
	}); err != nil {
		t.Fatalf("create error = %v", err)
	}

	buf.Reset()
	if err := RunArchive(&buf, nil, ArchiveOptions{
		List: true,
		File: archivePath,
		JSON: true,
	}); err != nil {
		t.Fatalf("list json error = %v", err)
	}

	var res ArchiveListResult
	if err := json.Unmarshal(buf.Bytes(), &res); err != nil {
		t.Fatalf("list json invalid: %v\n%s", err, buf.String())
	}

	if res.Count == 0 || len(res.Entries) == 0 {
		t.Errorf("expected entries in JSON listing, got %+v", res)
	}

	found := false
	for _, e := range res.Entries {
		if e.Name == "j.txt" {
			found = true

			if e.Type != "file" {
				t.Errorf("entry type = %q, want file", e.Type)
			}
		}
	}

	if !found {
		t.Errorf("j.txt not present in JSON entries: %+v", res.Entries)
	}
}

// TestListZip_Verbose exercises the verbose branch of listZipArchive.
func TestListZip_Verbose(t *testing.T) {
	srcDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(srcDir, "z.txt"), []byte("zip verbose"), 0o644); err != nil {
		t.Fatal(err)
	}

	archivePath := filepath.Join(srcDir, "z.zip")

	var buf bytes.Buffer
	if err := RunArchive(&buf, []string{"z.txt"}, ArchiveOptions{
		Create:    true,
		File:      archivePath,
		Directory: srcDir,
	}); err != nil {
		t.Fatalf("create error = %v", err)
	}

	buf.Reset()
	if err := RunArchive(&buf, nil, ArchiveOptions{
		List:    true,
		File:    archivePath,
		Verbose: true,
	}); err != nil {
		t.Fatalf("list verbose zip error = %v", err)
	}

	if !strings.Contains(buf.String(), "z.txt") {
		t.Errorf("verbose zip list missing file: %q", buf.String())
	}
}

// TestListZip_JSON drives the JSON branch of listZipArchive.
func TestListZip_JSON(t *testing.T) {
	srcDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(srcDir, "jz.txt"), []byte("jz"), 0o644); err != nil {
		t.Fatal(err)
	}

	archivePath := filepath.Join(srcDir, "jz.zip")

	var buf bytes.Buffer
	if err := RunArchive(&buf, []string{"jz.txt"}, ArchiveOptions{
		Create:    true,
		File:      archivePath,
		Directory: srcDir,
	}); err != nil {
		t.Fatalf("create error = %v", err)
	}

	buf.Reset()
	if err := RunArchive(&buf, nil, ArchiveOptions{List: true, File: archivePath, JSON: true}); err != nil {
		t.Fatalf("list json zip error = %v", err)
	}

	var res ArchiveListResult
	if err := json.Unmarshal(buf.Bytes(), &res); err != nil {
		t.Fatalf("zip list json invalid: %v", err)
	}

	if res.Count == 0 {
		t.Error("expected non-zero count in zip JSON listing")
	}
}

// TestExtractTar_StripComponents drives the strip-components branch of
// extractTarArchive. The strip logic splits entry names on "/", and
// createTarArchive stores names using the host separator, so on Windows the
// single-segment-on-"/" names mean the branch is exercised but no file lands.
// We build the archive with explicit forward-slash entry names to keep the
// behavior deterministic across platforms.
func TestExtractTar_StripComponents(t *testing.T) {
	srcDir := t.TempDir()
	archivePath := filepath.Join(srcDir, "strip.tar.gz")

	// Use the package test helper from archive_test.go to write known names.
	writeTarGz(t, archivePath, []*tar.Header{
		{Name: "top/inner.txt", Mode: 0o644, Size: 5, Typeflag: tar.TypeReg},
		{Name: "only-one", Mode: 0o644, Size: 1, Typeflag: tar.TypeReg},
	}, [][]byte{[]byte("inner"), []byte("x")})

	dstDir := t.TempDir()

	var buf bytes.Buffer
	if err := RunArchive(&buf, nil, ArchiveOptions{
		Extract:         true,
		File:            archivePath,
		Directory:       dstDir,
		StripComponents: 1,
	}); err != nil {
		t.Fatalf("extract strip error = %v", err)
	}

	// "top/inner.txt" with one component stripped becomes "inner.txt".
	// Retry the stat briefly: on Windows a freshly-extracted file can be
	// momentarily invisible to Stat under heavy concurrent test load (AV/indexer
	// holding the handle), which is environmental, not a logic error.
	var statErr error
	for i := 0; i < 20; i++ {
		if _, statErr = os.Stat(filepath.Join(dstDir, "inner.txt")); statErr == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if statErr != nil {
		t.Errorf("expected stripped file inner.txt: %v", statErr)
	}

	// "only-one" has fewer than StripComponents parts and must be skipped.
	if _, err := os.Stat(filepath.Join(dstDir, "only-one")); err == nil {
		t.Error("expected single-segment entry to be skipped under strip-components")
	}
}

// TestRefuseWriteThroughSymlink_NoSymlink verifies the happy path: when no path
// component is an existing symlink the function returns nil. This is portable
// (no symlink creation required) and covers the main loop + early-return cases.
func TestRefuseWriteThroughSymlink_NoSymlink(t *testing.T) {
	dest := t.TempDir()

	t.Run("nonexistent nested target", func(t *testing.T) {
		target := filepath.Join(dest, "a", "b", "c.txt")
		if err := refuseWriteThroughSymlink(dest, target); err != nil {
			t.Errorf("expected nil for nonexistent nested target, got %v", err)
		}
	})

	t.Run("existing regular dirs", func(t *testing.T) {
		if err := os.MkdirAll(filepath.Join(dest, "real", "sub"), 0o755); err != nil {
			t.Fatal(err)
		}

		target := filepath.Join(dest, "real", "sub", "file.txt")
		if err := refuseWriteThroughSymlink(dest, target); err != nil {
			t.Errorf("expected nil for regular dir chain, got %v", err)
		}
	})

	t.Run("target equals dest", func(t *testing.T) {
		if err := refuseWriteThroughSymlink(dest, dest); err != nil {
			t.Errorf("expected nil when target == dest, got %v", err)
		}
	})
}

// TestRefuseWriteThroughSymlink_RejectsSymlink pre-plants a symlink directory
// segment and verifies the function rejects writing through it. Skipped where
// the OS cannot create symlinks (e.g. Windows without privilege).
func TestRefuseWriteThroughSymlink_RejectsSymlink(t *testing.T) {
	dest := t.TempDir()
	outside := t.TempDir()

	link := filepath.Join(dest, "sub")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlink creation unsupported in this environment: %v", err)
	}

	target := filepath.Join(dest, "sub", "evil.txt")
	if err := refuseWriteThroughSymlink(dest, target); err == nil {
		t.Error("expected rejection when a path segment is a symlink")
	}
}
