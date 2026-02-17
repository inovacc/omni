package repo

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

func formatAnalyze(w io.Writer, report *AnalyzeReport, opts Options) error {
	if opts.JSON {
		return formatAnalyzeJSON(w, report)
	}

	return formatAnalyzeMarkdown(w, report, opts)
}

func formatAnalyzeJSON(w io.Writer, report *AnalyzeReport) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	return enc.Encode(report)
}

func formatAnalyzeMarkdown(w io.Writer, r *AnalyzeReport, opts Options) error {
	_, _ = fmt.Fprintf(w, "# Repository Context: %s\n\n", r.Name)

	// Overview
	if opts.wantSection("overview") || opts.Sections == "" {
		writeOverview(w, r)
	}

	// Tree
	if opts.wantSection("tree") && r.Tree != "" {
		_, _ = fmt.Fprint(w, "## Directory Tree\n\n")
		_, _ = fmt.Fprintln(w, "```")
		_, _ = fmt.Fprint(w, r.Tree)
		_, _ = fmt.Fprint(w, "```\n\n")
	}

	// Key files
	if opts.wantSection("keys") && len(r.KeyFiles) > 0 {
		_, _ = fmt.Fprint(w, "## Key Files\n\n")

		for _, kf := range r.KeyFiles {
			_, _ = fmt.Fprintf(w, "### %s\n\n", kf.Name)
			_, _ = fmt.Fprintln(w, "```")
			_, _ = fmt.Fprintln(w, kf.Content)
			_, _ = fmt.Fprint(w, "```\n\n")
		}
	}

	// Dependencies
	if opts.wantSection("deps") && r.Deps != nil {
		_, _ = fmt.Fprint(w, "## Dependencies\n\n")
		writeDeps(w, r)
	}

	// Entry points
	if len(r.EntryPoints) > 0 {
		_, _ = fmt.Fprint(w, "## Entry Points\n\n")

		for _, ep := range r.EntryPoints {
			desc := ""
			if ep.Description != "" {
				desc = " — " + ep.Description
			}

			_, _ = fmt.Fprintf(w, "- `%s`%s\n", ep.Path, desc)
		}

		_, _ = fmt.Fprintln(w)
	}

	// Architecture
	if r.Architecture != nil {
		_, _ = fmt.Fprint(w, "## Architecture\n\n")
		_, _ = fmt.Fprintf(w, "- **Pattern:** %s\n", r.Architecture.Pattern)
		_, _ = fmt.Fprintf(w, "- **Top-level packages:** %d\n", r.Architecture.PackageCount)

		if len(r.Architecture.Dirs) > 0 && !opts.Compact {
			_, _ = fmt.Fprintf(w, "- **Directories:** %s\n", strings.Join(r.Architecture.Dirs, ", "))
		}

		_, _ = fmt.Fprintln(w)
	}

	// API surface
	if opts.wantSection("api") && len(r.APISurface) > 0 {
		_, _ = fmt.Fprint(w, "## API Surface\n\n")
		_, _ = fmt.Fprintln(w, "| Package | Exported Funcs |")
		_, _ = fmt.Fprintln(w, "|---------|---------------|")

		for _, api := range r.APISurface {
			_, _ = fmt.Fprintf(w, "| %s | %d |\n", api.Package, api.ExportedFuncs)
		}

		_, _ = fmt.Fprintln(w)
	}

	// Git
	if opts.wantSection("git") && r.Git != nil && r.Git.IsRepo {
		_, _ = fmt.Fprint(w, "## Git\n\n")
		_, _ = fmt.Fprintf(w, "- **Branch:** %s\n", r.Git.Branch)

		if r.Git.RemoteURL != "" {
			_, _ = fmt.Fprintf(w, "- **Remote:** %s\n", r.Git.RemoteURL)
		}

		_, _ = fmt.Fprintf(w, "- **Total commits:** %d\n", r.Git.TotalCommits)
		_, _ = fmt.Fprintf(w, "- **Contributors:** %d\n", r.Git.Contributors)

		if len(r.Git.RecentCommits) > 0 {
			_, _ = fmt.Fprint(w, "\n### Recent Commits\n\n")

			for _, c := range r.Git.RecentCommits {
				_, _ = fmt.Fprintf(w, "- `%s`\n", c)
			}
		}

		_, _ = fmt.Fprintln(w)
	}

	// Test patterns
	if opts.wantSection("tests") && len(r.TestPatterns) > 0 {
		_, _ = fmt.Fprint(w, "## Test Patterns\n\n")

		for _, p := range r.TestPatterns {
			_, _ = fmt.Fprintf(w, "- %s\n", p)
		}

		_, _ = fmt.Fprintln(w)
	}

	// CI/CD
	if opts.wantSection("ci") && len(r.CIConfigs) > 0 {
		_, _ = fmt.Fprint(w, "## CI/CD\n\n")

		for _, c := range r.CIConfigs {
			_, _ = fmt.Fprintf(w, "- `%s`\n", c)
		}

		_, _ = fmt.Fprintln(w)
	}

	// Config files
	if opts.wantSection("ci") && len(r.ConfigFiles) > 0 {
		_, _ = fmt.Fprint(w, "## Config Files\n\n")

		for _, c := range r.ConfigFiles {
			_, _ = fmt.Fprintf(w, "- `%s`\n", c)
		}

		_, _ = fmt.Fprintln(w)
	}

	return nil
}

