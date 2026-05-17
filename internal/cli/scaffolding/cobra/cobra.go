package cobra

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/afero"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/internal/cli/scaffolding"
	cobratpl "github.com/inovacc/omni/internal/cli/scaffolding/cobra/templates"
)

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
	AIContext   bool   // Include aicontext command
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
func RunCobraInit(w io.Writer, fs afero.Fs, dir string, opts CobraInitOptions, genOpts scaffolding.Options) error {
	if opts.Module == "" {
		return cmderr.Wrap(cmderr.ErrInvalidInput, "scaffold: module path is required")
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
		AIContext:   opts.AIContext,
		Year:        time.Now().Year(),
	}

	// Create directory structure
	dirs := []string{
		filepath.Join(dir, "cmd", opts.AppName),
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
		if err := fs.MkdirAll(d, 0755); err != nil {
			return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("scaffold: failed to create directory %s: %v", d, err))
		}
	}

	var filesCreated []string

	// Generate cmd/{appName}/{appName}.go (combined entry point + root command)
	cmdAppDir := filepath.Join("cmd", opts.AppName)
	mainPath := filepath.Join(dir, cmdAppDir, opts.AppName+".go")
	if err := scaffolding.WriteTemplate(fs, mainPath, cobratpl.MainTemplate, tplData); err != nil {
		return fmt.Errorf("scaffold: failed to create %s/%s.go: %w", cmdAppDir, opts.AppName, err)
	}

	filesCreated = append(filesCreated, filepath.Join(cmdAppDir, opts.AppName+".go"))

	// Generate cmd/{appName}/cmd_version.go
	versionPath := filepath.Join(dir, cmdAppDir, "cmd_version.go")
	if err := scaffolding.WriteTemplate(fs, versionPath, cobratpl.VersionTemplate, tplData); err != nil {
		return fmt.Errorf("scaffold: failed to create %s/cmd_version.go: %w", cmdAppDir, err)
	}

	filesCreated = append(filesCreated, filepath.Join(cmdAppDir, "cmd_version.go"))

	// Generate cmd/{appName}/cmd_cmdtree.go (always included)
	cmdtreePath := filepath.Join(dir, cmdAppDir, "cmd_cmdtree.go")
	if err := scaffolding.WriteTemplate(fs, cmdtreePath, cobratpl.CmdtreeTemplate, tplData); err != nil {
		return fmt.Errorf("scaffold: failed to create %s/cmd_cmdtree.go: %w", cmdAppDir, err)
	}

	filesCreated = append(filesCreated, filepath.Join(cmdAppDir, "cmd_cmdtree.go"))

	// Generate cmd/{appName}/cmd_aicontext.go (when AIContext is enabled)
	if opts.AIContext {
		aicontextPath := filepath.Join(dir, cmdAppDir, "cmd_aicontext.go")
		if err := scaffolding.WriteTemplate(fs, aicontextPath, cobratpl.AIContextTemplate, tplData); err != nil {
			return fmt.Errorf("scaffold: failed to create %s/cmd_aicontext.go: %w", cmdAppDir, err)
		}

		filesCreated = append(filesCreated, filepath.Join(cmdAppDir, "cmd_aicontext.go"))
	}

	// Generate go.mod
	goModPath := filepath.Join(dir, "go.mod")
	if err := scaffolding.WriteTemplate(fs, goModPath, cobratpl.GoModTemplate, tplData); err != nil {
		return fmt.Errorf("scaffold: failed to create go.mod: %w", err)
	}

	filesCreated = append(filesCreated, "go.mod")

	// Generate config if viper is enabled (without service pattern)
	if opts.UseViper && !opts.UseService {
		configPath := filepath.Join(dir, "internal", "config", "config.go")
		if err := scaffolding.WriteTemplate(fs, configPath, cobratpl.ConfigTemplate, tplData); err != nil {
			return fmt.Errorf("scaffold: failed to create config.go: %w", err)
		}

		filesCreated = append(filesCreated, "internal/config/config.go")
	}

	// Generate service pattern files if enabled
	if opts.UseService {
		// internal/parameters/config.go
		paramsPath := filepath.Join(dir, "internal", "parameters", "config.go")
		if err := scaffolding.WriteTemplate(fs, paramsPath, cobratpl.ParametersTemplate, tplData); err != nil {
			return fmt.Errorf("scaffold: failed to create parameters/config.go: %w", err)
		}

		filesCreated = append(filesCreated, "internal/parameters/config.go")

		// internal/service/service.go
		servicePath := filepath.Join(dir, "internal", "service", "service.go")
		if err := scaffolding.WriteTemplate(fs, servicePath, cobratpl.ServiceTemplate, tplData); err != nil {
			return fmt.Errorf("scaffold: failed to create service/service.go: %w", err)
		}

		filesCreated = append(filesCreated, "internal/service/service.go")

		// cmd/{AppName}/cmd_service.go — service lifecycle command group
		svcCmdPath := filepath.Join(dir, "cmd", opts.AppName, "cmd_service.go")
		if err := scaffolding.WriteTemplate(fs, svcCmdPath, cobratpl.ServiceCmdTemplate, tplData); err != nil {
			return fmt.Errorf("scaffold: failed to create cmd/%s/cmd_service.go: %w", opts.AppName, err)
		}

		filesCreated = append(filesCreated, filepath.Join("cmd", opts.AppName, "cmd_service.go"))
	}

	// Generate LICENSE
	if opts.License != "" {
		licensePath := filepath.Join(dir, "LICENSE")
		if err := scaffolding.WriteLicense(fs, licensePath, opts.License, opts.Author); err != nil {
			return fmt.Errorf("scaffold: failed to create LICENSE: %w", err)
		}

		filesCreated = append(filesCreated, "LICENSE")
	}

	// Generate README.md
	readmePath := filepath.Join(dir, "README.md")
	if err := scaffolding.WriteTemplate(fs, readmePath, cobratpl.ReadmeTemplate, tplData); err != nil {
		return fmt.Errorf("scaffold: failed to create README.md: %w", err)
	}

	filesCreated = append(filesCreated, "README.md")

	// Generate Taskfile.yml
	taskfilePath := filepath.Join(dir, "Taskfile.yml")
	if err := scaffolding.WriteTemplate(fs, taskfilePath, cobratpl.TaskfileTemplate, tplData); err != nil {
		return fmt.Errorf("scaffold: failed to create Taskfile.yml: %w", err)
	}

	filesCreated = append(filesCreated, "Taskfile.yml")

	// Generate .gitignore
	gitignorePath := filepath.Join(dir, ".gitignore")
	if err := scaffolding.WriteTemplate(fs, gitignorePath, cobratpl.GitignoreTemplate, tplData); err != nil {
		return fmt.Errorf("scaffold: failed to create .gitignore: %w", err)
	}

	filesCreated = append(filesCreated, ".gitignore")

	// Generate .editorconfig
	editorconfigPath := filepath.Join(dir, ".editorconfig")
	if err := scaffolding.WriteTemplate(fs, editorconfigPath, cobratpl.EditorConfigTemplate, tplData); err != nil {
		return fmt.Errorf("scaffold: failed to create .editorconfig: %w", err)
	}

	filesCreated = append(filesCreated, ".editorconfig")

	// Generate full project files if requested
	if opts.Full {
		// .goreleaser.yaml
		goreleaserPath := filepath.Join(dir, ".goreleaser.yaml")
		if err := scaffolding.WriteTemplate(fs, goreleaserPath, cobratpl.GoreleaserTemplate, tplData); err != nil {
			return fmt.Errorf("scaffold: failed to create .goreleaser.yaml: %w", err)
		}

		filesCreated = append(filesCreated, ".goreleaser.yaml")

		// .golangci.yml
		golangciPath := filepath.Join(dir, ".golangci.yml")
		if err := scaffolding.WriteTemplate(fs, golangciPath, cobratpl.GolangciLintTemplate, tplData); err != nil {
			return fmt.Errorf("scaffold: failed to create .golangci.yml: %w", err)
		}

		filesCreated = append(filesCreated, ".golangci.yml")

		// tools.go
		toolsPath := filepath.Join(dir, "tools.go")
		if err := scaffolding.WriteTemplate(fs, toolsPath, cobratpl.ToolsTemplate, tplData); err != nil {
			return fmt.Errorf("scaffold: failed to create tools.go: %w", err)
		}

		filesCreated = append(filesCreated, "tools.go")

		// GitHub workflows
		buildWorkflowPath := filepath.Join(dir, ".github", "workflows", "build.yml")
		if err := scaffolding.WriteTemplate(fs, buildWorkflowPath, cobratpl.WorkflowBuildTemplate, tplData); err != nil {
			return fmt.Errorf("scaffold: failed to create build.yml: %w", err)
		}

		filesCreated = append(filesCreated, ".github/workflows/build.yml")

		testWorkflowPath := filepath.Join(dir, ".github", "workflows", "test.yml")
		if err := scaffolding.WriteTemplate(fs, testWorkflowPath, cobratpl.WorkflowTestTemplate, tplData); err != nil {
			return fmt.Errorf("scaffold: failed to create test.yml: %w", err)
		}

		filesCreated = append(filesCreated, ".github/workflows/test.yml")

		releaseWorkflowPath := filepath.Join(dir, ".github", "workflows", "release.yaml")
		if err := scaffolding.WriteTemplate(fs, releaseWorkflowPath, cobratpl.WorkflowReleaseTemplate, tplData); err != nil {
			return fmt.Errorf("scaffold: failed to create release.yaml: %w", err)
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

	_, _ = fmt.Fprintf(w, "  go run ./cmd/%s\n", opts.AppName)

	return nil
}

// AddToolsOptions configures the add-tools subcommand
type AddToolsOptions struct {
	AIContext bool // Include aicontext command
}

// AddToolsResult represents the result of adding tools
type AddToolsResult struct {
	Status       string   `json:"status"`
	FilesCreated []string `json:"files_created"`
}

// RunCobraAddTools adds cmdtree (and optionally aicontext) to an existing Cobra project
func RunCobraAddTools(w io.Writer, fs afero.Fs, dir string, opts AddToolsOptions, genOpts scaffolding.Options) error {
	// Read go.mod to get module name
	goModPath := filepath.Join(dir, "go.mod")
	goModData, err := afero.ReadFile(fs, goModPath)
	if err != nil {
		return cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("scaffold: failed to read go.mod: %v", err))
	}

	moduleName := parseModuleName(goModData)
	if moduleName == "" {
		return cmderr.Wrap(cmderr.ErrInvalidInput, "scaffold: failed to parse module name from go.mod")
	}

	appName := moduleName
	if parts := strings.Split(moduleName, "/"); len(parts) > 0 {
		appName = parts[len(parts)-1]
	}

	// Verify cmd/{appName} directory exists
	cmdDir := filepath.Join(dir, "cmd", appName)
	if _, err := fs.Stat(cmdDir); err != nil {
		return cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("scaffold: cmd/%s directory not found, is this a Cobra project?", appName))
	}

	tplData := cobratpl.TemplateData{
		Module:    moduleName,
		AppName:   appName,
		AIContext: opts.AIContext,
	}

	var filesCreated []string

	// Always generate cmd/{appName}/cmd_cmdtree.go
	cmdtreePath := filepath.Join(cmdDir, "cmd_cmdtree.go")
	if _, err := fs.Stat(cmdtreePath); err == nil {
		return cmderr.Wrap(cmderr.ErrConflict, fmt.Sprintf("scaffold: cmd/%s/cmd_cmdtree.go already exists", appName))
	}

	if err := scaffolding.WriteTemplate(fs, cmdtreePath, cobratpl.CmdtreeTemplate, tplData); err != nil {
		return fmt.Errorf("scaffold: failed to create cmd/%s/cmd_cmdtree.go: %w", appName, err)
	}

	filesCreated = append(filesCreated, filepath.Join("cmd", appName, "cmd_cmdtree.go"))

	// Optionally generate cmd/{appName}/cmd_aicontext.go
	if opts.AIContext {
		aicontextPath := filepath.Join(cmdDir, "cmd_aicontext.go")
		if _, err := fs.Stat(aicontextPath); err == nil {
			return cmderr.Wrap(cmderr.ErrConflict, fmt.Sprintf("scaffold: cmd/%s/cmd_aicontext.go already exists", appName))
		}

		if err := scaffolding.WriteTemplate(fs, aicontextPath, cobratpl.AIContextTemplate, tplData); err != nil {
			return fmt.Errorf("scaffold: failed to create cmd/%s/cmd_aicontext.go: %w", appName, err)
		}

		filesCreated = append(filesCreated, filepath.Join("cmd", appName, "cmd_aicontext.go"))
	}

	if genOpts.JSON {
		result := AddToolsResult{
			Status:       "created",
			FilesCreated: filesCreated,
		}

		return json.NewEncoder(w).Encode(result)
	}

	_, _ = fmt.Fprintln(w, "Added developer tools:")
	for _, f := range filesCreated {
		_, _ = fmt.Fprintf(w, "  - %s\n", f)
	}

	return nil
}

