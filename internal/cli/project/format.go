package project

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// --- Full Report Formatters ---

func formatReport(w io.Writer, report *ProjectReport, opts Options) error {
	if opts.JSON {
		return formatJSON(w, report)
	}

	if opts.Markdown {
		return formatMarkdown(w, report)
	}

	return formatText(w, report)
}

func formatJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	return enc.Encode(v)
}

func formatText(w io.Writer, r *ProjectReport) error {
	_, _ = fmt.Fprintf(w, "Project: %s\n", r.Name)
	_, _ = fmt.Fprintf(w, "Path:    %s\n", r.Path)
	_, _ = fmt.Fprintln(w)

	// Project types
	if len(r.Types) > 0 {
		_, _ = fmt.Fprintln(w, "Project Types:")

		for _, t := range r.Types {
			line := fmt.Sprintf("  %s (%s)", t.Language, t.BuildFile)
			if len(t.Frameworks) > 0 {
				line += " [" + strings.Join(t.Frameworks, ", ") + "]"
			}

			_, _ = fmt.Fprintln(w, line)
		}

		_, _ = fmt.Fprintln(w)
	}

	// Languages
	if len(r.Languages) > 0 {
		_, _ = fmt.Fprintln(w, "Languages:")
		for _, l := range r.Languages {
			_, _ = fmt.Fprintf(w, "  %-15s %5d files  %s\n", l.Name, l.FileCount, strings.Join(l.Extensions, ", "))
		}

		_, _ = fmt.Fprintln(w)
	}

	// Build tools
	if len(r.BuildTools) > 0 {
		_, _ = fmt.Fprintf(w, "Build Tools: %s\n\n", strings.Join(r.BuildTools, ", "))
	}

	// Dependencies
	if r.Deps != nil {
		_, _ = fmt.Fprintln(w, "Dependencies:")
		_ = formatDepsTextInline(w, r.Deps)
		_, _ = fmt.Fprintln(w)
	}

	// Git
	if r.Git != nil && r.Git.IsRepo {
		_ = formatGitText(w, r.Git)
		_, _ = fmt.Fprintln(w)
	}

	// Docs
	if r.Docs != nil {
		_ = formatDocsText(w, r.Docs)
		_, _ = fmt.Fprintln(w)
	}

	// Health
	if r.Health != nil {
		_ = formatHealthText(w, r.Health)
	}

	return nil
}

func formatMarkdown(w io.Writer, r *ProjectReport) error {
	_, _ = fmt.Fprintf(w, "# %s\n\n", r.Name)
	_, _ = fmt.Fprintf(w, "**Path:** `%s`\n\n", r.Path)

	// Project types
	if len(r.Types) > 0 {
		_, _ = fmt.Fprintf(w, "## Project Types\n\n")
		_, _ = fmt.Fprintln(w, "| Language | Build File | Frameworks |")
		_, _ = fmt.Fprintln(w, "|----------|------------|------------|")

		for _, t := range r.Types {
			fw := ""
			if len(t.Frameworks) > 0 {
				fw = strings.Join(t.Frameworks, ", ")
			}

			_, _ = fmt.Fprintf(w, "| %s | %s | %s |\n", t.Language, t.BuildFile, fw)
		}

		_, _ = fmt.Fprintln(w)
	}

	// Languages
	if len(r.Languages) > 0 {
		_, _ = fmt.Fprintf(w, "## Languages\n\n")
		_, _ = fmt.Fprintln(w, "| Language | Files | Extensions |")
		_, _ = fmt.Fprintln(w, "|----------|-------|------------|")

		for _, l := range r.Languages {
			_, _ = fmt.Fprintf(w, "| %s | %d | %s |\n", l.Name, l.FileCount, strings.Join(l.Extensions, ", "))
		}

		_, _ = fmt.Fprintln(w)
	}

	// Build tools
	if len(r.BuildTools) > 0 {
		_, _ = fmt.Fprintf(w, "## Build Tools\n\n%s\n\n", strings.Join(r.BuildTools, ", "))
	}

	// Dependencies
	if r.Deps != nil {
		_, _ = fmt.Fprintf(w, "## Dependencies\n\n")
		_ = formatDepsMarkdown(w, r.Deps)
		_, _ = fmt.Fprintln(w)
	}

	// Git
	if r.Git != nil && r.Git.IsRepo {
		_, _ = fmt.Fprintf(w, "## Git\n\n")
		_ = formatGitMarkdown(w, r.Git)
		_, _ = fmt.Fprintln(w)
	}

	// Docs
	if r.Docs != nil {
		_, _ = fmt.Fprintf(w, "## Documentation\n\n")
		_ = formatDocsMarkdown(w, r.Docs)
		_, _ = fmt.Fprintln(w)
	}

	// Health
	if r.Health != nil {
		_, _ = fmt.Fprintf(w, "## Health\n\n")
		_ = formatHealthMarkdown(w, r.Health)
	}

	return nil
}

