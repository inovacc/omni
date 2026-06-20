package project

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeFiles writes a map of relative path -> content into dir, creating
// parent directories as needed.
func writeFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()

	for name, content := range files {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestParsePyprojectToml(t *testing.T) {
	dir := t.TempDir()
	content := `[project]
name = "demo"
dependencies = [
  "flask>=2.3.0",
  "requests==2.31.0",
  "rich~=13.0",
  "fastapi[all]",
]
`
	path := filepath.Join(dir, "pyproject.toml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	deps := parsePyprojectToml(path)
	if deps == nil {
		t.Fatal("expected non-nil deps")
	}

	if deps.Source != "pyproject.toml" {
		t.Errorf("got source %q", deps.Source)
	}

	if deps.TotalCount != 4 {
		t.Errorf("expected 4 deps, got %d: %v", deps.TotalCount, deps.Dependencies)
	}

	// version specifiers and extras should be stripped
	for _, d := range deps.Dependencies {
		if strings.ContainsAny(d, ">=<~![") {
			t.Errorf("dependency %q still has version specifier", d)
		}
	}
}

func TestParsePyprojectToml_NoDeps(t *testing.T) {
	dir := t.TempDir()
	// valid TOML but no [project].dependencies
	path := filepath.Join(dir, "pyproject.toml")
	if err := os.WriteFile(path, []byte("[tool.black]\nline-length = 88\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if deps := parsePyprojectToml(path); deps != nil {
		t.Errorf("expected nil for empty dependencies, got %+v", deps)
	}

	// missing file -> nil
	if deps := parsePyprojectToml(filepath.Join(dir, "missing.toml")); deps != nil {
		t.Errorf("expected nil for missing file, got %+v", deps)
	}
}

func TestParseCargoToml(t *testing.T) {
	dir := t.TempDir()
	content := `[package]
name = "mycrate"
version = "0.1.0"
edition = "2021"

[dependencies]
serde = "1.0"
tokio = { version = "1", features = ["full"] }
`
	path := filepath.Join(dir, "Cargo.toml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	deps := parseCargoToml(path)
	if deps == nil {
		t.Fatal("expected non-nil deps")
	}

	if deps.Name != "mycrate" || deps.Version != "0.1.0" || deps.Edition != "2021" {
		t.Errorf("got %+v", deps)
	}

	if deps.TotalCount != 2 {
		t.Errorf("expected 2 deps, got %d: %v", deps.TotalCount, deps.Dependencies)
	}

	if deps := parseCargoToml(filepath.Join(dir, "missing.toml")); deps != nil {
		t.Errorf("expected nil for missing file")
	}
}

func TestParsePomXML(t *testing.T) {
	dir := t.TempDir()
	content := `<?xml version="1.0"?>
<project>
  <groupId>com.example</groupId>
  <artifactId>demo</artifactId>
  <dependencies>
    <dependency>
      <groupId>org.springframework</groupId>
      <artifactId>spring-core</artifactId>
    </dependency>
    <dependency>
      <groupId>junit</groupId>
      <artifactId>junit</artifactId>
    </dependency>
  </dependencies>
</project>`
	path := filepath.Join(dir, "pom.xml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	deps := parsePomXML(path)
	if deps == nil {
		t.Fatal("expected non-nil deps")
	}

	if deps.GroupID != "com.example" || deps.ArtifactID != "demo" {
		t.Errorf("got group/artifact %q/%q", deps.GroupID, deps.ArtifactID)
	}

	if deps.TotalCount != 2 {
		t.Errorf("expected 2 deps, got %d: %v", deps.TotalCount, deps.Dependencies)
	}

	if deps.Dependencies[0] != "org.springframework:spring-core" {
		t.Errorf("got dep %q", deps.Dependencies[0])
	}

	if deps := parsePomXML(filepath.Join(dir, "missing.xml")); deps != nil {
		t.Errorf("expected nil for missing file")
	}
}

func TestParseBuildGradle(t *testing.T) {
	dir := t.TempDir()
	content := `dependencies {
    implementation 'org.springframework:spring-core:5.0'
    api 'com.google.guava:guava:31.0'
    testImplementation 'junit:junit:4.13'
}`
	path := filepath.Join(dir, "build.gradle")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	deps := parseBuildGradle(path)
	if deps == nil {
		t.Fatal("expected non-nil deps")
	}

	if deps.TotalCount != 3 {
		t.Errorf("expected 3 deps, got %d: %v", deps.TotalCount, deps.Dependencies)
	}

	// no-dependency file -> nil
	empty := filepath.Join(dir, "empty.gradle")
	if err := os.WriteFile(empty, []byte("plugins { id 'java' }\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if d := parseBuildGradle(empty); d != nil {
		t.Errorf("expected nil for gradle with no deps, got %+v", d)
	}

	if d := parseBuildGradle(filepath.Join(dir, "missing.gradle")); d != nil {
		t.Errorf("expected nil for missing file")
	}
}

func TestParseGemfile(t *testing.T) {
	dir := t.TempDir()
	content := `source 'https://rubygems.org'
gem 'rails', '~> 7.0'
gem "puma"
# gem 'commented'
gem 'rspec', group: :test
`
	path := filepath.Join(dir, "Gemfile")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	deps := parseGemfile(path)
	if deps == nil {
		t.Fatal("expected non-nil deps")
	}

	if deps.TotalCount != 3 {
		t.Errorf("expected 3 gems, got %d: %v", deps.TotalCount, deps.Dependencies)
	}

	// no-gem file -> nil
	empty := filepath.Join(dir, "Empty")
	if err := os.WriteFile(empty, []byte("source 'https://rubygems.org'\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if d := parseGemfile(empty); d != nil {
		t.Errorf("expected nil for gemfile with no gems, got %+v", d)
	}

	if d := parseGemfile(filepath.Join(dir, "missing")); d != nil {
		t.Errorf("expected nil for missing file")
	}
}

func TestParseComposerJSON(t *testing.T) {
	dir := t.TempDir()
	content := `{
  "name": "vendor/pkg",
  "require": {
    "php": ">=8.0",
    "ext-json": "*",
    "monolog/monolog": "^3.0",
    "guzzlehttp/guzzle": "^7.0"
  }
}`
	path := filepath.Join(dir, "composer.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	deps := parseComposerJSON(path)
	if deps == nil {
		t.Fatal("expected non-nil deps")
	}

	if deps.Name != "vendor/pkg" {
		t.Errorf("got name %q", deps.Name)
	}

	// php and ext-* should be excluded
	if deps.TotalCount != 2 {
		t.Errorf("expected 2 deps, got %d: %v", deps.TotalCount, deps.Dependencies)
	}

	for _, d := range deps.Dependencies {
		if d == "php" || strings.HasPrefix(d, "ext-") {
			t.Errorf("dependency %q should have been excluded", d)
		}
	}

	if d := parseComposerJSON(filepath.Join(dir, "missing.json")); d != nil {
		t.Errorf("expected nil for missing file")
	}
}

func TestParseCsproj(t *testing.T) {
	dir := t.TempDir()
	content := `<Project Sdk="Microsoft.NET.Sdk">
  <ItemGroup>
    <PackageReference Include="Newtonsoft.Json" Version="13.0.1" />
    <PackageReference Include="Serilog" Version="2.0.0" />
  </ItemGroup>
</Project>`
	if err := os.WriteFile(filepath.Join(dir, "App.csproj"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	deps := parseCsproj(dir)
	if deps == nil {
		t.Fatal("expected non-nil deps")
	}

	if deps.TotalCount != 2 {
		t.Errorf("expected 2 packages, got %d: %v", deps.TotalCount, deps.Dependencies)
	}

	// directory with no csproj -> nil
	if d := parseCsproj(t.TempDir()); d != nil {
		t.Errorf("expected nil when no csproj present, got %+v", d)
	}
}

func TestParseInvalidFiles(t *testing.T) {
	dir := t.TempDir()

	// malformed package.json -> nil
	bad := filepath.Join(dir, "package.json")
	if err := os.WriteFile(bad, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	if d := parsePackageJSON(bad); d != nil {
		t.Errorf("expected nil for malformed package.json")
	}

	// missing files for the simple readers
	if d := parseGoMod(filepath.Join(dir, "nope.mod")); d != nil {
		t.Errorf("expected nil for missing go.mod")
	}

	if d := parseRequirementsTxt(filepath.Join(dir, "nope.txt")); d != nil {
		t.Errorf("expected nil for missing requirements.txt")
	}
}

func TestPackageManagerDetection(t *testing.T) {
	tests := []struct {
		lockfile string
		want     string
	}{
		{"pnpm-lock.yaml", "pnpm"},
		{"yarn.lock", "yarn"},
		{"bun.lockb", "bun"},
		{"", "npm"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			dir := t.TempDir()
			writeFiles(t, dir, map[string]string{
				"package.json": `{"name":"x","dependencies":{"a":"1"}}`,
			})

			if tt.lockfile != "" {
				if err := os.WriteFile(filepath.Join(dir, tt.lockfile), []byte(""), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			deps := parsePackageJSON(filepath.Join(dir, "package.json"))
			if deps == nil {
				t.Fatal("expected non-nil deps")
			}

			if deps.PackageManager != tt.want {
				t.Errorf("got package manager %q, want %q", deps.PackageManager, tt.want)
			}
		})
	}
}

func TestAnalyzeDeps(t *testing.T) {
	t.Run("multi-language", func(t *testing.T) {
		dir := t.TempDir()
		writeFiles(t, dir, map[string]string{
			"go.mod":           "module x\n\ngo 1.22\n",
			"package.json":     `{"name":"x","dependencies":{"react":"18"}}`,
			"requirements.txt": "flask\n",
			"Cargo.toml":       "[package]\nname=\"c\"\nversion=\"0.1.0\"\n",
			"Gemfile":          "gem 'rails'\n",
			"composer.json":    `{"name":"v/p","require":{"monolog/monolog":"^3"}}`,
		})

		types := DetectProjectTypes(dir)

		report := AnalyzeDeps(dir, types)
		if report == nil {
			t.Fatal("expected non-nil report")
		}

		if report.Go == nil || report.Node == nil || report.Python == nil ||
			report.Rust == nil || report.Ruby == nil || report.PHP == nil {
			t.Errorf("expected all ecosystems populated, got %+v", report)
		}
	})

	t.Run("pyproject fallback", func(t *testing.T) {
		dir := t.TempDir()
		writeFiles(t, dir, map[string]string{
			"pyproject.toml": "[project]\nname=\"p\"\ndependencies=[\"flask\"]\n",
		})

		types := DetectProjectTypes(dir)

		report := AnalyzeDeps(dir, types)
		if report == nil || report.Python == nil {
			t.Fatalf("expected python deps via pyproject fallback, got %+v", report)
		}

		if report.Python.Source != "pyproject.toml" {
			t.Errorf("got source %q", report.Python.Source)
		}
	})

	t.Run("empty returns nil", func(t *testing.T) {
		dir := t.TempDir()
		// Go type detected but go.mod content unreadable as deps still yields a GoDeps,
		// so use a type with no backing file to force the empty path.
		report := AnalyzeDeps(dir, []ProjectType{{Language: "Java", BuildFile: "pom.xml"}})
		if report != nil {
			t.Errorf("expected nil report when no deps files exist, got %+v", report)
		}
	})
}

func TestDetectFrameworks(t *testing.T) {
	t.Run("go and node", func(t *testing.T) {
		types := []ProjectType{
			{Language: "Go", BuildFile: "go.mod"},
			{Language: "JavaScript/TypeScript", BuildFile: "package.json"},
		}
		deps := &DepsReport{
			Go:   &GoDeps{Direct: []string{"github.com/spf13/cobra", "google.golang.org/grpc"}},
			Node: &NodeDeps{Dependencies: []string{"react"}, DevDependencies: []string{"svelte"}},
		}

		DetectFrameworks("", types, deps)

		if len(types[0].Frameworks) != 2 {
			t.Errorf("expected 2 go frameworks, got %v", types[0].Frameworks)
		}

		if len(types[1].Frameworks) != 2 {
			t.Errorf("expected 2 node frameworks, got %v", types[1].Frameworks)
		}
	})

	t.Run("nil deps is a no-op", func(t *testing.T) {
		types := []ProjectType{{Language: "Go", BuildFile: "go.mod"}}
		DetectFrameworks("", types, nil)

		if types[0].Frameworks != nil {
			t.Errorf("expected no frameworks with nil deps, got %v", types[0].Frameworks)
		}
	})
}

func TestDetectLicenseType_AllBranches(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{"BSD-3", "Redistribution and use in source and binary forms\nNeither the name of the foo", "BSD-3-Clause"},
		{"BSD-2", "Redistribution and use in source and binary forms are permitted", "BSD-2-Clause"},
		{"GPL2", "GNU General Public License\nVersion 2", "GPL-2.0"},
		{"LGPL", "GNU Lesser General Public License", "LGPL"},
		{"MPL2", "Mozilla Public License\nVersion 2.0", "MPL-2.0"},
		{"ISC", "ISC License\n", "ISC"},
		{"Unlicense", "This is free and unencumbered software released into the public domain.\nThe Unlicense", "Unlicense"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "LICENSE")
			if err := os.WriteFile(path, []byte(tt.content), 0o644); err != nil {
				t.Fatal(err)
			}

			if got := detectLicenseType(path); got != tt.want {
				t.Errorf("detectLicenseType() = %q, want %q", got, tt.want)
			}
		})
	}

	t.Run("missing file is Unknown", func(t *testing.T) {
		if got := detectLicenseType(filepath.Join(t.TempDir(), "nope")); got != "Unknown" {
			t.Errorf("got %q, want Unknown", got)
		}
	})
}

func TestDetectCIConfigs(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		".github/workflows/test.yml":  "name: test",
		".github/workflows/lint.yaml": "name: lint",
		".gitlab-ci.yml":              "stages: []",
		"Jenkinsfile":                 "pipeline {}",
		".circleci/config.yml":        "version: 2",
		".travis.yml":                 "language: go",
	})

	configs := detectCIConfigs(dir)
	if len(configs) < 6 {
		t.Errorf("expected at least 6 CI configs, got %d: %v", len(configs), configs)
	}

	joined := strings.Join(configs, " ")
	for _, want := range []string{"test.yml", ".gitlab-ci.yml", "Jenkinsfile", "config.yml", ".travis.yml"} {
		if !strings.Contains(joined, want) {
			t.Errorf("expected %q in CI configs: %v", want, configs)
		}
	}
}

func TestDetectLinterConfigs(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		".golangci.yml":  "linters: {}",
		".eslintrc.json": "{}",
		"rustfmt.toml":   "",
	})

	configs := detectLinterConfigs(dir)
	if len(configs) != 3 {
		t.Errorf("expected 3 linter configs, got %d: %v", len(configs), configs)
	}
}

// --- Run* entry points (text/JSON/markdown) ---

func newSampleProject(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"go.mod": `module github.com/example/demo

go 1.22

require github.com/spf13/cobra v1.8.0
`,
		"main.go":                  "package main\n",
		"main_test.go":             "package main\n",
		"README.md":                "# Demo",
		"LICENSE":                  "MIT License\n\nPermission is hereby granted, free of charge",
		".gitignore":               "*.exe",
		"Taskfile.yml":             "version: 3",
		".golangci.yml":            "linters: {}",
		".github/workflows/ci.yml": "name: ci",
		"CHANGELOG.md":             "# Changelog",
		"CONTRIBUTING.md":          "# Contributing",
		".editorconfig":            "root = true",
		"docs/guide.md":            "# Guide",
	})

	return dir
}

func TestRunInfo(t *testing.T) {
	dir := newSampleProject(t)

	t.Run("text", func(t *testing.T) {
		var buf bytes.Buffer
		if err := RunInfo(&buf, []string{dir}, Options{}); err != nil {
			t.Fatalf("RunInfo: %v", err)
		}

		out := buf.String()
		for _, want := range []string{"Project:", "Go", "Health Score", "Dependencies:"} {
			if !strings.Contains(out, want) {
				t.Errorf("expected %q in output:\n%s", want, out)
			}
		}
	})

	t.Run("json", func(t *testing.T) {
		var buf bytes.Buffer
		if err := RunInfo(&buf, []string{dir}, Options{JSON: true}); err != nil {
			t.Fatalf("RunInfo json: %v", err)
		}

		var report ProjectReport
		if err := json.Unmarshal(buf.Bytes(), &report); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}

		if report.Health == nil || report.Health.Score <= 0 {
			t.Errorf("expected positive health score, got %+v", report.Health)
		}
	})

	t.Run("markdown", func(t *testing.T) {
		var buf bytes.Buffer
		if err := RunInfo(&buf, []string{dir}, Options{Markdown: true}); err != nil {
			t.Fatalf("RunInfo markdown: %v", err)
		}

		out := buf.String()
		wantHeader := "# " + filepath.Base(dir)
		if !strings.HasPrefix(out, wantHeader) {
			t.Errorf("expected markdown header %q, got: %q", wantHeader, out[:min(20, len(out))])
		}

		if !strings.Contains(out, "## Health") {
			t.Error("expected Health section in markdown")
		}
	})

	t.Run("bad path", func(t *testing.T) {
		var buf bytes.Buffer
		if err := RunInfo(&buf, []string{filepath.Join(dir, "does", "not", "exist")}, Options{}); err == nil {
			t.Error("expected error for nonexistent path")
		}
	})
}

func TestRunDeps(t *testing.T) {
	dir := newSampleProject(t)

	for _, tc := range []struct {
		name string
		opts Options
		want string
	}{
		{"text", Options{}, "Go Module"},
		{"json", Options{JSON: true}, `"go"`},
		{"markdown", Options{Markdown: true}, "### Go"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := RunDeps(&buf, []string{dir}, tc.opts); err != nil {
				t.Fatalf("RunDeps: %v", err)
			}

			if !strings.Contains(buf.String(), tc.want) {
				t.Errorf("expected %q in output:\n%s", tc.want, buf.String())
			}
		})
	}

	t.Run("no deps", func(t *testing.T) {
		var buf bytes.Buffer
		if err := RunDeps(&buf, []string{t.TempDir()}, Options{}); err != nil {
			t.Fatalf("RunDeps: %v", err)
		}

		if !strings.Contains(buf.String(), "No dependency files found") {
			t.Errorf("expected no-deps message, got %q", buf.String())
		}
	})

	t.Run("bad path", func(t *testing.T) {
		var buf bytes.Buffer
		if err := RunDeps(&buf, []string{filepath.Join(dir, "nope-xyz")}, Options{}); err == nil {
			t.Error("expected error for nonexistent path")
		}
	})
}

func TestRunDocs(t *testing.T) {
	dir := newSampleProject(t)

	for _, tc := range []struct {
		name string
		opts Options
		want string
	}{
		{"text", Options{}, "Documentation:"},
		{"json", Options{JSON: true}, `"has_readme"`},
		{"markdown", Options{Markdown: true}, "| README |"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := RunDocs(&buf, []string{dir}, tc.opts); err != nil {
				t.Fatalf("RunDocs: %v", err)
			}

			if !strings.Contains(buf.String(), tc.want) {
				t.Errorf("expected %q in output:\n%s", tc.want, buf.String())
			}
		})
	}

	t.Run("bad path", func(t *testing.T) {
		var buf bytes.Buffer
		if err := RunDocs(&buf, []string{filepath.Join(dir, "nope-xyz")}, Options{}); err == nil {
			t.Error("expected error for nonexistent path")
		}
	})
}

