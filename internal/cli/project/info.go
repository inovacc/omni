package project

import (
	"fmt"
	"io"
	"path/filepath"
)

// RunInfo runs the full project analysis.
func RunInfo(w io.Writer, args []string, opts Options) error {
	dir, err := resolvePath(args)
	if err != nil {
		return fmt.Errorf("project info: %w", err)
	}

	report := &ProjectReport{
		Path: dir,
		Name: filepath.Base(dir),
	}

	// Detect project types
	report.Types = detectProjectTypes(dir)

	// Count languages
	report.Languages = countLanguages(dir)

	// Detect build tools
	report.BuildTools = detectBuildTools(dir)

	// Parse dependencies
	report.Deps = analyzeDeps(dir, report.Types)

	// Detect frameworks (needs deps)
	detectFrameworks(dir, report.Types, report.Deps)

	// Git info
	report.Git = analyzeGit(dir, opts.Limit)

	// Docs check
	report.Docs = checkDocs(dir)

	// Health score
	report.Health = computeHealth(dir, report)

	return formatReport(w, report, opts)
}
