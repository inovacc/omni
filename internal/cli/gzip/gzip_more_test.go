package gzip

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestGzipRoundTripFiles compresses a real file, decompresses it back, and zcats
// it, verifying the bytes survive the round trip.
func TestGzipRoundTripFiles(t *testing.T) {
	dir := t.TempDir()
	content := []byte("the quick brown fox jumps over the lazy dog\n")

	src := filepath.Join(dir, "input.txt")
	if err := os.WriteFile(src, content, 0o644); err != nil {
		t.Fatal(err)
	}

	// Compress with -k (keep original) so we can compare both.
	var out bytes.Buffer
	if err := RunGzip(&out, []string{src}, GzipOptions{Keep: true, Verbose: true}); err != nil {
		t.Fatalf("gzip: %v", err)
	}

	gzPath := src + ".gz"
	if _, err := os.Stat(gzPath); err != nil {
		t.Fatalf("expected %s to exist: %v", gzPath, err)
	}

	// zcat the .gz back and confirm it matches.
	var zout bytes.Buffer
	if err := RunZcat(&zout, []string{gzPath}); err != nil {
		t.Fatalf("zcat: %v", err)
	}
	if !bytes.Equal(zout.Bytes(), content) {
		t.Errorf("zcat output = %q, want %q", zout.String(), content)
	}

	// Now remove the original and gunzip the .gz.
	if err := os.Remove(src); err != nil {
		t.Fatal(err)
	}
	var gout bytes.Buffer
	if err := RunGunzip(&gout, []string{gzPath}, GzipOptions{Keep: true, Verbose: true}); err != nil {
		t.Fatalf("gunzip: %v", err)
	}
	got, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("read decompressed: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("gunzip output = %q, want %q", got, content)
	}
}

// TestGzipStdout exercises the -c (stdout) path for both compress and decompress.
func TestGzipStdout(t *testing.T) {
	dir := t.TempDir()
	content := []byte("stdout streaming data\n")
	src := filepath.Join(dir, "s.txt")
	if err := os.WriteFile(src, content, 0o644); err != nil {
		t.Fatal(err)
	}

	var comp bytes.Buffer
	if err := RunGzip(&comp, []string{src}, GzipOptions{Stdout: true, Keep: true}); err != nil {
		t.Fatalf("gzip -c: %v", err)
	}
	if comp.Len() == 0 {
		t.Fatal("expected compressed output on stdout")
	}

	// Decompress the compressed bytes directly through gunzipReader.
	var decomp bytes.Buffer
	if err := gunzipReader(&decomp, &comp); err != nil {
		t.Fatalf("gunzipReader: %v", err)
	}
	if !bytes.Equal(decomp.Bytes(), content) {
		t.Errorf("round-trip stdout = %q, want %q", decomp.String(), content)
	}
}

// TestGzipReaderRoundTrip directly tests gzipReader+gunzipReader at varied levels.
func TestGzipReaderRoundTrip(t *testing.T) {
	content := bytes.Repeat([]byte("abcd"), 1000)
	for _, level := range []int{1, 5, 9} {
		var comp bytes.Buffer
		if err := gzipReader(&comp, bytes.NewReader(content), level); err != nil {
			t.Fatalf("level %d gzipReader: %v", level, err)
		}
		var decomp bytes.Buffer
		if err := gunzipReader(&decomp, &comp); err != nil {
			t.Fatalf("level %d gunzipReader: %v", level, err)
		}
		if !bytes.Equal(decomp.Bytes(), content) {
			t.Errorf("level %d round trip mismatch", level)
		}
	}
}

// TestGunzipBadSuffix verifies gunzip rejects files without a .gz suffix.
func TestGunzipBadSuffix(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "plain.txt")
	if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	// gunzipFile writes the error to stderr via RunGzip and returns nil overall,
	// so call gunzipFile directly to assert the suffix check.
	if err := gunzipFile(&bytes.Buffer{}, p, GzipOptions{}); err == nil {
		t.Error("expected error for non-.gz file")
	}
}

// TestGzipAlreadyCompressed verifies gzipFile rejects a .gz input.
func TestGzipAlreadyCompressed(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "a.gz")
	if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := gzipFile(&bytes.Buffer{}, p, GzipOptions{}); err == nil {
		t.Error("expected error for already-compressed file")
	}
}

// TestZcatPlainFile confirms zcat falls back to copying a non-gz file.
func TestZcatPlainFile(t *testing.T) {
	dir := t.TempDir()
	content := []byte("not compressed\n")
	p := filepath.Join(dir, "plain")
	if err := os.WriteFile(p, content, 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := RunZcat(&out, []string{p}); err != nil {
		t.Fatalf("zcat plain: %v", err)
	}
	if !bytes.Equal(out.Bytes(), content) {
		t.Errorf("zcat plain = %q, want %q", out.String(), content)
	}
}

// TestGzipForceOverwrite covers the -f overwrite branch of gzipFile when the
// .gz target already exists.
func TestGzipForceOverwrite(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "doc.txt")
	if err := os.WriteFile(src, []byte("payload\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Pre-create the .gz target so the conflict/force branch is taken.
	if err := os.WriteFile(src+".gz", []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Without -f, gzipFile reports a conflict error.
	if err := gzipFile(&bytes.Buffer{}, src, GzipOptions{Keep: true}); err == nil {
		t.Error("expected conflict error without -f")
	}

	// With -f it overwrites and produces a valid gzip.
	var out bytes.Buffer
	if err := gzipFile(&out, src, GzipOptions{Keep: true, Force: true, Verbose: true}); err != nil {
		t.Fatalf("force overwrite: %v", err)
	}
	var decomp bytes.Buffer
	f, err := os.Open(src + ".gz")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	if err := gunzipReader(&decomp, f); err != nil {
		t.Fatalf("gunzip overwritten: %v", err)
	}
	if decomp.String() != "payload\n" {
		t.Errorf("overwritten content = %q", decomp.String())
	}
}

// TestGunzipForceOverwrite covers the -f overwrite branch of gunzipFile.
func TestGunzipForceOverwrite(t *testing.T) {
	dir := t.TempDir()
	// Build a real .gz file.
	gzPath := filepath.Join(dir, "data.txt.gz")
	var comp bytes.Buffer
	if err := gzipReader(&comp, bytes.NewReader([]byte("decompressed\n")), 6); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(gzPath, comp.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	// Pre-create the output target.
	outPath := filepath.Join(dir, "data.txt")
	if err := os.WriteFile(outPath, []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := gunzipFile(&bytes.Buffer{}, gzPath, GzipOptions{Keep: true}); err == nil {
		t.Error("expected conflict error without -f")
	}
	if err := gunzipFile(&bytes.Buffer{}, gzPath, GzipOptions{Keep: true, Force: true, Verbose: true}); err != nil {
		t.Fatalf("force gunzip: %v", err)
	}
	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "decompressed\n" {
		t.Errorf("force gunzip content = %q", got)
	}
}

// TestDecompressBombGuard lowers the cap to exercise the CWE-409 guard.
func TestDecompressBombGuard(t *testing.T) {
	orig := decompressByteCap
	decompressByteCap = 8
	defer func() { decompressByteCap = orig }()

	var comp bytes.Buffer
	if err := gzipReader(&comp, bytes.NewReader(bytes.Repeat([]byte("z"), 1000)), 9); err != nil {
		t.Fatal(err)
	}
	if err := gunzipReader(&bytes.Buffer{}, &comp); err == nil {
		t.Error("expected decompression-bomb error")
	}
}
