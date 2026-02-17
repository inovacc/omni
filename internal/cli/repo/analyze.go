package repo

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/inovacc/omni/internal/cli/project"
)

// AnalyzeReport holds the full analysis result.
type AnalyzeReport struct {
	Name         string                `json:"name"`
	Path         string                `json:"path"`
	Types        []project.ProjectType `json:"types"`
	Languages    []project.LanguageInfo `json:"languages"`
	BuildTools   []string              `json:"build_tools"`
	Health       *project.HealthReport `json:"health,omitempty"`
	Tree         string                `json:"tree,omitempty"`
	KeyFiles     []KeyFile             `json:"key_files,omitempty"`
	Deps         *project.DepsReport   `json:"deps,omitempty"`
	EntryPoints  []EntryPoint          `json:"entry_points,omitempty"`
	Architecture *Architecture         `json:"architecture,omitempty"`
	APISurface   []PackageAPI          `json:"api_surface,omitempty"`
	Git          *project.GitReport    `json:"git,omitempty"`
	TestPatterns []string              `json:"test_patterns,omitempty"`
	CIConfigs    []string              `json:"ci_configs,omitempty"`
	ConfigFiles  []string              `json:"config_files,omitempty"`
}

// KeyFile holds a file name and its content (or truncated content).
type KeyFile struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// EntryPoint describes a program entry point.
type EntryPoint struct {
	Path        string `json:"path"`
	Description string `json:"description,omitempty"`
}

// Architecture describes the inferred project architecture.
type Architecture struct {
	Pattern     string `json:"pattern"`
	PackageCount int   `json:"package_count"`
	Dirs        []string `json:"dirs,omitempty"`
}

// PackageAPI describes exported functions in a package.
type PackageAPI struct {
	Package       string `json:"package"`
	ExportedFuncs int    `json:"exported_funcs"`
}

// RunAnalyze orchestrates the repository analysis.
func RunAnalyze(w io.Writer, args []string, opts Options) error {
	target := "."
	if len(args) > 0 && args[0] != "" {
		target = args[0]
	}

	var dir string
	var cleanup func()

	if isRemote(target) {
		tmpDir, err := cloneToTemp(target)
		if err != nil {
			return fmt.Errorf("repo analyze: %w", err)
		}

		dir = tmpDir
		cleanup = func() { _ = os.RemoveAll(tmpDir) }
	} else {
		abs, err := resolvePath(target)
		if err != nil {
			return fmt.Errorf("repo analyze: %w", err)
		}

		dir = abs
	}

	if cleanup != nil {
		defer cleanup()
	}

	report := analyze(dir, opts)

	// Write to file if -o specified
	if opts.Output != "" {
		f, err := os.Create(opts.Output)
		if err != nil {
			return fmt.Errorf("repo analyze: %w", err)
		}
		defer func() { _ = f.Close() }()

		return formatAnalyze(f, report, opts)
	}

	return formatAnalyze(w, report, opts)
}

func analyze(dir string, opts Options) *AnalyzeReport {
	report := &AnalyzeReport{
		Name: filepath.Base(dir),
		Path: dir,
	}

	// Always collect overview data
	report.Types = project.DetectProjectTypes(dir)
	report.Languages = project.CountLanguages(dir)
	report.BuildTools = project.DetectBuildTools(dir)

	// Deps (needed for frameworks and health)
	if opts.wantSection("deps") || opts.wantSection("overview") {
		report.Deps = project.AnalyzeDeps(dir, report.Types)
		project.DetectFrameworks(dir, report.Types, report.Deps)
	}

	// Health (for overview grade)
	projReport := &project.ProjectReport{
		Path:       dir,
		Name:       report.Name,
		Types:      report.Types,
		BuildTools: report.BuildTools,
		Deps:       report.Deps,
	}

	projReport.Docs = project.CheckDocs(dir)

	if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
		projReport.Git = project.AnalyzeGit(dir, 1)
	}

	report.Health = project.ComputeHealth(dir, projReport)

	// Tree
	if opts.wantSection("tree") {
		report.Tree = buildTree(dir, 2)
	}

	// Key files
	if opts.wantSection("keys") {
		report.KeyFiles = collectKeyFiles(dir, opts.Compact)
	}

	// Entry points
	if opts.wantSection("overview") || opts.wantSection("api") {
		report.EntryPoints = findEntryPoints(dir)
	}

	// Architecture
	if opts.wantSection("overview") || opts.wantSection("api") {
		report.Architecture = inferArchitecture(dir)
	}

	// API surface
	if opts.wantSection("api") {
		report.APISurface = analyzeAPISurface(dir)
	}

	// Git
	if opts.wantSection("git") {
		report.Git = project.AnalyzeGit(dir, 10)
	}

	// Test patterns
	if opts.wantSection("tests") {
		report.TestPatterns = detectTestPatterns(dir)
	}

	// CI/CD
	if opts.wantSection("ci") {
		report.CIConfigs = projReport.Docs.CIConfigs
	}

	// Config files
	if opts.wantSection("ci") {
		report.ConfigFiles = findConfigFiles(dir)
	}

	return report
}

