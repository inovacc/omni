package generate

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	cobratpl "github.com/inovacc/omni/internal/cli/generate/templates/cobra"
)

// Options configures the generate command behavior
type Options struct {
	JSON bool // --json: output as JSON
}

// CobraInitOptions configures Cobra app initialization
type CobraInitOptions struct {
	Module      string // Go module path (e.g., github.com/user/myapp)
	AppName     string // Application name
	Description string // Application description
	Author      string // Author name
	License     string // License type (MIT, Apache-2.0, BSD-3)
	UseViper    bool   // Include viper for configuration
	UseService  bool   // Include service pattern with inovacc/config
	Full        bool   // Full project with goreleaser, workflows, etc.
}

// CobraAddOptions configures adding a new command
type CobraAddOptions struct {
	Name        string // Command name
	Parent      string // Parent command (default: root)
	Description string // Command description
}

// InitResult represents the result of initialization
type InitResult struct {
	Status       string   `json:"status"`
	Path         string   `json:"path"`
	Module       string   `json:"module"`
	FilesCreated []string `json:"files_created"`
}

// AddResult represents the result of adding a command
type AddResult struct {
	Status  string `json:"status"`
	Command string `json:"command"`
	Parent  string `json:"parent"`
	File    string `json:"file"`
}

// RunCobraInit initializes a new Cobra CLI application
func RunCobraInit(w io.Writer, dir string, opts CobraInitOptions, genOpts Options) error {
	if opts.Module == "" {
		return fmt.Errorf("generate: module path is required")
	}

	if opts.AppName == "" {
		// Extract app name from module path
		parts := strings.Split(opts.Module, "/")
		opts.AppName = parts[len(parts)-1]
	}

	if opts.Description == "" {
		opts.Description = fmt.Sprintf("%s is a CLI application", opts.AppName)
	}

	// Create template data
	tplData := cobratpl.TemplateData{
		Module:      opts.Module,
		AppName:     opts.AppName,
		Description: opts.Description,
		Author:      opts.Author,
		License:     opts.License,
		UseViper:    opts.UseViper,
		UseService:  opts.UseService,
		Full:        opts.Full,
		Year:        time.Now().Year(),
	}

	// Create directory structure
	dirs := []string{
		filepath.Join(dir, "cmd"),
		filepath.Join(dir, "internal"),
	}

	if opts.UseViper && !opts.UseService {
		dirs = append(dirs, filepath.Join(dir, "internal", "config"))
	}

	if opts.UseService {
		dirs = append(dirs, filepath.Join(dir, "internal", "parameters"))
		dirs = append(dirs, filepath.Join(dir, "internal", "service"))
	}

	if opts.Full {
		dirs = append(dirs, filepath.Join(dir, ".github", "workflows"))
	}

	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("generate: failed to create directory %s: %w", d, err)
		}
	}

	var filesCreated []string

	// Generate main.go
	mainPath := filepath.Join(dir, "main.go")
	if err := writeTemplate(mainPath, cobratpl.MainTemplate, tplData); err != nil {
		return fmt.Errorf("generate: failed to create main.go: %w", err)
	}

	filesCreated = append(filesCreated, "main.go")

	// Generate cmd/root.go
	rootPath := filepath.Join(dir, "cmd", "root.go")
	if err := writeTemplate(rootPath, cobratpl.RootTemplate, tplData); err != nil {
		return fmt.Errorf("generate: failed to create cmd/root.go: %w", err)
	}

	filesCreated = append(filesCreated, "cmd/root.go")

	// Generate cmd/version.go
	versionPath := filepath.Join(dir, "cmd", "version.go")
	if err := writeTemplate(versionPath, cobratpl.VersionTemplate, tplData); err != nil {
		return fmt.Errorf("generate: failed to create cmd/version.go: %w", err)
	}

	filesCreated = append(filesCreated, "cmd/version.go")

	// Generate go.mod
	goModPath := filepath.Join(dir, "go.mod")
	if err := writeTemplate(goModPath, cobratpl.GoModTemplate, tplData); err != nil {
		return fmt.Errorf("generate: failed to create go.mod: %w", err)
	}

	filesCreated = append(filesCreated, "go.mod")

	// Generate config if viper is enabled (without service pattern)
	if opts.UseViper && !opts.UseService {
		configPath := filepath.Join(dir, "internal", "config", "config.go")
		if err := writeTemplate(configPath, cobratpl.ConfigTemplate, tplData); err != nil {
			return fmt.Errorf("generate: failed to create config.go: %w", err)
		}

		filesCreated = append(filesCreated, "internal/config/config.go")
	}

	// Generate service pattern files if enabled
	if opts.UseService {
		// internal/parameters/config.go
		paramsPath := filepath.Join(dir, "internal", "parameters", "config.go")
		if err := writeTemplate(paramsPath, cobratpl.ParametersTemplate, tplData); err != nil {
			return fmt.Errorf("generate: failed to create parameters/config.go: %w", err)
		}

		filesCreated = append(filesCreated, "internal/parameters/config.go")

		// internal/service/service.go
		servicePath := filepath.Join(dir, "internal", "service", "service.go")
		if err := writeTemplate(servicePath, cobratpl.ServiceTemplate, tplData); err != nil {
			return fmt.Errorf("generate: failed to create service/service.go: %w", err)
		}

		filesCreated = append(filesCreated, "internal/service/service.go")
	}

	// Generate LICENSE
	if opts.License != "" {
		licensePath := filepath.Join(dir, "LICENSE")
		if err := writeLicense(licensePath, opts.License, opts.Author); err != nil {
			return fmt.Errorf("generate: failed to create LICENSE: %w", err)
		}

		filesCreated = append(filesCreated, "LICENSE")
	}

	// Generate README.md
	readmePath := filepath.Join(dir, "README.md")
	if err := writeTemplate(readmePath, cobratpl.ReadmeTemplate, tplData); err != nil {
		return fmt.Errorf("generate: failed to create README.md: %w", err)
	}

	filesCreated = append(filesCreated, "README.md")

	// Generate Taskfile.yml
	taskfilePath := filepath.Join(dir, "Taskfile.yml")
	if err := writeTemplate(taskfilePath, cobratpl.TaskfileTemplate, tplData); err != nil {
		return fmt.Errorf("generate: failed to create Taskfile.yml: %w", err)
	}

	filesCreated = append(filesCreated, "Taskfile.yml")

	// Generate .gitignore
	gitignorePath := filepath.Join(dir, ".gitignore")
	if err := writeTemplate(gitignorePath, cobratpl.GitignoreTemplate, tplData); err != nil {
		return fmt.Errorf("generate: failed to create .gitignore: %w", err)
	}

	filesCreated = append(filesCreated, ".gitignore")

	// Generate .editorconfig
	editorconfigPath := filepath.Join(dir, ".editorconfig")
	if err := writeTemplate(editorconfigPath, cobratpl.EditorConfigTemplate, tplData); err != nil {
		return fmt.Errorf("generate: failed to create .editorconfig: %w", err)
	}

	filesCreated = append(filesCreated, ".editorconfig")

	// Generate full project files if requested
	if opts.Full {
		// .goreleaser.yaml
		goreleaserPath := filepath.Join(dir, ".goreleaser.yaml")
		if err := writeTemplate(goreleaserPath, cobratpl.GoreleaserTemplate, tplData); err != nil {
			return fmt.Errorf("generate: failed to create .goreleaser.yaml: %w", err)
		}

		filesCreated = append(filesCreated, ".goreleaser.yaml")

		// .golangci.yml
		golangciPath := filepath.Join(dir, ".golangci.yml")
		if err := writeTemplate(golangciPath, cobratpl.GolangciLintTemplate, tplData); err != nil {
			return fmt.Errorf("generate: failed to create .golangci.yml: %w", err)
		}

		filesCreated = append(filesCreated, ".golangci.yml")

		// tools.go
		toolsPath := filepath.Join(dir, "tools.go")
		if err := writeTemplate(toolsPath, cobratpl.ToolsTemplate, tplData); err != nil {
			return fmt.Errorf("generate: failed to create tools.go: %w", err)
		}

		filesCreated = append(filesCreated, "tools.go")

		// GitHub workflows
		buildWorkflowPath := filepath.Join(dir, ".github", "workflows", "build.yml")
		if err := writeTemplate(buildWorkflowPath, cobratpl.WorkflowBuildTemplate, tplData); err != nil {
			return fmt.Errorf("generate: failed to create build.yml: %w", err)
		}

		filesCreated = append(filesCreated, ".github/workflows/build.yml")

		testWorkflowPath := filepath.Join(dir, ".github", "workflows", "test.yml")
		if err := writeTemplate(testWorkflowPath, cobratpl.WorkflowTestTemplate, tplData); err != nil {
			return fmt.Errorf("generate: failed to create test.yml: %w", err)
		}

		filesCreated = append(filesCreated, ".github/workflows/test.yml")

		releaseWorkflowPath := filepath.Join(dir, ".github", "workflows", "release.yaml")
		if err := writeTemplate(releaseWorkflowPath, cobratpl.WorkflowReleaseTemplate, tplData); err != nil {
			return fmt.Errorf("generate: failed to create release.yaml: %w", err)
		}

		filesCreated = append(filesCreated, ".github/workflows/release.yaml")
	}

	if genOpts.JSON {
		result := InitResult{
			Status:       "created",
			Path:         dir,
			Module:       opts.Module,
			FilesCreated: filesCreated,
		}

		return json.NewEncoder(w).Encode(result)
	}

	_, _ = fmt.Fprintf(w, "Created Cobra CLI application: %s\n", opts.AppName)
	_, _ = fmt.Fprintf(w, "Module: %s\n", opts.Module)

	_, _ = fmt.Fprintf(w, "Path: %s\n", dir)
	if opts.Full {
		_, _ = fmt.Fprintln(w, "Mode: Full (includes goreleaser, linting, GitHub workflows)")
	}

	if opts.UseService {
		_, _ = fmt.Fprintln(w, "Config: Service pattern (inovacc/config)")
	} else if opts.UseViper {
		_, _ = fmt.Fprintln(w, "Config: Viper")
	}

	_, _ = fmt.Fprintln(w, "\nFiles created:")
	for _, f := range filesCreated {
		_, _ = fmt.Fprintf(w, "  - %s\n", f)
	}

	_, _ = fmt.Fprintln(w, "\nNext steps:")
	_, _ = fmt.Fprintf(w, "  cd %s\n", dir)

	_, _ = fmt.Fprintln(w, "  go mod tidy")
	if opts.Full {
		_, _ = fmt.Fprintln(w, "  task build")
	} else {
		_, _ = fmt.Fprintln(w, "  go build")
	}

	_, _ = fmt.Fprintf(w, "  ./%s --help\n", opts.AppName)

	return nil
}

