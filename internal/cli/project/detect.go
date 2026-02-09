package project

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// markerFile maps a build/config file to its language.
type markerFile struct {
	file     string
	language string
}

var markers = []markerFile{
	{"go.mod", "Go"},
	{"package.json", "JavaScript/TypeScript"},
	{"Cargo.toml", "Rust"},
	{"pom.xml", "Java"},
	{"build.gradle", "Java"},
	{"build.gradle.kts", "Kotlin"},
	{"requirements.txt", "Python"},
	{"pyproject.toml", "Python"},
	{"setup.py", "Python"},
	{"Gemfile", "Ruby"},
	{"composer.json", "PHP"},
	{"CMakeLists.txt", "C/C++"},
	{"Makefile", "C/C++"},
	{"mix.exs", "Elixir"},
	{"stack.yaml", "Haskell"},
}

// detectProjectTypes checks for known marker files and returns detected project types.
func detectProjectTypes(dir string) []ProjectType {
	var types []ProjectType
	seen := make(map[string]bool)

	for _, m := range markers {
		path := filepath.Join(dir, m.file)
		if _, err := os.Stat(path); err == nil {
			if !seen[m.language] {
				types = append(types, ProjectType{
					Language:  m.language,
					BuildFile: m.file,
				})
				seen[m.language] = true
			}
		}
	}

	// Check for *.csproj files
	entries, err := os.ReadDir(dir)
	if err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".csproj") {
				if !seen["C#"] {
					types = append(types, ProjectType{
						Language:  "C#",
						BuildFile: e.Name(),
					})
					seen["C#"] = true
				}

				break
			}
		}
	}

	return types
}

// frameworkMarker maps a dependency name to a framework name.
type frameworkMarker struct {
	dep       string
	framework string
}

var goFrameworks = []frameworkMarker{
	{"github.com/spf13/cobra", "Cobra"},
	{"github.com/go-chi/chi", "Chi"},
	{"github.com/gin-gonic/gin", "Gin"},
	{"github.com/labstack/echo", "Echo"},
	{"github.com/gofiber/fiber", "Fiber"},
	{"google.golang.org/grpc", "gRPC"},
	{"gorm.io/gorm", "GORM"},
	{"entgo.io/ent", "Ent"},
	{"github.com/charmbracelet/bubbletea", "Bubbletea"},
}

var nodeFrameworks = []frameworkMarker{
	{"react", "React"},
	{"next", "Next.js"},
	{"vue", "Vue"},
	{"nuxt", "Nuxt"},
	{"svelte", "Svelte"},
	{"express", "Express"},
	{"fastify", "Fastify"},
	{"nestjs", "NestJS"},
	{"@nestjs/core", "NestJS"},
	{"angular", "Angular"},
	{"@angular/core", "Angular"},
}

// detectFrameworks checks dependencies for known framework names.
func detectFrameworks(dir string, types []ProjectType, deps *DepsReport) {
	if deps == nil {
		return
	}

	for i := range types {
		switch types[i].Language {
		case "Go":
			if deps.Go != nil {
				types[i].Frameworks = matchFrameworks(deps.Go.Direct, goFrameworks)
			}
		case "JavaScript/TypeScript":
			if deps.Node != nil {
				all := append(deps.Node.Dependencies, deps.Node.DevDependencies...)
				types[i].Frameworks = matchFrameworks(all, nodeFrameworks)
			}
		}
	}
}

func matchFrameworks(deps []string, markers []frameworkMarker) []string {
	depSet := make(map[string]bool, len(deps))
	for _, d := range deps {
		depSet[d] = true
	}

	var found []string
	seen := make(map[string]bool)

	for _, m := range markers {
		if depSet[m.dep] && !seen[m.framework] {
			found = append(found, m.framework)
			seen[m.framework] = true
		}
	}

	return found
}

