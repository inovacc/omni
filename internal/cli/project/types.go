package project

// ProjectReport is the top-level analysis result.
type ProjectReport struct {
	Path       string         `json:"path"`
	Name       string         `json:"name"`
	Types      []ProjectType  `json:"types"`
	Languages  []LanguageInfo `json:"languages"`
	BuildTools []string       `json:"build_tools"`
	Deps       *DepsReport    `json:"deps,omitempty"`
	Git        *GitReport     `json:"git,omitempty"`
	Docs       *DocsReport    `json:"docs,omitempty"`
	Health     *HealthReport  `json:"health,omitempty"`
}

// ProjectType describes a detected project type.
type ProjectType struct {
	Language   string   `json:"language"`
	BuildFile  string   `json:"build_file"`
	Frameworks []string `json:"frameworks,omitempty"`
}

// LanguageInfo holds file count and extensions for a detected language.
type LanguageInfo struct {
	Name       string   `json:"name"`
	FileCount  int      `json:"file_count"`
	Extensions []string `json:"extensions"`
}

// DepsReport aggregates dependency info across ecosystems.
type DepsReport struct {
	Go       *GoDeps     `json:"go,omitempty"`
	Node     *NodeDeps   `json:"node,omitempty"`
	Python   *PythonDeps `json:"python,omitempty"`
	Rust     *RustDeps   `json:"rust,omitempty"`
	Java     *JavaDeps   `json:"java,omitempty"`
	Ruby     *RubyDeps   `json:"ruby,omitempty"`
	PHP      *PHPDeps    `json:"php,omitempty"`
	DotNet   *DotNetDeps `json:"dotnet,omitempty"`
}

// GoDeps holds Go module dependency info.
type GoDeps struct {
	Module     string   `json:"module"`
	GoVersion  string   `json:"go_version"`
	Direct     []string `json:"direct"`
	Indirect   []string `json:"indirect"`
	TotalCount int      `json:"total_count"`
}

// NodeDeps holds Node.js dependency info.
type NodeDeps struct {
	Name            string   `json:"name"`
	Version         string   `json:"version"`
	PackageManager  string   `json:"package_manager"`
	Dependencies    []string `json:"dependencies"`
	DevDependencies []string `json:"dev_dependencies"`
	TotalCount      int      `json:"total_count"`
}

// PythonDeps holds Python dependency info.
type PythonDeps struct {
	Source       string   `json:"source"`
	Dependencies []string `json:"dependencies"`
	TotalCount   int      `json:"total_count"`
}

// RustDeps holds Rust/Cargo dependency info.
type RustDeps struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Edition      string   `json:"edition,omitempty"`
	Dependencies []string `json:"dependencies"`
	TotalCount   int      `json:"total_count"`
}

// JavaDeps holds Java dependency info.
type JavaDeps struct {
	Source       string   `json:"source"`
	GroupID      string   `json:"group_id,omitempty"`
	ArtifactID   string   `json:"artifact_id,omitempty"`
	Dependencies []string `json:"dependencies"`
	TotalCount   int      `json:"total_count"`
}

// RubyDeps holds Ruby/Bundler dependency info.
type RubyDeps struct {
	Dependencies []string `json:"dependencies"`
	TotalCount   int      `json:"total_count"`
}

// PHPDeps holds PHP/Composer dependency info.
type PHPDeps struct {
	Name         string   `json:"name"`
	Dependencies []string `json:"dependencies"`
	TotalCount   int      `json:"total_count"`
}

// DotNetDeps holds .NET/C# dependency info.
type DotNetDeps struct {
	Dependencies []string `json:"dependencies"`
	TotalCount   int      `json:"total_count"`
}

// GitReport holds git repository info.
type GitReport struct {
	IsRepo        bool     `json:"is_repo"`
	Branch        string   `json:"branch,omitempty"`
	Remote        string   `json:"remote,omitempty"`
	RemoteURL     string   `json:"remote_url,omitempty"`
	Clean         bool     `json:"clean"`
	Ahead         int      `json:"ahead,omitempty"`
	Behind        int      `json:"behind,omitempty"`
	RecentCommits []string `json:"recent_commits,omitempty"`
	Tags          []string `json:"tags,omitempty"`
	TotalCommits  int      `json:"total_commits,omitempty"`
	Contributors  int      `json:"contributors,omitempty"`
}

// DocsReport holds documentation status info.
type DocsReport struct {
	HasReadme      bool     `json:"has_readme"`
	ReadmeFile     string   `json:"readme_file,omitempty"`
	HasLicense     bool     `json:"has_license"`
	LicenseType    string   `json:"license_type,omitempty"`
	LicenseFile    string   `json:"license_file,omitempty"`
	HasChangelog   bool     `json:"has_changelog"`
	HasContributing bool   `json:"has_contributing"`
	HasClaudeMD    bool     `json:"has_claude_md"`
	HasDocsDir     bool     `json:"has_docs_dir"`
	HasDocGo       bool     `json:"has_doc_go"`
	CIConfigs      []string `json:"ci_configs,omitempty"`
	HasGitignore   bool     `json:"has_gitignore"`
	HasEditorconfig bool   `json:"has_editorconfig"`
	LinterConfigs  []string `json:"linter_configs,omitempty"`
}

// HealthReport holds a project health score.
type HealthReport struct {
	Score  int           `json:"score"`
	Grade  string        `json:"grade"`
	Checks []HealthCheck `json:"checks"`
}

// HealthCheck is a single health check result.
type HealthCheck struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Points  int    `json:"points"`
	MaxPts  int    `json:"max_points"`
	Details string `json:"details,omitempty"`
}