// --- Deps Formatters ---

func formatDepsJSON(w io.Writer, report *DepsReport) error {
	return formatJSON(w, report)
}

func formatDepsText(w io.Writer, report *DepsReport) error {
	_, _ = fmt.Fprintln(w, "Dependencies:")
	return formatDepsTextInline(w, report)
}

func formatDepsTextInline(w io.Writer, report *DepsReport) error {
	if report.Go != nil {
		_, _ = fmt.Fprintf(w, "  Go Module: %s (go %s)\n", report.Go.Module, report.Go.GoVersion)
		_, _ = fmt.Fprintf(w, "    Direct: %d, Indirect: %d, Total: %d\n",
			len(report.Go.Direct), len(report.Go.Indirect), report.Go.TotalCount)
	}

	if report.Node != nil {
		_, _ = fmt.Fprintf(w, "  Node.js: %s@%s (%s)\n", report.Node.Name, report.Node.Version, report.Node.PackageManager)
		_, _ = fmt.Fprintf(w, "    Dependencies: %d, DevDependencies: %d, Total: %d\n",
			len(report.Node.Dependencies), len(report.Node.DevDependencies), report.Node.TotalCount)
	}

	if report.Python != nil {
		_, _ = fmt.Fprintf(w, "  Python (%s): %d dependencies\n", report.Python.Source, report.Python.TotalCount)
	}

	if report.Rust != nil {
		_, _ = fmt.Fprintf(w, "  Rust: %s@%s (edition %s)\n", report.Rust.Name, report.Rust.Version, report.Rust.Edition)
		_, _ = fmt.Fprintf(w, "    Dependencies: %d\n", report.Rust.TotalCount)
	}

	if report.Java != nil {
		_, _ = fmt.Fprintf(w, "  Java (%s): %d dependencies\n", report.Java.Source, report.Java.TotalCount)
	}

	if report.Ruby != nil {
		_, _ = fmt.Fprintf(w, "  Ruby (Gemfile): %d gems\n", report.Ruby.TotalCount)
	}

	if report.PHP != nil {
		_, _ = fmt.Fprintf(w, "  PHP: %s, %d dependencies\n", report.PHP.Name, report.PHP.TotalCount)
	}

	if report.DotNet != nil {
		_, _ = fmt.Fprintf(w, "  .NET: %d packages\n", report.DotNet.TotalCount)
	}

	return nil
}

func formatDepsMarkdown(w io.Writer, report *DepsReport) error {
	if report.Go != nil {
		_, _ = fmt.Fprintf(w, "### Go\n\n")
		_, _ = fmt.Fprintf(w, "- **Module:** `%s`\n", report.Go.Module)
		_, _ = fmt.Fprintf(w, "- **Go Version:** %s\n", report.Go.GoVersion)
		_, _ = fmt.Fprintf(w, "- **Direct:** %d | **Indirect:** %d | **Total:** %d\n\n",
			len(report.Go.Direct), len(report.Go.Indirect), report.Go.TotalCount)
	}

	if report.Node != nil {
		_, _ = fmt.Fprintf(w, "### Node.js\n\n")
		_, _ = fmt.Fprintf(w, "- **Name:** %s@%s\n", report.Node.Name, report.Node.Version)
		_, _ = fmt.Fprintf(w, "- **Package Manager:** %s\n", report.Node.PackageManager)
		_, _ = fmt.Fprintf(w, "- **Dependencies:** %d | **Dev:** %d | **Total:** %d\n\n",
			len(report.Node.Dependencies), len(report.Node.DevDependencies), report.Node.TotalCount)
	}

	if report.Python != nil {
		_, _ = fmt.Fprintf(w, "### Python (%s)\n\n", report.Python.Source)
		_, _ = fmt.Fprintf(w, "- **Dependencies:** %d\n\n", report.Python.TotalCount)
	}

	if report.Rust != nil {
		_, _ = fmt.Fprintf(w, "### Rust\n\n")

		_, _ = fmt.Fprintf(w, "- **Name:** %s@%s\n", report.Rust.Name, report.Rust.Version)
		if report.Rust.Edition != "" {
			_, _ = fmt.Fprintf(w, "- **Edition:** %s\n", report.Rust.Edition)
		}

		_, _ = fmt.Fprintf(w, "- **Dependencies:** %d\n\n", report.Rust.TotalCount)
	}

	if report.Java != nil {
		_, _ = fmt.Fprintf(w, "### Java (%s)\n\n", report.Java.Source)
		_, _ = fmt.Fprintf(w, "- **Dependencies:** %d\n\n", report.Java.TotalCount)
	}

	if report.Ruby != nil {
		_, _ = fmt.Fprintf(w, "### Ruby\n\n")
		_, _ = fmt.Fprintf(w, "- **Gems:** %d\n\n", report.Ruby.TotalCount)
	}

	if report.PHP != nil {
		_, _ = fmt.Fprintf(w, "### PHP\n\n")
		_, _ = fmt.Fprintf(w, "- **Name:** %s\n", report.PHP.Name)
		_, _ = fmt.Fprintf(w, "- **Dependencies:** %d\n\n", report.PHP.TotalCount)
	}

	if report.DotNet != nil {
		_, _ = fmt.Fprintf(w, "### .NET\n\n")
		_, _ = fmt.Fprintf(w, "- **Packages:** %d\n\n", report.DotNet.TotalCount)
	}

	return nil
}

