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
	report.Types = DetectProjectTypes(dir)

	// Count languages
	report.Languages = CountLanguages(dir)

	// Detect build tools
	report.BuildTools = DetectBuildTools(dir)

	// Parse dependencies
	report.Deps = AnalyzeDeps(dir, report.Types)

	// Detect frameworks (needs deps)
	DetectFrameworks(dir, report.Types, report.Deps)

	// Git info
	report.Git = AnalyzeGit(dir, opts.Limit)

	// Docs check
	report.Docs = CheckDocs(dir)

	// Health score
	report.Health = ComputeHealth(dir, report)

	return formatReport(w, report, opts)
}
