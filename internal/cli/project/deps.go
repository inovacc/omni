package project

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
)

// analyzeDeps detects and parses dependency files.
func analyzeDeps(dir string, types []ProjectType) *DepsReport {
	report := &DepsReport{}
	empty := true

	for _, t := range types {
		switch t.Language {
		case "Go":
			if d := parseGoMod(filepath.Join(dir, "go.mod")); d != nil {
				report.Go = d
				empty = false
			}
		case "JavaScript/TypeScript":
			if d := parsePackageJSON(filepath.Join(dir, "package.json")); d != nil {
				report.Node = d
				empty = false
			}
		case "Python":
			if d := parseRequirementsTxt(filepath.Join(dir, "requirements.txt")); d != nil {
				report.Python = d
				empty = false
			} else if d := parsePyprojectToml(filepath.Join(dir, "pyproject.toml")); d != nil {
				report.Python = d
				empty = false
			}
		case "Rust":
			if d := parseCargoToml(filepath.Join(dir, "Cargo.toml")); d != nil {
				report.Rust = d
				empty = false
			}
		case "Java":
			if d := parsePomXML(filepath.Join(dir, "pom.xml")); d != nil {
				report.Java = d
				empty = false
			} else if d := parseBuildGradle(filepath.Join(dir, "build.gradle")); d != nil {
				report.Java = d
				empty = false
			} else if d := parseBuildGradle(filepath.Join(dir, "build.gradle.kts")); d != nil {
				report.Java = d
				empty = false
			}
		case "Ruby":
			if d := parseGemfile(filepath.Join(dir, "Gemfile")); d != nil {
				report.Ruby = d
				empty = false
			}
		case "PHP":
			if d := parseComposerJSON(filepath.Join(dir, "composer.json")); d != nil {
				report.PHP = d
				empty = false
			}
		case "C#":
			if d := parseCsproj(dir); d != nil {
				report.DotNet = d
				empty = false
			}
		}
	}

	if empty {
		return nil
	}

	return report
}

func parseGoMod(path string) *GoDeps {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	deps := &GoDeps{}
	inRequire := false
	indirect := false

	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)

		if after, ok := strings.CutPrefix(line, "module "); ok {
			deps.Module = after
			continue
		}

		if after, ok := strings.CutPrefix(line, "go "); ok {
			deps.GoVersion = after
			continue
		}

		if line == "require (" {
			inRequire = true
			continue
		}

		if line == ")" {
			inRequire = false
			continue
		}

		if inRequire {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				indirect = strings.Contains(line, "// indirect")
				if indirect {
					deps.Indirect = append(deps.Indirect, parts[0])
				} else {
					deps.Direct = append(deps.Direct, parts[0])
				}
			}
		}

		// Single-line require
		if strings.HasPrefix(line, "require ") && !strings.HasSuffix(line, "(") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				if strings.Contains(line, "// indirect") {
					deps.Indirect = append(deps.Indirect, parts[1])
				} else {
					deps.Direct = append(deps.Direct, parts[1])
				}
			}
		}
	}

	deps.TotalCount = len(deps.Direct) + len(deps.Indirect)

	return deps
}

func parsePackageJSON(path string) *NodeDeps {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var pkg struct {
		Name            string            `json:"name"`
		Version         string            `json:"version"`
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}

	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil
	}

	deps := &NodeDeps{
		Name:    pkg.Name,
		Version: pkg.Version,
	}

	for name := range pkg.Dependencies {
		deps.Dependencies = append(deps.Dependencies, name)
	}

	for name := range pkg.DevDependencies {
		deps.DevDependencies = append(deps.DevDependencies, name)
	}

	// Detect package manager
	dir := filepath.Dir(path)
	switch {
	case fileExists(filepath.Join(dir, "pnpm-lock.yaml")):
		deps.PackageManager = "pnpm"
	case fileExists(filepath.Join(dir, "yarn.lock")):
		deps.PackageManager = "yarn"
	case fileExists(filepath.Join(dir, "bun.lockb")):
		deps.PackageManager = "bun"
	default:
		deps.PackageManager = "npm"
	}

	deps.TotalCount = len(deps.Dependencies) + len(deps.DevDependencies)

	return deps
}

func parseRequirementsTxt(path string) *PythonDeps {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	deps := &PythonDeps{Source: "requirements.txt"}

	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "-") {
			continue
		}

		// Strip version specifier for display
		name := line
		for _, sep := range []string{"==", ">=", "<=", "~=", "!=", ">"} {
			if idx := strings.Index(name, sep); idx > 0 {
				name = name[:idx]
				break
			}
		}

		deps.Dependencies = append(deps.Dependencies, strings.TrimSpace(name))
	}

	deps.TotalCount = len(deps.Dependencies)

	return deps
}

func parsePyprojectToml(path string) *PythonDeps {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var pyproject struct {
		Project struct {
			Dependencies []string `toml:"dependencies"`
		} `toml:"project"`
	}

	if err := toml.Unmarshal(data, &pyproject); err != nil {
		return nil
	}

	if len(pyproject.Project.Dependencies) == 0 {
		return nil
	}

	deps := &PythonDeps{Source: "pyproject.toml"}

	for _, d := range pyproject.Project.Dependencies {
		// Strip version specifier
		name := d
		for _, sep := range []string{"==", ">=", "<=", "~=", "!=", ">", "<", "["} {
			if idx := strings.Index(name, sep); idx > 0 {
				name = name[:idx]
				break
			}
		}

		deps.Dependencies = append(deps.Dependencies, strings.TrimSpace(name))
	}

	deps.TotalCount = len(deps.Dependencies)

	return deps
}