func TestRunHealth(t *testing.T) {
	dir := newSampleProject(t)

	for _, tc := range []struct {
		name string
		opts Options
		want string
	}{
		{"text", Options{}, "Health Score:"},
		{"json", Options{JSON: true}, `"grade"`},
		{"markdown", Options{Markdown: true}, "**Score:**"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := RunHealth(&buf, []string{dir}, tc.opts); err != nil {
				t.Fatalf("RunHealth: %v", err)
			}

			if !strings.Contains(buf.String(), tc.want) {
				t.Errorf("expected %q in output:\n%s", tc.want, buf.String())
			}
		})
	}

	t.Run("bad path", func(t *testing.T) {
		var buf bytes.Buffer
		if err := RunHealth(&buf, []string{filepath.Join(dir, "nope-xyz")}, Options{}); err == nil {
			t.Error("expected error for nonexistent path")
		}
	})
}

// --- Direct format function coverage (incl. git formatters w/o exec) ---

func TestFormatGitFormatters(t *testing.T) {
	git := &GitReport{
		IsRepo:        true,
		Branch:        "main",
		Remote:        "origin",
		RemoteURL:     "https://example.com/x.git",
		Clean:         false,
		Ahead:         1,
		Behind:        2,
		TotalCommits:  10,
		Contributors:  3,
		Tags:          []string{"v1.0", "v1.1", "v1.2", "v1.3"},
		RecentCommits: []string{"abc fix", "def feat"},
	}

	t.Run("text", func(t *testing.T) {
		var buf bytes.Buffer
		if err := formatGitText(&buf, git); err != nil {
			t.Fatal(err)
		}

		out := buf.String()
		for _, want := range []string{"Branch:", "origin", "dirty", "Ahead/Behind: +1/-2", "Recent Commits:"} {
			if !strings.Contains(out, want) {
				t.Errorf("expected %q in git text:\n%s", want, out)
			}
		}
	})

	t.Run("markdown", func(t *testing.T) {
		var buf bytes.Buffer
		if err := formatGitMarkdown(&buf, git); err != nil {
			t.Fatal(err)
		}

		out := buf.String()
		for _, want := range []string{"**Branch:**", "**Remote:**", "dirty", "Recent Commits"} {
			if !strings.Contains(out, want) {
				t.Errorf("expected %q in git markdown:\n%s", want, out)
			}
		}
	})

	t.Run("json", func(t *testing.T) {
		var buf bytes.Buffer
		if err := formatGitJSON(&buf, git); err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(buf.String(), `"branch": "main"`) {
			t.Errorf("expected branch in git json:\n%s", buf.String())
		}
	})

	t.Run("clean repo no remote", func(t *testing.T) {
		var buf bytes.Buffer
		clean := &GitReport{IsRepo: true, Branch: "dev", Clean: true}
		if err := formatGitText(&buf, clean); err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(buf.String(), "clean") {
			t.Errorf("expected clean status: %s", buf.String())
		}

		var mbuf bytes.Buffer
		if err := formatGitMarkdown(&mbuf, clean); err != nil {
			t.Fatal(err)
		}

		if strings.Contains(mbuf.String(), "**Remote:**") {
			t.Error("did not expect Remote line when remote empty")
		}
	})
}

