package tagfixer

import (
	"os"
	"path/filepath"
	"testing"
)

// sampleGoSource is a Go file with structs whose field tags use mixed casing,
// exercising the AST walkers in AnalyzePath, ListStructTags, and processFile.
const sampleGoSource = `package sample

type User struct {
	UserName  string ` + "`json:\"user_name\" yaml:\"userName\"`" + `
	EmailAddr string ` + "`json:\"email_addr\"`" + `
	Untagged  int
}

type Config struct {
	MaxRetries int    ` + "`json:\"maxRetries\" toml:\"max_retries\"`" + `
	Endpoint   string ` + "`json:\"endpoint\"`" + `
}
`

func TestConvertToCase(t *testing.T) {
	tests := []struct {
		name   string
		in     string
		target CaseType
		want   string
	}{
		{"snake to camel", "user_name", CaseCamel, "userName"},
		{"snake to pascal", "user_name", CasePascal, "UserName"},
		{"camel to snake", "userName", CaseSnake, "user_name"},
		{"camel to kebab", "userName", CaseKebab, "user-name"},
		{"pascal to snake", "UserName", CaseSnake, "user_name"},
		{"kebab to camel", "user-name", CaseCamel, "userName"},
		{"single word camel", "name", CaseCamel, "name"},
		{"single word pascal", "name", CasePascal, "Name"},
		{"acronym split snake", "HTTPServer", CaseSnake, "http_server"},
		{"three words kebab", "max_retry_count", CaseKebab, "max-retry-count"},
		{"unknown case passthrough", "user_name", CaseType("bogus"), "user_name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertToCase(tt.in, tt.target); got != tt.want {
				t.Errorf("ConvertToCase(%q, %q) = %q, want %q", tt.in, tt.target, got, tt.want)
			}
		})
	}
}

func writeSampleGoFile(t *testing.T) (dir, file string) {
	t.Helper()

	dir = t.TempDir()
	file = filepath.Join(dir, "sample.go")

	if err := os.WriteFile(file, []byte(sampleGoSource), 0o644); err != nil {
		t.Fatal(err)
	}

	return dir, file
}

func TestAnalyzePath(t *testing.T) {
	dir, file := writeSampleGoFile(t)

	t.Run("single file", func(t *testing.T) {
		res, err := AnalyzePath(file, []string{"json", "yaml", "toml"}, false)
		if err != nil {
			t.Fatalf("AnalyzePath() error = %v", err)
		}

		if res.TotalFiles != 1 {
			t.Errorf("TotalFiles = %d, want 1", res.TotalFiles)
		}

		if res.TotalStructs != 2 {
			t.Errorf("TotalStructs = %d, want 2", res.TotalStructs)
		}

		if _, ok := res.TagStats["json"]; !ok {
			t.Error("expected json tag stats")
		}

		if res.Recommended == "" {
			t.Error("expected a recommended case")
		}
	})

	t.Run("directory", func(t *testing.T) {
		res, err := AnalyzePath(dir, []string{"json"}, false)
		if err != nil {
			t.Fatalf("AnalyzePath() dir error = %v", err)
		}

		if res.TotalFiles != 1 {
			t.Errorf("TotalFiles = %d, want 1", res.TotalFiles)
		}
	})

	t.Run("missing path", func(t *testing.T) {
		if _, err := AnalyzePath(filepath.Join(dir, "nope"), []string{"json"}, false); err == nil {
			t.Error("AnalyzePath() should error on missing path")
		}
	})
}