// parseModuleName extracts the module name from go.mod content
func parseModuleName(data []byte) string {
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}

	return ""
}

// RunCobraAdd adds a new command to an existing Cobra application
func RunCobraAdd(w io.Writer, fs afero.Fs, dir string, opts CobraAddOptions, genOpts scaffolding.Options) error {
	if opts.Name == "" {
		return cmderr.Wrap(cmderr.ErrInvalidInput, "scaffold: command name is required")
	}

	if opts.Parent == "" {
		opts.Parent = "root"
	}

	if opts.Description == "" {
		opts.Description = fmt.Sprintf("%s command", opts.Name)
	}

	// Read go.mod to get app name
	goModData, err := afero.ReadFile(fs, filepath.Join(dir, "go.mod"))
	if err != nil {
		return cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("scaffold: failed to read go.mod: %v", err))
	}

	moduleName := parseModuleName(goModData)
	if moduleName == "" {
		return cmderr.Wrap(cmderr.ErrInvalidInput, "scaffold: failed to parse module name from go.mod")
	}

	appName := moduleName
	if parts := strings.Split(moduleName, "/"); len(parts) > 0 {
		appName = parts[len(parts)-1]
	}

	// Check if cmd/{appName} directory exists
	cmdDir := filepath.Join(dir, "cmd", appName)
	if _, err := fs.Stat(cmdDir); err != nil {
		return cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("scaffold: cmd/%s directory not found, is this a Cobra project?", appName))
	}

	// Generate the command file with cmd_ prefix
	cmdPath := filepath.Join(cmdDir, "cmd_"+opts.Name+".go")
	if _, err := fs.Stat(cmdPath); err == nil {
		return cmderr.Wrap(cmderr.ErrConflict, fmt.Sprintf("scaffold: command %s already exists", opts.Name))
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

	if err := scaffolding.WriteTemplate(fs, cmdPath, cobratpl.CommandTemplate, data); err != nil {
		return fmt.Errorf("scaffold: failed to create cmd_%s.go: %w", opts.Name, err)
	}

	relPath := filepath.Join("cmd", appName, "cmd_"+opts.Name+".go")

	if genOpts.JSON {
		result := AddResult{
			Status:  "created",
			Command: opts.Name,
			Parent:  opts.Parent,
			File:    relPath,
		}

		return json.NewEncoder(w).Encode(result)
	}

	_, _ = fmt.Fprintf(w, "Created command: %s\n", opts.Name)
	_, _ = fmt.Fprintf(w, "Parent: %s\n", opts.Parent)
	_, _ = fmt.Fprintf(w, "File: %s\n", relPath)

	return nil
}