// buildTree generates a directory tree string up to maxDepth levels.
func buildTree(dir string, maxDepth int) string {
	var b strings.Builder

	skipSet := map[string]bool{
		".git": true, "node_modules": true, "vendor": true,
		"__pycache__": true, ".idea": true, ".vscode": true,
		"target": true, "dist": true, "bin": true,
		".next": true, ".nuxt": true, "build": true,
	}

	_ = buildTreeRecurse(&b, dir, "", 0, maxDepth, skipSet)

	return b.String()
}

func buildTreeRecurse(b *strings.Builder, dir, prefix string, depth, maxDepth int, skip map[string]bool) error {
	if depth >= maxDepth {
		return nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	// Filter to dirs and important files
	var filtered []os.DirEntry
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") && e.IsDir() {
			continue
		}

		if e.IsDir() && skip[name] {
			continue
		}

		filtered = append(filtered, e)
	}

	for i, e := range filtered {
		connector := "├── "
		childPrefix := prefix + "│   "

		if i == len(filtered)-1 {
			connector = "└── "
			childPrefix = prefix + "    "
		}

		name := e.Name()
		if e.IsDir() {
			name += "/"
		}

		_, _ = fmt.Fprintf(b, "%s%s%s\n", prefix, connector, name)

		if e.IsDir() {
			_ = buildTreeRecurse(b, filepath.Join(dir, e.Name()), childPrefix, depth+1, maxDepth, skip)
		}
	}

	return nil
}

// collectKeyFiles reads important project files.
func collectKeyFiles(dir string, compact bool) []KeyFile {
	candidates := []struct {
		name     string
		maxLines int
	}{
		{"CLAUDE.md", 0},       // full content always
		{"README.md", 100},     // first 100 lines
		{"Taskfile.yml", 0},    // full
		{"Taskfile.yaml", 0},   // full
		{"Makefile", 80},       // first 80 lines
		{"go.mod", 0},          // full
		{"package.json", 0},    // full
		{"Cargo.toml", 0},      // full
		{"pyproject.toml", 0},  // full
		{"Dockerfile", 0},      // full
	}

	if compact {
		candidates[0].maxLines = 50  // CLAUDE.md truncated
		candidates[1].maxLines = 30  // README.md more truncated
	}

	var files []KeyFile

	for _, c := range candidates {
		path := filepath.Join(dir, c.name)

		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		content := string(data)

		if c.maxLines > 0 {
			lines := strings.SplitN(content, "\n", c.maxLines+1)
			if len(lines) > c.maxLines {
				content = strings.Join(lines[:c.maxLines], "\n") + "\n...(truncated)"
			}
		}

		files = append(files, KeyFile{Name: c.name, Content: content})
	}

	return files
}