func parseCargoToml(path string) *RustDeps {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var cargo struct {
		Package struct {
			Name    string `toml:"name"`
			Version string `toml:"version"`
			Edition string `toml:"edition"`
		} `toml:"package"`
		Dependencies map[string]any `toml:"dependencies"`
	}

	if err := toml.Unmarshal(data, &cargo); err != nil {
		return nil
	}

	deps := &RustDeps{
		Name:    cargo.Package.Name,
		Version: cargo.Package.Version,
		Edition: cargo.Package.Edition,
	}

	for name := range cargo.Dependencies {
		deps.Dependencies = append(deps.Dependencies, name)
	}

	deps.TotalCount = len(deps.Dependencies)

	return deps
}

func parsePomXML(path string) *JavaDeps {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var pom struct {
		XMLName      xml.Name `xml:"project"`
		GroupID      string   `xml:"groupId"`
		ArtifactID   string   `xml:"artifactId"`
		Dependencies struct {
			Dependency []struct {
				GroupID    string `xml:"groupId"`
				ArtifactID string `xml:"artifactId"`
			} `xml:"dependency"`
		} `xml:"dependencies"`
	}

	if err := xml.Unmarshal(data, &pom); err != nil {
		return nil
	}

	deps := &JavaDeps{
		Source:     "pom.xml",
		GroupID:    pom.GroupID,
		ArtifactID: pom.ArtifactID,
	}

	for _, d := range pom.Dependencies.Dependency {
		deps.Dependencies = append(deps.Dependencies, d.GroupID+":"+d.ArtifactID)
	}

	deps.TotalCount = len(deps.Dependencies)

	return deps
}

var gradleDepRegex = regexp.MustCompile(`(?:implementation|api|compileOnly|runtimeOnly|testImplementation)\s*[\('"]\s*([^'")\s]+)`)

func parseBuildGradle(path string) *JavaDeps {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	deps := &JavaDeps{Source: filepath.Base(path)}

	matches := gradleDepRegex.FindAllSubmatch(data, -1)
	for _, m := range matches {
		if len(m) >= 2 {
			deps.Dependencies = append(deps.Dependencies, string(m[1]))
		}
	}

	deps.TotalCount = len(deps.Dependencies)

	if deps.TotalCount == 0 {
		return nil
	}

	return deps
}

var gemRegex = regexp.MustCompile(`^\s*gem\s+['"]([^'"]+)['"]`)

func parseGemfile(path string) *RubyDeps {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	deps := &RubyDeps{}

	for line := range strings.SplitSeq(string(data), "\n") {
		if m := gemRegex.FindStringSubmatch(line); len(m) >= 2 {
			deps.Dependencies = append(deps.Dependencies, m[1])
		}
	}

	deps.TotalCount = len(deps.Dependencies)

	if deps.TotalCount == 0 {
		return nil
	}

	return deps
}

func parseComposerJSON(path string) *PHPDeps {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var composer struct {
		Name    string            `json:"name"`
		Require map[string]string `json:"require"`
	}

	if err := json.Unmarshal(data, &composer); err != nil {
		return nil
	}

	deps := &PHPDeps{Name: composer.Name}

	for name := range composer.Require {
		if name != "php" && !strings.HasPrefix(name, "ext-") {
			deps.Dependencies = append(deps.Dependencies, name)
		}
	}

	deps.TotalCount = len(deps.Dependencies)

	return deps
}

func parseCsproj(dir string) *DotNetDeps {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".csproj") {
			return parseCsprojFile(filepath.Join(dir, e.Name()))
		}
	}

	return nil
}

func parseCsprojFile(path string) *DotNetDeps {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var proj struct {
		XMLName   xml.Name `xml:"Project"`
		ItemGroup []struct {
			PackageReference []struct {
				Include string `xml:"Include,attr"`
			} `xml:"PackageReference"`
		} `xml:"ItemGroup"`
	}

	if err := xml.Unmarshal(data, &proj); err != nil {
		return nil
	}

	deps := &DotNetDeps{}

	for _, group := range proj.ItemGroup {
		for _, ref := range group.PackageReference {
			if ref.Include != "" {
				deps.Dependencies = append(deps.Dependencies, ref.Include)
			}
		}
	}

	deps.TotalCount = len(deps.Dependencies)

	return deps
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// RunDeps runs the dependency analysis subcommand.
func RunDeps(w io.Writer, args []string, opts Options) error {
	dir, err := resolvePath(args)
	if err != nil {
		return fmt.Errorf("project deps: %w", err)
	}

	types := detectProjectTypes(dir)
	report := analyzeDeps(dir, types)

	if report == nil {
		_, _ = fmt.Fprintln(w, "No dependency files found.")
		return nil
	}

	if opts.JSON {
		return formatDepsJSON(w, report)
	}

	if opts.Markdown {
		return formatDepsMarkdown(w, report)
	}

	return formatDepsText(w, report)
}
