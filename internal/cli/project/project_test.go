package project

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolvePath(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"empty args uses cwd", nil, false},
		{"dot uses cwd", []string{"."}, false},
		{"nonexistent dir", []string{"/nonexistent/path/xyz"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := resolvePath(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolvePath(%v) error = %v, wantErr %v", tt.args, err, tt.wantErr)
			}
		})
	}
}

func TestDetectProjectTypes(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		wantLang string
	}{
		{"go project", []string{"go.mod"}, "Go"},
		{"node project", []string{"package.json"}, "JavaScript/TypeScript"},
		{"rust project", []string{"Cargo.toml"}, "Rust"},
		{"python requirements", []string{"requirements.txt"}, "Python"},
		{"python pyproject", []string{"pyproject.toml"}, "Python"},
		{"java maven", []string{"pom.xml"}, "Java"},
		{"java gradle", []string{"build.gradle"}, "Java"},
		{"ruby project", []string{"Gemfile"}, "Ruby"},
		{"php project", []string{"composer.json"}, "PHP"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, f := range tt.files {
				if err := os.WriteFile(filepath.Join(dir, f), []byte(""), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			types := detectProjectTypes(dir)
			if len(types) == 0 {
				t.Fatal("expected at least one project type")
			}

			if types[0].Language != tt.wantLang {
				t.Errorf("got language %q, want %q", types[0].Language, tt.wantLang)
			}
		})
	}
}