// extension-to-language mapping for counting.
var extLang = map[string]string{
	".go":    "Go",
	".js":    "JavaScript",
	".mjs":   "JavaScript",
	".cjs":   "JavaScript",
	".jsx":   "JavaScript",
	".ts":    "TypeScript",
	".tsx":   "TypeScript",
	".py":    "Python",
	".pyw":   "Python",
	".rs":    "Rust",
	".java":  "Java",
	".kt":    "Kotlin",
	".kts":   "Kotlin",
	".scala": "Scala",
	".c":     "C",
	".h":     "C",
	".cpp":   "C++",
	".cc":    "C++",
	".cxx":   "C++",
	".hpp":   "C++",
	".cs":    "C#",
	".rb":    "Ruby",
	".php":   "PHP",
	".swift": "Swift",
	".sh":    "Shell",
	".bash":  "Shell",
	".zsh":   "Shell",
	".lua":   "Lua",
	".ex":    "Elixir",
	".exs":   "Elixir",
	".erl":   "Erlang",
	".hs":    "Haskell",
	".ml":    "OCaml",
	".fs":    "F#",
	".r":     "R",
	".R":     "R",
	".sql":   "SQL",
	".html":  "HTML",
	".htm":   "HTML",
	".css":   "CSS",
	".scss":  "SCSS",
	".yaml":  "YAML",
	".yml":   "YAML",
	".toml":  "TOML",
	".json":  "JSON",
	".xml":   "XML",
	".md":    "Markdown",
	".proto": "Protobuf",
	".tf":    "Terraform",
	".zig":   "Zig",
	".nim":   "Nim",
	".dart":  "Dart",
}

// directories to skip when walking.
var skipDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"vendor":       true,
	"__pycache__":  true,
	".idea":        true,
	".vscode":      true,
	"target":       true,
	"build":        true,
	"dist":         true,
	"bin":          true,
	".next":        true,
	".nuxt":        true,
}

// countLanguages walks the directory counting files by language.
func countLanguages(dir string) []LanguageInfo {
	type langData struct {
		count int
		exts  map[string]bool
	}

	langs := make(map[string]*langData)

	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil //nolint:nilerr // skip inaccessible
		}

		name := d.Name()

		if d.IsDir() {
			if skipDirs[name] || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}

			return nil
		}

		ext := filepath.Ext(name)
		lang, ok := extLang[ext]
		if !ok {
			return nil
		}

		if langs[lang] == nil {
			langs[lang] = &langData{exts: make(map[string]bool)}
		}

		langs[lang].count++
		langs[lang].exts[ext] = true

		return nil
	})

	result := make([]LanguageInfo, 0, len(langs))
	for name, data := range langs {
		exts := make([]string, 0, len(data.exts))
		for e := range data.exts {
			exts = append(exts, e)
		}

		sort.Strings(exts)

		result = append(result, LanguageInfo{
			Name:       name,
			FileCount:  data.count,
			Extensions: exts,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].FileCount > result[j].FileCount
	})

	return result
}

// detectBuildTools checks for common build/CI tool config files.
func detectBuildTools(dir string) []string {
	checks := []struct {
		path string
		name string
	}{
		{"Taskfile.yml", "Taskfile"},
		{"Taskfile.yaml", "Taskfile"},
		{"Makefile", "Make"},
		{"Dockerfile", "Docker"},
		{"docker-compose.yml", "Docker Compose"},
		{"docker-compose.yaml", "Docker Compose"},
		{".goreleaser.yml", "GoReleaser"},
		{".goreleaser.yaml", "GoReleaser"},
		{"Jenkinsfile", "Jenkins"},
		{"Vagrantfile", "Vagrant"},
		{"Procfile", "Heroku"},
		{"fly.toml", "Fly.io"},
		{"vercel.json", "Vercel"},
		{"netlify.toml", "Netlify"},
	}

	var found []string
	seen := make(map[string]bool)

	for _, c := range checks {
		if _, err := os.Stat(filepath.Join(dir, c.path)); err == nil {
			if !seen[c.name] {
				found = append(found, c.name)
				seen[c.name] = true
			}
		}
	}

	return found
}