// --- Git Formatters ---

func formatGitJSON(w io.Writer, report *GitReport) error {
	return formatJSON(w, report)
}

func formatGitText(w io.Writer, r *GitReport) error {
	_, _ = fmt.Fprintln(w, "Git:")
	_, _ = fmt.Fprintf(w, "  Branch:       %s\n", r.Branch)

	if r.Remote != "" {
		_, _ = fmt.Fprintf(w, "  Remote:       %s\n", r.Remote)
		_, _ = fmt.Fprintf(w, "  Remote URL:   %s\n", r.RemoteURL)
	}

	status := "dirty"
	if r.Clean {
		status = "clean"
	}

	_, _ = fmt.Fprintf(w, "  Status:       %s\n", status)

	if r.Ahead > 0 || r.Behind > 0 {
		_, _ = fmt.Fprintf(w, "  Ahead/Behind: +%d/-%d\n", r.Ahead, r.Behind)
	}

	_, _ = fmt.Fprintf(w, "  Commits:      %d\n", r.TotalCommits)
	_, _ = fmt.Fprintf(w, "  Contributors: %d\n", r.Contributors)

	if len(r.Tags) > 0 {
		_, _ = fmt.Fprintf(w, "  Latest Tags:  %s\n", strings.Join(r.Tags[:min(3, len(r.Tags))], ", "))
	}

	if len(r.RecentCommits) > 0 {
		_, _ = fmt.Fprintln(w, "  Recent Commits:")
		for _, c := range r.RecentCommits {
			_, _ = fmt.Fprintf(w, "    %s\n", c)
		}
	}

	return nil
}

func formatGitMarkdown(w io.Writer, r *GitReport) error {
	_, _ = fmt.Fprintf(w, "- **Branch:** %s\n", r.Branch)

	if r.Remote != "" {
		_, _ = fmt.Fprintf(w, "- **Remote:** %s (`%s`)\n", r.Remote, r.RemoteURL)
	}

	status := "dirty"
	if r.Clean {
		status = "clean"
	}

	_, _ = fmt.Fprintf(w, "- **Status:** %s\n", status)

	if r.Ahead > 0 || r.Behind > 0 {
		_, _ = fmt.Fprintf(w, "- **Ahead/Behind:** +%d/-%d\n", r.Ahead, r.Behind)
	}

	_, _ = fmt.Fprintf(w, "- **Total Commits:** %d\n", r.TotalCommits)
	_, _ = fmt.Fprintf(w, "- **Contributors:** %d\n", r.Contributors)

	if len(r.Tags) > 0 {
		_, _ = fmt.Fprintf(w, "- **Latest Tags:** %s\n", strings.Join(r.Tags[:min(3, len(r.Tags))], ", "))
	}

	if len(r.RecentCommits) > 0 {
		_, _ = fmt.Fprintf(w, "\n### Recent Commits\n\n")
		for _, c := range r.RecentCommits {
			_, _ = fmt.Fprintf(w, "- `%s`\n", c)
		}
	}

	return nil
}

// --- Docs Formatters ---

func formatDocsJSON(w io.Writer, report *DocsReport) error {
	return formatJSON(w, report)
}