func TestFormatDepsJSON_Direct(t *testing.T) {
	report := &DepsReport{
		Go:     &GoDeps{Module: "x", GoVersion: "1.22", Direct: []string{"a"}, TotalCount: 1},
		Node:   &NodeDeps{Name: "n", Version: "1", PackageManager: "npm", TotalCount: 0},
		Python: &PythonDeps{Source: "requirements.txt", TotalCount: 5},
		Rust:   &RustDeps{Name: "r", Version: "0.1", Edition: "2021", TotalCount: 2},
		Java:   &JavaDeps{Source: "pom.xml", TotalCount: 1},
		Ruby:   &RubyDeps{TotalCount: 3},
		PHP:    &PHPDeps{Name: "p", TotalCount: 1},
		DotNet: &DotNetDeps{TotalCount: 4},
	}

	var jbuf bytes.Buffer
	if err := formatDepsJSON(&jbuf, report); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(jbuf.String(), `"module": "x"`) {
		t.Errorf("expected module in deps json:\n%s", jbuf.String())
	}

	var tbuf bytes.Buffer
	if err := formatDepsText(&tbuf, report); err != nil {
		t.Fatal(err)
	}

	out := tbuf.String()
	for _, want := range []string{"Go Module:", "Node.js:", "Python (", "Rust:", "Java (", "Ruby (", "PHP:", ".NET:"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in deps text:\n%s", want, out)
		}
	}

	var mbuf bytes.Buffer
	if err := formatDepsMarkdown(&mbuf, report); err != nil {
		t.Fatal(err)
	}

	mout := mbuf.String()
	for _, want := range []string{"### Go", "### Node.js", "### Python", "### Rust", "### Java", "### Ruby", "### PHP", "### .NET", "**Edition:**"} {
		if !strings.Contains(mout, want) {
			t.Errorf("expected %q in deps markdown:\n%s", want, mout)
		}
	}
}