func TestDetectBuildTools(t *testing.T) {
	dir := t.TempDir()

	// Create some build tool files
	for _, f := range []string{"Taskfile.yml", "Dockerfile"} {
		if err := os.WriteFile(filepath.Join(dir, f), []byte(""), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	tools := detectBuildTools(dir)
	if len(tools) != 2 {
		t.Fatalf("expected 2 build tools, got %d: %v", len(tools), tools)
	}

	found := map[string]bool{}
	for _, tool := range tools {
		found[tool] = true
	}

	if !found["Taskfile"] {
		t.Error("expected Taskfile in build tools")
	}

	if !found["Docker"] {
		t.Error("expected Docker in build tools")
	}
}

func TestCountLanguages(t *testing.T) {
	dir := t.TempDir()

	// Create some source files
	files := map[string]string{
		"main.go":     "package main",
		"lib.go":      "package main",
		"handler.go":  "package main",
		"index.js":    "console.log()",
		"README.md":   "# Hello",
		"config.yaml": "key: value",
	}

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	langs := countLanguages(dir)

	langMap := map[string]int{}
	for _, l := range langs {
		langMap[l.Name] = l.FileCount
	}

	if langMap["Go"] != 3 {
		t.Errorf("expected 3 Go files, got %d", langMap["Go"])
	}

	if langMap["JavaScript"] != 1 {
		t.Errorf("expected 1 JavaScript file, got %d", langMap["JavaScript"])
	}

	if langMap["Markdown"] != 1 {
		t.Errorf("expected 1 Markdown file, got %d", langMap["Markdown"])
	}
}

func TestParseGoMod(t *testing.T) {
	dir := t.TempDir()
	content := `module github.com/example/project

go 1.22

require (
	github.com/spf13/cobra v1.8.0
	github.com/stretchr/testify v1.9.0
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
)
`

	path := filepath.Join(dir, "go.mod")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	deps := parseGoMod(path)
	if deps == nil {
		t.Fatal("expected non-nil deps")
	}

	if deps.Module != "github.com/example/project" {
		t.Errorf("got module %q", deps.Module)
	}

	if deps.GoVersion != "1.22" {
		t.Errorf("got go version %q", deps.GoVersion)
	}

	if len(deps.Direct) != 2 {
		t.Errorf("expected 2 direct deps, got %d", len(deps.Direct))
	}

	if len(deps.Indirect) != 2 {
		t.Errorf("expected 2 indirect deps, got %d", len(deps.Indirect))
	}

	if deps.TotalCount != 4 {
		t.Errorf("expected total 4, got %d", deps.TotalCount)
	}
}

func TestParsePackageJSON(t *testing.T) {
	dir := t.TempDir()
	content := `{
  "name": "my-app",
  "version": "1.0.0",
  "dependencies": {
    "express": "^4.18.0",
    "lodash": "^4.17.0"
  },
  "devDependencies": {
    "jest": "^29.0.0"
  }
}`

	path := filepath.Join(dir, "package.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	deps := parsePackageJSON(path)
	if deps == nil {
		t.Fatal("expected non-nil deps")
	}

	if deps.Name != "my-app" {
		t.Errorf("got name %q", deps.Name)
	}

	if deps.PackageManager != "npm" {
		t.Errorf("got package manager %q", deps.PackageManager)
	}

	if len(deps.Dependencies) != 2 {
		t.Errorf("expected 2 deps, got %d", len(deps.Dependencies))
	}

	if len(deps.DevDependencies) != 1 {
		t.Errorf("expected 1 dev dep, got %d", len(deps.DevDependencies))
	}

	if deps.TotalCount != 3 {
		t.Errorf("expected total 3, got %d", deps.TotalCount)
	}
}

func TestParseRequirementsTxt(t *testing.T) {
	dir := t.TempDir()
	content := `flask==2.3.0
requests>=2.31.0
# comment
numpy~=1.25.0

-r other.txt
pandas
`

	path := filepath.Join(dir, "requirements.txt")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	deps := parseRequirementsTxt(path)
	if deps == nil {
		t.Fatal("expected non-nil deps")
	}

	if deps.TotalCount != 4 {
		t.Errorf("expected 4 deps, got %d: %v", deps.TotalCount, deps.Dependencies)
	}
}

func TestHealthScoring(t *testing.T) {
	dir := t.TempDir()

	// Create a project with some files
	files := map[string]string{
		"README.md":    "# Project",
		"LICENSE":      "MIT License\n\nPermission is hereby granted, free of charge",
		".gitignore":   "*.exe",
		"Taskfile.yml": "version: 3",
	}

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	report := &ProjectReport{
		Path:       dir,
		Name:       "test-project",
		BuildTools: []string{"Taskfile"},
		Docs:       checkDocs(dir),
	}

	health := computeHealth(dir, report)

	if health.Score <= 0 {
		t.Error("expected positive health score")
	}

	if health.Grade == "" {
		t.Error("expected non-empty grade")
	}

	// README(15) + LICENSE(10) + .gitignore(5) + Build(5) = 35
	expectedMin := 35
	if health.Score < expectedMin {
		t.Errorf("expected score >= %d, got %d", expectedMin, health.Score)
	}
}

func TestHealthGrades(t *testing.T) {
	tests := []struct {
		score int
		grade string
	}{
		{95, "A"},
		{90, "A"},
		{85, "B"},
		{80, "B"},
		{75, "C"},
		{70, "C"},
		{65, "D"},
		{60, "D"},
		{55, "F"},
		{0, "F"},
	}

	for _, tt := range tests {
		t.Run(tt.grade, func(t *testing.T) {
			grade := scoreToGrade(tt.score)
			if grade != tt.grade {
				t.Errorf("score %d: got grade %q, want %q", tt.score, grade, tt.grade)
			}
		})
	}
}

func scoreToGrade(score int) string {
	switch {
	case score >= 90:
		return "A"
	case score >= 80:
		return "B"
	case score >= 70:
		return "C"
	case score >= 60:
		return "D"
	default:
		return "F"
	}
}

func TestDetectTests(t *testing.T) {
	tests := []struct {
		name  string
		files []string
		want  bool
	}{
		{"go test files", []string{"foo_test.go"}, true},
		{"js test files", []string{"app.test.js"}, true},
		{"python test files", []string{"test_main.py"}, true},
		{"test directory", []string{"tests/"}, true},
		{"no tests", []string{"main.go"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			for _, f := range tt.files {
				if strings.HasSuffix(f, "/") {
					_ = os.MkdirAll(filepath.Join(dir, f), 0o755)
				} else {
					_ = os.WriteFile(filepath.Join(dir, f), []byte(""), 0o644)
				}
			}

			got := detectTests(dir)
			if got != tt.want {
				t.Errorf("detectTests() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectLicenseType(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{"MIT", "MIT License\n\nPermission is hereby granted, free of charge", "MIT"},
		{"Apache", "Apache License\nVersion 2.0", "Apache-2.0"},
		{"GPL3", "GNU General Public License\nVersion 3", "GPL-3.0"},
		{"Unknown", "Some custom license text", "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			path := filepath.Join(dir, "LICENSE")
			if err := os.WriteFile(path, []byte(tt.content), 0o644); err != nil {
				t.Fatal(err)
			}

			got := detectLicenseType(path)
			if got != tt.want {
				t.Errorf("detectLicenseType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatJSON(t *testing.T) {
	report := &ProjectReport{
		Path: "/tmp/test",
		Name: "test",
		Types: []ProjectType{
			{Language: "Go", BuildFile: "go.mod"},
		},
	}

	var buf bytes.Buffer
	if err := formatJSON(&buf, report); err != nil {
		t.Fatal(err)
	}

	var parsed ProjectReport
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if parsed.Name != "test" {
		t.Errorf("got name %q, want %q", parsed.Name, "test")
	}
}

func TestFormatText(t *testing.T) {
	report := &ProjectReport{
		Path: "/tmp/test",
		Name: "test-project",
		Types: []ProjectType{
			{Language: "Go", BuildFile: "go.mod", Frameworks: []string{"Cobra"}},
		},
		Languages: []LanguageInfo{
			{Name: "Go", FileCount: 10, Extensions: []string{".go"}},
		},
		BuildTools: []string{"Taskfile"},
	}

	var buf bytes.Buffer
	if err := formatText(&buf, report); err != nil {
		t.Fatal(err)
	}

	output := buf.String()

	if !strings.Contains(output, "test-project") {
		t.Error("expected project name in output")
	}

	if !strings.Contains(output, "Go") {
		t.Error("expected Go language in output")
	}

	if !strings.Contains(output, "Cobra") {
		t.Error("expected Cobra framework in output")
	}

	if !strings.Contains(output, "Taskfile") {
		t.Error("expected Taskfile in output")
	}
}

func TestFormatMarkdown(t *testing.T) {
	report := &ProjectReport{
		Path: "/tmp/test",
		Name: "test-project",
		Types: []ProjectType{
			{Language: "Go", BuildFile: "go.mod"},
		},
		Languages: []LanguageInfo{
			{Name: "Go", FileCount: 5, Extensions: []string{".go"}},
		},
	}

	var buf bytes.Buffer
	if err := formatMarkdown(&buf, report); err != nil {
		t.Fatal(err)
	}

	output := buf.String()

	if !strings.Contains(output, "# test-project") {
		t.Error("expected markdown header in output")
	}

	if !strings.Contains(output, "| Language |") {
		t.Error("expected markdown table in output")
	}
}

func TestCheckDocs(t *testing.T) {
	dir := t.TempDir()

	// Create documentation files
	_ = os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Hello"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("*.exe"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# Project"), 0o644)
	_ = os.MkdirAll(filepath.Join(dir, "docs"), 0o755)

	report := checkDocs(dir)

	if !report.HasReadme {
		t.Error("expected HasReadme = true")
	}

	if report.ReadmeFile != "README.md" {
		t.Errorf("got ReadmeFile %q", report.ReadmeFile)
	}

	if !report.HasGitignore {
		t.Error("expected HasGitignore = true")
	}

	if !report.HasClaudeMD {
		t.Error("expected HasClaudeMD = true")
	}

	if !report.HasDocsDir {
		t.Error("expected HasDocsDir = true")
	}

	if report.HasLicense {
		t.Error("expected HasLicense = false")
	}
}

func TestMatchFrameworks(t *testing.T) {
	deps := []string{
		"github.com/spf13/cobra",
		"github.com/go-chi/chi",
		"github.com/stretchr/testify",
	}

	found := matchFrameworks(deps, goFrameworks)

	if len(found) != 2 {
		t.Fatalf("expected 2 frameworks, got %d: %v", len(found), found)
	}

	fwSet := map[string]bool{}
	for _, f := range found {
		fwSet[f] = true
	}

	if !fwSet["Cobra"] {
		t.Error("expected Cobra framework")
	}

	if !fwSet["Chi"] {
		t.Error("expected Chi framework")
	}
}
