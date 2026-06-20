package loc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeLocTree(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		full := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestRunLoc_TableOutput(t *testing.T) {
	dir := writeLocTree(t, map[string]string{
		"main.go": "package main\n\n// a comment\nfunc main() {\n\tprintln(\"hi\")\n}\n",
		"util.go": "package main\n\nfunc util() int {\n\treturn 1\n}\n",
		"app.py":  "# comment\nimport sys\n\ndef f():\n    return 2\n",
		"r.md":    "# Title\n\nSome text.\n",
	})

	var buf strings.Builder
	if err := RunLoc(&buf, []string{dir}, Options{}); err != nil {
		t.Fatalf("RunLoc err=%v", err)
	}
	out := buf.String()
	for _, want := range []string{"Language", "Go", "Total", "Files"} {
		if !strings.Contains(out, want) {
			t.Errorf("loc table output missing %q:\n%s", want, out)
		}
	}
}

func TestRunLoc_EmbeddedLanguages(t *testing.T) {
	// An HTML file with embedded <script> and <style> triggers the children/
	// embedded-language branch of printTable.
	html := `<!DOCTYPE html>
<html>
<head>
<style>
body { color: red; }
.x { margin: 0; }
</style>
</head>
<body>
<h1>Hi</h1>
<script>
function greet() {
  console.log("hello");
}
greet();
</script>
</body>
</html>
`
	dir := writeLocTree(t, map[string]string{"index.html": html})

	var buf strings.Builder
	if err := RunLoc(&buf, []string{dir}, Options{}); err != nil {
		t.Fatalf("RunLoc html err=%v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "HTML") {
		t.Errorf("expected HTML language in output:\n%s", out)
	}
	// Children rows are prefixed with "|-"; total subtotal line says "(Total)".
	// These appear only when embedded languages were detected.
	if strings.Contains(out, "|-") && !strings.Contains(out, "(Total)") {
		t.Errorf("embedded children present but no subtotal line:\n%s", out)
	}
}

func TestRunLoc_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	var buf strings.Builder
	if err := RunLoc(&buf, []string{dir}, Options{}); err != nil {
		t.Fatalf("RunLoc empty err=%v", err)
	}
	if !strings.Contains(buf.String(), "No files found") {
		t.Errorf("expected 'No files found' for empty dir, got: %q", buf.String())
	}
}

// TestPrintTableDirect drives printTable with an explicit Result so the
// children/subtotal/total branches are exercised deterministically.
func TestPrintTableDirect(t *testing.T) {
	res := Result{
		Languages: []LanguageStats{
			{
				Language: "Markdown", Files: 1, Lines: 20, Code: 10, Comments: 5, Blanks: 5,
				Children: map[string]*LanguageStats{
					"Go":   {Language: "Go", Lines: 8, Code: 6, Comments: 1, Blanks: 1},
					"Bash": {Language: "Bash", Lines: 4, Code: 3, Comments: 0, Blanks: 1},
				},
			},
			{Language: "Go", Files: 2, Lines: 100, Code: 80, Comments: 10, Blanks: 10},
		},
		Total: LanguageStats{Files: 3, Lines: 120, Code: 90, Comments: 15, Blanks: 15},
	}

	var buf strings.Builder
	if err := printTable(&buf, res); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"Markdown", "|-", "(Total)", "Total"} {
		if !strings.Contains(out, want) {
			t.Errorf("printTable output missing %q:\n%s", want, out)
		}
	}
}

// TestPrintTableEmpty covers the no-files early return.
func TestPrintTableEmpty(t *testing.T) {
	var buf strings.Builder
	if err := printTable(&buf, Result{}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "No files found.") {
		t.Errorf("expected empty message, got %q", buf.String())
	}
}

// TestTruncateLoc covers loc's truncate helper.
func TestTruncateLoc(t *testing.T) {
	if got := truncate("short", 17); got != "short" {
		t.Errorf("short = %q", got)
	}
	if got := truncate("averyverylongname", 8); got != "averyve." {
		t.Errorf("long = %q", got)
	}
}