func TestFormatDocsAndHealthJSON_Direct(t *testing.T) {
	docs := &DocsReport{
		HasReadme:     true,
		ReadmeFile:    "README.md",
		HasLicense:    true,
		LicenseType:   "MIT",
		CIConfigs:     []string{".github/workflows/ci.yml"},
		LinterConfigs: []string{".golangci.yml"},
	}

	var djson bytes.Buffer
	if err := formatDocsJSON(&djson, docs); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(djson.String(), `"has_readme": true`) {
		t.Errorf("expected has_readme in docs json:\n%s", djson.String())
	}

	// docs with no CI/linters exercises the "none" branch
	var dtext bytes.Buffer
	if err := formatDocsText(&dtext, &DocsReport{}); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(dtext.String(), "none") {
		t.Errorf("expected 'none' branch in docs text:\n%s", dtext.String())
	}

	var dmd bytes.Buffer
	if err := formatDocsMarkdown(&dmd, &DocsReport{}); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(dmd.String(), "| CI/CD | none |") {
		t.Errorf("expected none CI in docs markdown:\n%s", dmd.String())
	}

	health := &HealthReport{
		Score: 75,
		Grade: "C",
		Checks: []HealthCheck{
			{Name: "README", Passed: true, Points: 15, MaxPts: 15, Details: "ok"},
			{Name: "LICENSE", Passed: false, Points: 0, MaxPts: 10},
		},
	}

	var hjson bytes.Buffer
	if err := formatHealthJSON(&hjson, health); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(hjson.String(), `"grade": "C"`) {
		t.Errorf("expected grade in health json:\n%s", hjson.String())
	}
}

func TestCheckMarkHelpers(t *testing.T) {
	tests := []struct {
		ok     bool
		detail string
		text   string
		md     string
	}{
		{true, "MIT", "yes (MIT)", "Yes (MIT)"},
		{true, "", "yes", "Yes"},
		{false, "", "no", "No"},
		{false, "ignored", "no", "No"},
	}

	for _, tt := range tests {
		if got := checkMark(tt.ok, tt.detail); got != tt.text {
			t.Errorf("checkMark(%v,%q) = %q, want %q", tt.ok, tt.detail, got, tt.text)
		}

		if got := mdCheck(tt.ok, tt.detail); got != tt.md {
			t.Errorf("mdCheck(%v,%q) = %q, want %q", tt.ok, tt.detail, got, tt.md)
		}
	}
}

func TestResolvePath_File(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	// passing a file should resolve to its parent directory
	got, err := resolvePath([]string{f})
	if err != nil {
		t.Fatalf("resolvePath(file): %v", err)
	}

	if got != dir {
		t.Errorf("resolvePath(file) = %q, want parent %q", got, dir)
	}
}
