package project

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// computeHealth computes a health score for the project.
func computeHealth(dir string, report *ProjectReport) *HealthReport {
	health := &HealthReport{}

	docs := report.Docs
	if docs == nil {
		docs = checkDocs(dir)
	}

	// 1. README (15 points)
	health.addCheck("README", 15, docs.HasReadme, "")

	// 2. LICENSE (10 points)
	detail := ""
	if docs.HasLicense {
		detail = docs.LicenseType
	}

	health.addCheck("LICENSE", 10, docs.HasLicense, detail)

	// 3. .gitignore (5 points)
	health.addCheck(".gitignore", 5, docs.HasGitignore, "")

	// 4. CI/CD (15 points)
	health.addCheck("CI/CD", 15, len(docs.CIConfigs) > 0, "")

	// 5. Tests (15 points)
	hasTests := detectTests(dir)
	health.addCheck("Tests", 15, hasTests, "")

	// 6. Linter config (10 points)
	health.addCheck("Linter config", 10, len(docs.LinterConfigs) > 0, "")

	// 7. Git clean (5 points)
	gitClean := false
	if report.Git != nil {
		gitClean = report.Git.Clean
	}

	health.addCheck("Git clean", 5, gitClean, "")

	// 8. CONTRIBUTING (5 points)
	health.addCheck("CONTRIBUTING", 5, docs.HasContributing, "")

	// 9. docs/ directory (5 points)
	health.addCheck("docs/ directory", 5, docs.HasDocsDir, "")

	// 10. CHANGELOG (5 points)
	health.addCheck("CHANGELOG", 5, docs.HasChangelog, "")

	// 11. .editorconfig (5 points)
	health.addCheck(".editorconfig", 5, docs.HasEditorconfig, "")

	// 12. Build automation (5 points)
	hasBuild := false
	if len(report.BuildTools) > 0 {
		hasBuild = true
	}

	health.addCheck("Build automation", 5, hasBuild, "")

	// Calculate total score
	for _, c := range health.Checks {
		health.Score += c.Points
	}

	// Assign grade
	switch {
	case health.Score >= 90:
		health.Grade = "A"
	case health.Score >= 80:
		health.Grade = "B"
	case health.Score >= 70:
		health.Grade = "C"
	case health.Score >= 60:
		health.Grade = "D"
	default:
		health.Grade = "F"
	}

	return health
}

func (h *HealthReport) addCheck(name string, maxPts int, passed bool, details string) {
	pts := 0
	if passed {
		pts = maxPts
	}

	h.Checks = append(h.Checks, HealthCheck{
		Name:    name,
		Passed:  passed,
		Points:  pts,
		MaxPts:  maxPts,
		Details: details,
	})
}

// detectTests checks for test files in the project.
func detectTests(dir string) bool {
	found := false

	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || found {
			return filepath.SkipAll
		}

		name := d.Name()

		if d.IsDir() {
			if skipDirs[name] || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}

			// Check for test directories
			lower := strings.ToLower(name)
			if lower == "tests" || lower == "test" || lower == "__tests__" || lower == "spec" {
				found = true

				return filepath.SkipAll
			}

			return nil
		}

		// Check for test files
		lower := strings.ToLower(name)
		if strings.HasSuffix(lower, "_test.go") ||
			strings.HasSuffix(lower, ".test.js") ||
			strings.HasSuffix(lower, ".test.ts") ||
			strings.HasSuffix(lower, ".test.tsx") ||
			strings.HasSuffix(lower, ".spec.js") ||
			strings.HasSuffix(lower, ".spec.ts") ||
			strings.HasSuffix(lower, "_test.py") ||
			strings.HasSuffix(lower, "_test.rs") ||
			strings.HasPrefix(lower, "test_") {
			found = true

			return filepath.SkipAll
		}

		return nil
	})

	return found
}

// RunHealth runs the health score subcommand.
func RunHealth(w io.Writer, args []string, opts Options) error {
	dir, err := resolvePath(args)
	if err != nil {
		return fmt.Errorf("project health: %w", err)
	}

	report := &ProjectReport{
		Path: dir,
		Name: filepath.Base(dir),
	}

	report.Types = detectProjectTypes(dir)
	report.BuildTools = detectBuildTools(dir)
	report.Docs = checkDocs(dir)

	// Only check git if .git exists
	if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
		report.Git = analyzeGit(dir, 1)
	}

	health := computeHealth(dir, report)

	if opts.JSON {
		return formatHealthJSON(w, health)
	}

	if opts.Markdown {
		return formatHealthMarkdown(w, health)
	}

	return formatHealthText(w, health)
}