func TestListStructTags(t *testing.T) {
	_, file := writeSampleGoFile(t)

	t.Run("lists distinct tags sorted", func(t *testing.T) {
		tags, err := ListStructTags(file)
		if err != nil {
			t.Fatalf("ListStructTags() error = %v", err)
		}

		want := map[string]bool{"json": true, "yaml": true, "toml": true}
		got := make(map[string]bool)

		for _, tag := range tags {
			got[tag] = true
		}

		for tag := range want {
			if !got[tag] {
				t.Errorf("ListStructTags() missing tag %q in %v", tag, tags)
			}
		}

		// Output must be sorted.
		for i := 1; i < len(tags); i++ {
			if tags[i-1] > tags[i] {
				t.Errorf("ListStructTags() not sorted: %v", tags)
			}
		}
	})

	t.Run("parse error", func(t *testing.T) {
		dir := t.TempDir()
		bad := filepath.Join(dir, "bad.go")
		if err := os.WriteFile(bad, []byte("this is not go"), 0o644); err != nil {
			t.Fatal(err)
		}

		if _, err := ListStructTags(bad); err == nil {
			t.Error("ListStructTags() should error on invalid Go source")
		}
	})
}

func TestCollectGoFilesMore(t *testing.T) {
	dir := t.TempDir()

	// Layout: a.go, a_test.go (skipped), sub/b.go, vendor/c.go (skipped).
	mustWrite := func(rel, content string) {
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	mustWrite("a.go", "package a")
	mustWrite("a_test.go", "package a")
	mustWrite("sub/b.go", "package b")
	mustWrite("vendor/c.go", "package c")

	t.Run("non-recursive skips subdirs", func(t *testing.T) {
		files, err := collectGoFiles(dir, false)
		if err != nil {
			t.Fatalf("collectGoFiles() error = %v", err)
		}

		for _, f := range files {
			if filepath.Base(f) != "a.go" {
				t.Errorf("non-recursive should only include a.go, got %s", f)
			}
		}

		if len(files) != 1 {
			t.Errorf("non-recursive got %d files, want 1: %v", len(files), files)
		}
	})

	t.Run("recursive includes subdirs but skips vendor and tests", func(t *testing.T) {
		files, err := collectGoFiles(dir, true)
		if err != nil {
			t.Fatalf("collectGoFiles() recursive error = %v", err)
		}

		got := make(map[string]bool)
		for _, f := range files {
			got[filepath.Base(f)] = true
		}

		if !got["a.go"] || !got["b.go"] {
			t.Errorf("recursive should include a.go and b.go, got %v", files)
		}

		if got["a_test.go"] {
			t.Error("recursive must skip _test.go files")
		}

		if got["c.go"] {
			t.Error("recursive must skip vendor dir")
		}
	})

	t.Run("single non-go file returns nil", func(t *testing.T) {
		txt := filepath.Join(dir, "note.txt")
		if err := os.WriteFile(txt, []byte("hi"), 0o644); err != nil {
			t.Fatal(err)
		}

		files, err := collectGoFiles(txt, false)
		if err != nil {
			t.Fatalf("collectGoFiles() txt error = %v", err)
		}

		if files != nil {
			t.Errorf("non-go file should yield nil, got %v", files)
		}
	})

	t.Run("missing path errors", func(t *testing.T) {
		if _, err := collectGoFiles(filepath.Join(dir, "nope"), false); err == nil {
			t.Error("collectGoFiles() should error on missing path")
		}
	})
}

func TestProcessFile(t *testing.T) {
	_, file := writeSampleGoFile(t)

	t.Run("dry run reports changes", func(t *testing.T) {
		res := processFile(file, Options{
			Case:   CaseCamel,
			Tags:   []string{"json"},
			DryRun: true,
		})

		if res.Error != "" {
			t.Fatalf("processFile() error = %s", res.Error)
		}

		// user_name -> userName etc. should produce changes.
		if len(res.Changes) == 0 {
			t.Error("processFile() expected changes for snake_case json tags")
		}
	})

	t.Run("parse error", func(t *testing.T) {
		dir := t.TempDir()
		bad := filepath.Join(dir, "bad.go")
		if err := os.WriteFile(bad, []byte("not valid go !!!"), 0o644); err != nil {
			t.Fatal(err)
		}

		res := processFile(bad, Options{Case: CaseCamel, Tags: []string{"json"}})
		if res.Error == "" {
			t.Error("processFile() should report a parse error")
		}
	})
}