// findEntryPoints locates program entry points.
func findEntryPoints(dir string) []EntryPoint {
	var entries []EntryPoint

	// Check cmd/*/main.go pattern
	cmdDir := filepath.Join(dir, "cmd")
	if dirEntries, err := os.ReadDir(cmdDir); err == nil {
		for _, e := range dirEntries {
			if !e.IsDir() {
				continue
			}

			mainPath := filepath.Join(cmdDir, e.Name(), "main.go")
			if _, err := os.Stat(mainPath); err == nil {
				rel, _ := filepath.Rel(dir, mainPath)
				entries = append(entries, EntryPoint{
					Path:        rel,
					Description: e.Name() + " binary",
				})
			}
		}
	}

	// Check root main.go
	rootMain := filepath.Join(dir, "main.go")
	if _, err := os.Stat(rootMain); err == nil {
		entries = append(entries, EntryPoint{
			Path:        "main.go",
			Description: "root entry point",
		})
	}

	return entries
}

// inferArchitecture detects the project structure pattern.
func inferArchitecture(dir string) *Architecture {
	arch := &Architecture{}

	hasCMD := dirExists(filepath.Join(dir, "cmd"))
	hasInternal := dirExists(filepath.Join(dir, "internal"))
	hasPkg := dirExists(filepath.Join(dir, "pkg"))
	hasSrc := dirExists(filepath.Join(dir, "src"))

	switch {
	case hasCMD && hasInternal && hasPkg:
		arch.Pattern = "hexagonal/clean"
	case hasCMD && hasInternal:
		arch.Pattern = "standard Go layout"
	case hasSrc:
		arch.Pattern = "src-based"
	case hasCMD:
		arch.Pattern = "multi-binary"
	default:
		arch.Pattern = "flat"
	}

	// Count top-level packages
	var topDirs []string

	entries, err := os.ReadDir(dir)
	if err == nil {
		for _, e := range entries {
			if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
				topDirs = append(topDirs, e.Name())
			}
		}
	}

	arch.PackageCount = len(topDirs)
	arch.Dirs = topDirs

	return arch
}

// analyzeAPISurface counts exported functions in pkg/ packages for Go projects.
func analyzeAPISurface(dir string) []PackageAPI {
	pkgDir := filepath.Join(dir, "pkg")
	if !dirExists(pkgDir) {
		return nil
	}

	var apis []PackageAPI

	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		return nil
	}

	fset := token.NewFileSet()

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		subDir := filepath.Join(pkgDir, e.Name())
		count := countExportedFuncs(fset, subDir)

		if count > 0 {
			apis = append(apis, PackageAPI{
				Package:       "pkg/" + e.Name(),
				ExportedFuncs: count,
			})
		}
	}

	return apis
}

func countExportedFuncs(fset *token.FileSet, dir string) int {
	pkgs, err := parser.ParseDir(fset, dir, func(fi fs.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil {
		return 0
	}

	count := 0

	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.IsExported() {
					count++
				}
			}
		}
	}

	return count
}

// detectTestPatterns identifies testing approaches used in the project.
func detectTestPatterns(dir string) []string {
	var patterns []string
	seen := make(map[string]bool)

	add := func(p string) {
		if !seen[p] {
			patterns = append(patterns, p)
			seen[p] = true
		}
	}

	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		name := d.Name()

		if d.IsDir() {
			switch {
			case name == ".git" || name == "node_modules" || name == "vendor":
				return filepath.SkipDir
			case name == "testdata" || name == "fixtures":
				add("fixtures/testdata")
			case name == "golden":
				add("golden master")
			case name == "integration" || name == "e2e":
				add("integration/e2e")
			}

			return nil
		}

		if strings.HasSuffix(name, "_test.go") {
			add("table-driven (Go)")
		}

		if strings.HasSuffix(name, ".test.js") || strings.HasSuffix(name, ".spec.ts") {
			add("JS/TS test files")
		}

		if strings.HasPrefix(name, "test_") && strings.HasSuffix(name, ".py") {
			add("pytest")
		}

		return nil
	})

	return patterns
}

// findConfigFiles looks for infrastructure/config files.
func findConfigFiles(dir string) []string {
	candidates := []string{
		".env.example", ".env.sample",
		"docker-compose.yml", "docker-compose.yaml",
		"k8s", "kubernetes", "deploy",
		"terraform", ".terraform",
		"ansible", "helm",
	}

	var found []string

	for _, name := range candidates {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err == nil {
			found = append(found, name)
		}
	}

	return found
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