func writeOverview(w io.Writer, r *AnalyzeReport) {
	_, _ = fmt.Fprint(w, "## Overview\n\n")

	// Languages
	if len(r.Types) > 0 {
		var langs []string
		for _, t := range r.Types {
			s := t.Language
			if len(t.Frameworks) > 0 {
				s += " (" + strings.Join(t.Frameworks, ", ") + ")"
			}

			langs = append(langs, s)
		}

		_, _ = fmt.Fprintf(w, "- **Languages:** %s\n", strings.Join(langs, ", "))
	}

	// Build tools
	if len(r.BuildTools) > 0 {
		_, _ = fmt.Fprintf(w, "- **Build tools:** %s\n", strings.Join(r.BuildTools, ", "))
	}

	// Health
	if r.Health != nil {
		_, _ = fmt.Fprintf(w, "- **Health:** %d/100 (Grade: %s)\n", r.Health.Score, r.Health.Grade)
	}

	_, _ = fmt.Fprintln(w)
}

func writeDeps(w io.Writer, r *AnalyzeReport) {
	d := r.Deps

	if d.Go != nil {
		_, _ = fmt.Fprintf(w, "### Go (`%s`, go %s)\n\n", d.Go.Module, d.Go.GoVersion)
		_, _ = fmt.Fprintf(w, "- Direct: %d, Indirect: %d, Total: %d\n", len(d.Go.Direct), len(d.Go.Indirect), d.Go.TotalCount)

		if len(d.Go.Direct) > 0 {
			_, _ = fmt.Fprintln(w, "- **Key deps:**", strings.Join(d.Go.Direct[:min(10, len(d.Go.Direct))], ", "))
		}

		_, _ = fmt.Fprintln(w)
	}

	if d.Node != nil {
		_, _ = fmt.Fprintf(w, "### Node.js (`%s@%s`, %s)\n\n", d.Node.Name, d.Node.Version, d.Node.PackageManager)
		_, _ = fmt.Fprintf(w, "- Dependencies: %d, Dev: %d\n\n", len(d.Node.Dependencies), len(d.Node.DevDependencies))
	}

	if d.Python != nil {
		_, _ = fmt.Fprintf(w, "### Python (%s) — %d deps\n\n", d.Python.Source, d.Python.TotalCount)
	}

	if d.Rust != nil {
		_, _ = fmt.Fprintf(w, "### Rust (`%s@%s`) — %d deps\n\n", d.Rust.Name, d.Rust.Version, d.Rust.TotalCount)
	}

	if d.Java != nil {
		_, _ = fmt.Fprintf(w, "### Java (%s) — %d deps\n\n", d.Java.Source, d.Java.TotalCount)
	}

	if d.Ruby != nil {
		_, _ = fmt.Fprintf(w, "### Ruby — %d gems\n\n", d.Ruby.TotalCount)
	}

	if d.PHP != nil {
		_, _ = fmt.Fprintf(w, "### PHP (`%s`) — %d deps\n\n", d.PHP.Name, d.PHP.TotalCount)
	}

	if d.DotNet != nil {
		_, _ = fmt.Fprintf(w, "### .NET — %d packages\n\n", d.DotNet.TotalCount)
	}
}