// RunCobraAdd adds a new command to an existing Cobra application
func RunCobraAdd(w io.Writer, dir string, opts CobraAddOptions, genOpts Options) error {
	if opts.Name == "" {
		return fmt.Errorf("generate: command name is required")
	}

	if opts.Parent == "" {
		opts.Parent = "root"
	}

	if opts.Description == "" {
		opts.Description = fmt.Sprintf("%s command", opts.Name)
	}

	// Check if cmd directory exists
	cmdDir := filepath.Join(dir, "cmd")
	if _, err := os.Stat(cmdDir); os.IsNotExist(err) {
		return fmt.Errorf("generate: cmd directory not found, is this a Cobra project?")
	}

	// Generate the command file
	cmdPath := filepath.Join(cmdDir, opts.Name+".go")
	if _, err := os.Stat(cmdPath); err == nil {
		return fmt.Errorf("generate: command %s already exists", opts.Name)
	}

	data := struct {
		Name        string
		Parent      string
		Description string
		NameTitle   string
	}{
		Name:        opts.Name,
		Parent:      opts.Parent,
		Description: opts.Description,
		NameTitle:   strings.Title(opts.Name), //nolint:staticcheck
	}

	if err := writeTemplate(cmdPath, cobratpl.CommandTemplate, data); err != nil {
		return fmt.Errorf("generate: failed to create %s.go: %w", opts.Name, err)
	}

	if genOpts.JSON {
		result := AddResult{
			Status:  "created",
			Command: opts.Name,
			Parent:  opts.Parent,
			File:    filepath.Join("cmd", opts.Name+".go"),
		}

		return json.NewEncoder(w).Encode(result)
	}

	_, _ = fmt.Fprintf(w, "Created command: %s\n", opts.Name)
	_, _ = fmt.Fprintf(w, "Parent: %s\n", opts.Parent)
	_, _ = fmt.Fprintf(w, "File: cmd/%s.go\n", opts.Name)

	return nil
}

func writeTemplate(path string, tmpl string, data any) error {
	t, err := template.New("").Parse(tmpl)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	defer func() { _ = f.Close() }()

	return t.Execute(f, data)
}

func writeLicense(path, licenseType, author string) error {
	year := time.Now().Year()

	var content string

	switch strings.ToLower(licenseType) {
	case "mit":
		content = fmt.Sprintf(cobratpl.MITLicense, year, author)
	case "apache-2.0", "apache":
		content = fmt.Sprintf(cobratpl.ApacheLicense, year, author)
	case "bsd-3", "bsd":
		content = fmt.Sprintf(cobratpl.BSDLicense, year, author)
	default:
		return fmt.Errorf("unknown license type: %s", licenseType)
	}

	return os.WriteFile(path, []byte(content), 0644)
}
