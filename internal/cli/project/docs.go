package project

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// checkDocs inspects documentation and config file presence.
func checkDocs(dir string) *DocsReport {
	report := &DocsReport{}

	// README
	for _, name := range []string{"README.md", "README", "README.txt", "README.rst", "readme.md"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			report.HasReadme = true
			report.ReadmeFile = name

			break
		}
	}

	// LICENSE
	for _, name := range []string{"LICENSE", "LICENSE.md", "LICENSE.txt", "LICENCE", "LICENCE.md"} {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err == nil {
			report.HasLicense = true
			report.LicenseFile = name
			report.LicenseType = detectLicenseType(path)

			break
		}
	}

	// CHANGELOG
	for _, name := range []string{"CHANGELOG.md", "CHANGELOG", "CHANGELOG.txt", "CHANGES.md", "HISTORY.md"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			report.HasChangelog = true

			break
		}
	}

	// CONTRIBUTING
	for _, name := range []string{"CONTRIBUTING.md", "CONTRIBUTING", "CONTRIBUTING.txt"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			report.HasContributing = true

			break
		}
	}

	// CLAUDE.md
	if _, err := os.Stat(filepath.Join(dir, "CLAUDE.md")); err == nil {
		report.HasClaudeMD = true
	}

	// docs/ directory
	if info, err := os.Stat(filepath.Join(dir, "docs")); err == nil && info.IsDir() {
		report.HasDocsDir = true
	}

	// doc.go files
	if _, err := os.Stat(filepath.Join(dir, "doc.go")); err == nil {
		report.HasDocGo = true
	}

	// CI configs
	report.CIConfigs = detectCIConfigs(dir)

	// .gitignore
	if _, err := os.Stat(filepath.Join(dir, ".gitignore")); err == nil {
		report.HasGitignore = true
	}

	// .editorconfig
	if _, err := os.Stat(filepath.Join(dir, ".editorconfig")); err == nil {
		report.HasEditorconfig = true
	}

	// Linter configs
	report.LinterConfigs = detectLinterConfigs(dir)

	return report
}

func detectCIConfigs(dir string) []string {
	var configs []string

	// GitHub Actions
	ghDir := filepath.Join(dir, ".github", "workflows")
	if entries, err := os.ReadDir(ghDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() && (strings.HasSuffix(e.Name(), ".yml") || strings.HasSuffix(e.Name(), ".yaml")) {
				configs = append(configs, ".github/workflows/"+e.Name())
			}
		}
	}

	// GitLab CI
	if _, err := os.Stat(filepath.Join(dir, ".gitlab-ci.yml")); err == nil {
		configs = append(configs, ".gitlab-ci.yml")
	}

	// Jenkins
	if _, err := os.Stat(filepath.Join(dir, "Jenkinsfile")); err == nil {
		configs = append(configs, "Jenkinsfile")
	}

	// CircleCI
	if _, err := os.Stat(filepath.Join(dir, ".circleci", "config.yml")); err == nil {
		configs = append(configs, ".circleci/config.yml")
	}

	// Travis CI
	if _, err := os.Stat(filepath.Join(dir, ".travis.yml")); err == nil {
		configs = append(configs, ".travis.yml")
	}

	return configs
}

func detectLinterConfigs(dir string) []string {
	checks := []string{
		".golangci.yml",
		".golangci.yaml",
		".eslintrc",
		".eslintrc.js",
		".eslintrc.json",
		".eslintrc.yml",
		".prettierrc",
		".prettierrc.json",
		".prettierrc.yml",
		"biome.json",
		".stylelintrc",
		".rubocop.yml",
		".flake8",
		"pyproject.toml",
		"rustfmt.toml",
		".clang-format",
	}

	var found []string
	for _, name := range checks {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			found = append(found, name)
		}
	}

	return found
}

func detectLicenseType(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return "Unknown"
	}

	content := strings.ToLower(string(data))

	switch {
	case strings.Contains(content, "mit license") || strings.Contains(content, "permission is hereby granted, free of charge"):
		return "MIT"
	case strings.Contains(content, "apache license") && strings.Contains(content, "version 2.0"):
		return "Apache-2.0"
	case strings.Contains(content, "bsd 3-clause") || strings.Contains(content, "redistribution and use in source and binary forms"):
		if strings.Contains(content, "neither the name") {
			return "BSD-3-Clause"
		}

		return "BSD-2-Clause"
	case strings.Contains(content, "gnu general public license") && strings.Contains(content, "version 3"):
		return "GPL-3.0"
	case strings.Contains(content, "gnu general public license") && strings.Contains(content, "version 2"):
		return "GPL-2.0"
	case strings.Contains(content, "gnu lesser general public license"):
		return "LGPL"
	case strings.Contains(content, "mozilla public license") && strings.Contains(content, "version 2.0"):
		return "MPL-2.0"
	case strings.Contains(content, "isc license"):
		return "ISC"
	case strings.Contains(content, "the unlicense"):
		return "Unlicense"
	default:
		return "Unknown"
	}
}

// RunDocs runs the documentation check subcommand.
func RunDocs(w io.Writer, args []string, opts Options) error {
	dir, err := resolvePath(args)
	if err != nil {
		return fmt.Errorf("project docs: %w", err)
	}

	report := checkDocs(dir)

	if opts.JSON {
		return formatDocsJSON(w, report)
	}

	if opts.Markdown {
		return formatDocsMarkdown(w, report)
	}

	return formatDocsText(w, report)
}