func formatDocsText(w io.Writer, r *DocsReport) error {
	_, _ = fmt.Fprintln(w, "Documentation:")
	_, _ = fmt.Fprintf(w, "  README:        %s\n", checkMark(r.HasReadme, r.ReadmeFile))
	_, _ = fmt.Fprintf(w, "  LICENSE:       %s\n", checkMark(r.HasLicense, r.LicenseType))
	_, _ = fmt.Fprintf(w, "  CHANGELOG:     %s\n", checkMark(r.HasChangelog, ""))
	_, _ = fmt.Fprintf(w, "  CONTRIBUTING:  %s\n", checkMark(r.HasContributing, ""))
	_, _ = fmt.Fprintf(w, "  CLAUDE.md:     %s\n", checkMark(r.HasClaudeMD, ""))
	_, _ = fmt.Fprintf(w, "  docs/:         %s\n", checkMark(r.HasDocsDir, ""))
	_, _ = fmt.Fprintf(w, "  .gitignore:    %s\n", checkMark(r.HasGitignore, ""))
	_, _ = fmt.Fprintf(w, "  .editorconfig: %s\n", checkMark(r.HasEditorconfig, ""))

	if len(r.CIConfigs) > 0 {
		_, _ = fmt.Fprintf(w, "  CI/CD:         %s\n", strings.Join(r.CIConfigs, ", "))
	} else {
		_, _ = fmt.Fprintln(w, "  CI/CD:         none")
	}

	if len(r.LinterConfigs) > 0 {
		_, _ = fmt.Fprintf(w, "  Linters:       %s\n", strings.Join(r.LinterConfigs, ", "))
	} else {
		_, _ = fmt.Fprintln(w, "  Linters:       none")
	}

	return nil
}

func formatDocsMarkdown(w io.Writer, r *DocsReport) error {
	_, _ = fmt.Fprintln(w, "| Item | Status |")
	_, _ = fmt.Fprintln(w, "|------|--------|")
	_, _ = fmt.Fprintf(w, "| README | %s |\n", mdCheck(r.HasReadme, r.ReadmeFile))
	_, _ = fmt.Fprintf(w, "| LICENSE | %s |\n", mdCheck(r.HasLicense, r.LicenseType))
	_, _ = fmt.Fprintf(w, "| CHANGELOG | %s |\n", mdCheck(r.HasChangelog, ""))
	_, _ = fmt.Fprintf(w, "| CONTRIBUTING | %s |\n", mdCheck(r.HasContributing, ""))
	_, _ = fmt.Fprintf(w, "| CLAUDE.md | %s |\n", mdCheck(r.HasClaudeMD, ""))
	_, _ = fmt.Fprintf(w, "| docs/ | %s |\n", mdCheck(r.HasDocsDir, ""))
	_, _ = fmt.Fprintf(w, "| .gitignore | %s |\n", mdCheck(r.HasGitignore, ""))
	_, _ = fmt.Fprintf(w, "| .editorconfig | %s |\n", mdCheck(r.HasEditorconfig, ""))

	ciStatus := "none"
	if len(r.CIConfigs) > 0 {
		ciStatus = strings.Join(r.CIConfigs, ", ")
	}

	_, _ = fmt.Fprintf(w, "| CI/CD | %s |\n", ciStatus)

	lintStatus := "none"
	if len(r.LinterConfigs) > 0 {
		lintStatus = strings.Join(r.LinterConfigs, ", ")
	}

	_, _ = fmt.Fprintf(w, "| Linters | %s |\n", lintStatus)

	return nil
}

// --- Health Formatters ---

func formatHealthJSON(w io.Writer, report *HealthReport) error {
	return formatJSON(w, report)
}

func formatHealthText(w io.Writer, r *HealthReport) error {
	_, _ = fmt.Fprintf(w, "Health Score: %d/100 (Grade: %s)\n\n", r.Score, r.Grade)

	for _, c := range r.Checks {
		mark := "[x]"
		if !c.Passed {
			mark = "[ ]"
		}

		line := fmt.Sprintf("  %s %-25s %d/%d", mark, c.Name, c.Points, c.MaxPts)
		if c.Details != "" {
			line += "  " + c.Details
		}

		_, _ = fmt.Fprintln(w, line)
	}

	return nil
}

func formatHealthMarkdown(w io.Writer, r *HealthReport) error {
	_, _ = fmt.Fprintf(w, "**Score:** %d/100 | **Grade:** %s\n\n", r.Score, r.Grade)
	_, _ = fmt.Fprintln(w, "| Check | Score | Status |")
	_, _ = fmt.Fprintln(w, "|-------|-------|--------|")

	for _, c := range r.Checks {
		status := "Pass"
		if !c.Passed {
			status = "Fail"
		}

		_, _ = fmt.Fprintf(w, "| %s | %d/%d | %s |\n", c.Name, c.Points, c.MaxPts, status)
	}

	return nil
}

// --- Helpers ---

func checkMark(ok bool, detail string) string {
	if ok {
		if detail != "" {
			return "yes (" + detail + ")"
		}

		return "yes"
	}

	return "no"
}

func mdCheck(ok bool, detail string) string {
	if ok {
		if detail != "" {
			return "Yes (" + detail + ")"
		}

		return "Yes"
	}

	return "No"
}
