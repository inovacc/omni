package generate

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	cobratpl "github.com/inovacc/omni/internal/cli/generate/templates/cobra"
	handlertpl "github.com/inovacc/omni/internal/cli/generate/templates/handler"
	repotpl "github.com/inovacc/omni/internal/cli/generate/templates/repository"
	testtpl "github.com/inovacc/omni/internal/cli/generate/templates/test"
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

// HandlerOptions configures handler generation
type HandlerOptions struct {
	Package    string   // Package name (default: "handler")
	Dir        string   // Output directory (default: "internal/handler")
	Methods    []string // HTTP methods: GET, POST, PUT, DELETE, PATCH
	Path       string   // URL path pattern
	Middleware bool     // Include middleware support
	Framework  string   // chi, gin, echo, stdlib (default: stdlib)
}

// HandlerResult represents the result of handler generation
type HandlerResult struct {
	Status       string   `json:"status"`
	Name         string   `json:"name"`
	Package      string   `json:"package"`
	Framework    string   `json:"framework"`
	FilesCreated []string `json:"files_created"`
}

// RunHandlerInit generates a new handler
func RunHandlerInit(w io.Writer, name string, opts HandlerOptions, genOpts Options) error {
	if name == "" {
		return fmt.Errorf("generate: handler name is required")
	}

	// Set defaults
	if opts.Package == "" {
		opts.Package = "handler"
	}

	if opts.Dir == "" {
		opts.Dir = filepath.Join("internal", "handler")
	}

	if len(opts.Methods) == 0 {
		opts.Methods = []string{"GET", "POST", "PUT", "DELETE"}
	}

	if opts.Framework == "" {
		opts.Framework = "stdlib"
	}

	// Normalize methods to uppercase
	for i, m := range opts.Methods {
		opts.Methods[i] = strings.ToUpper(m)
	}

	// Create directory
	if err := os.MkdirAll(opts.Dir, 0755); err != nil {
		return fmt.Errorf("generate: failed to create directory %s: %w", opts.Dir, err)
	}

	// Prepare template data
	data := handlertpl.TemplateData{
		Name:       strings.Title(name), //nolint:staticcheck
		NameLower:  strings.ToLower(name),
		Package:    opts.Package,
		Methods:    opts.Methods,
		Path:       opts.Path,
		Middleware: opts.Middleware,
		Framework:  opts.Framework,
	}

	var filesCreated []string

	// Select template based on framework
	var tpl string

	switch strings.ToLower(opts.Framework) {
	case "chi":
		tpl = handlertpl.ChiHandlerTemplate
	case "gin":
		tpl = handlertpl.GinHandlerTemplate
	case "echo":
		tpl = handlertpl.EchoHandlerTemplate
	default:
		tpl = handlertpl.StdlibHandlerTemplate
	}

	// Generate handler file
	handlerPath := filepath.Join(opts.Dir, strings.ToLower(name)+".go")

	if err := writeTemplate(handlerPath, tpl, data); err != nil {
		return fmt.Errorf("generate: failed to create %s: %w", handlerPath, err)
	}

	filesCreated = append(filesCreated, handlerPath)

	// Generate test file (only for stdlib for now)
	if opts.Framework == "stdlib" {
		testPath := filepath.Join(opts.Dir, strings.ToLower(name)+"_test.go")

		if err := writeTemplate(testPath, handlertpl.HandlerTestTemplate, data); err != nil {
			return fmt.Errorf("generate: failed to create %s: %w", testPath, err)
		}

		filesCreated = append(filesCreated, testPath)
	}

	if genOpts.JSON {
		result := HandlerResult{
			Status:       "created",
			Name:         name,
			Package:      opts.Package,
			Framework:    opts.Framework,
			FilesCreated: filesCreated,
		}

		return json.NewEncoder(w).Encode(result)
	}

	_, _ = fmt.Fprintf(w, "Created handler: %s\n", name)
	_, _ = fmt.Fprintf(w, "Framework: %s\n", opts.Framework)
	_, _ = fmt.Fprintf(w, "Methods: %s\n", strings.Join(opts.Methods, ", "))

	_, _ = fmt.Fprintln(w, "\nFiles created:")

	for _, f := range filesCreated {
		_, _ = fmt.Fprintf(w, "  - %s\n", f)
	}

	return nil
}

// RepositoryOptions configures repository generation
type RepositoryOptions struct {
	Package   string // Package name (default: "repository")
	Dir       string // Output directory (default: "internal/repository")
	Entity    string // Entity struct name
	Table     string // Database table name
	DB        string // Database type: postgres, mysql, sqlite (default: postgres)
	Interface bool   // Generate interface (default: true)
}

// RepositoryResult represents the result of repository generation
type RepositoryResult struct {
	Status       string   `json:"status"`
	Name         string   `json:"name"`
	Entity       string   `json:"entity"`
	Table        string   `json:"table"`
	DB           string   `json:"db"`
	FilesCreated []string `json:"files_created"`
}

// RunRepositoryInit generates a new repository
func RunRepositoryInit(w io.Writer, name string, opts RepositoryOptions, genOpts Options) error {
	if name == "" {
		return fmt.Errorf("generate: repository name is required")
	}

	// Set defaults
	if opts.Package == "" {
		opts.Package = "repository"
	}

	if opts.Dir == "" {
		opts.Dir = filepath.Join("internal", "repository")
	}

	if opts.Entity == "" {
		opts.Entity = strings.Title(name) //nolint:staticcheck
	}

	if opts.Table == "" {
		opts.Table = strings.ToLower(name) + "s"
	}

	if opts.DB == "" {
		opts.DB = "postgres"
	}

	// Create directory
	if err := os.MkdirAll(opts.Dir, 0755); err != nil {
		return fmt.Errorf("generate: failed to create directory %s: %w", opts.Dir, err)
	}

	// Prepare template data
	data := repotpl.TemplateData{
		Name:      strings.Title(name), //nolint:staticcheck
		NameLower: strings.ToLower(name),
		Package:   opts.Package,
		Entity:    opts.Entity,
		Table:     opts.Table,
		DB:        opts.DB,
		IDType:    "int64",
		Interface: opts.Interface,
	}

	var filesCreated []string

	// Generate interface file if requested
	if opts.Interface {
		interfacePath := filepath.Join(opts.Dir, "interface.go")

		if err := writeTemplate(interfacePath, repotpl.InterfaceTemplate, data); err != nil {
			return fmt.Errorf("generate: failed to create %s: %w", interfacePath, err)
		}

		filesCreated = append(filesCreated, interfacePath)
	}

	// Select template based on DB
	var tpl string

	switch strings.ToLower(opts.DB) {
	case "mysql":
		tpl = repotpl.MySQLRepositoryTemplate
	case "sqlite":
		tpl = repotpl.SQLiteRepositoryTemplate
	default:
		tpl = repotpl.PostgresRepositoryTemplate
	}

	// Generate repository file
	repoPath := filepath.Join(opts.Dir, strings.ToLower(name)+".go")

	if err := writeTemplate(repoPath, tpl, data); err != nil {
		return fmt.Errorf("generate: failed to create %s: %w", repoPath, err)
	}

	filesCreated = append(filesCreated, repoPath)

	// Generate test file
	testPath := filepath.Join(opts.Dir, strings.ToLower(name)+"_test.go")

	if err := writeTemplate(testPath, repotpl.RepositoryTestTemplate, data); err != nil {
		return fmt.Errorf("generate: failed to create %s: %w", testPath, err)
	}

	filesCreated = append(filesCreated, testPath)

	if genOpts.JSON {
		result := RepositoryResult{
			Status:       "created",
			Name:         name,
			Entity:       opts.Entity,
			Table:        opts.Table,
			DB:           opts.DB,
			FilesCreated: filesCreated,
		}

		return json.NewEncoder(w).Encode(result)
	}

	_, _ = fmt.Fprintf(w, "Created repository: %s\n", name)
	_, _ = fmt.Fprintf(w, "Entity: %s\n", opts.Entity)
	_, _ = fmt.Fprintf(w, "Table: %s\n", opts.Table)
	_, _ = fmt.Fprintf(w, "Database: %s\n", opts.DB)

	_, _ = fmt.Fprintln(w, "\nFiles created:")

	for _, f := range filesCreated {
		_, _ = fmt.Fprintf(w, "  - %s\n", f)
	}

	return nil
}

// TestOptions configures test generation
type TestOptions struct {
	Table     bool // Generate table-driven tests (default: true)
	Parallel  bool // Add t.Parallel() calls
	Mock      bool // Generate mock setup
	Benchmark bool // Include benchmark tests
	Fuzz      bool // Include fuzz tests (Go 1.18+)
}

// TestResult represents the result of test generation
type TestResult struct {
	Status    string   `json:"status"`
	SourceFile string  `json:"source_file"`
	TestFile  string   `json:"test_file"`
	Functions []string `json:"functions"`
}

// RunTestInit generates tests for a Go source file
func RunTestInit(w io.Writer, sourcePath string, opts TestOptions, genOpts Options) error {
	if sourcePath == "" {
		return fmt.Errorf("generate: source file path is required")
	}

	// Check if file exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("generate: file not found: %s", sourcePath)
	}

	// Parse the Go source file
	fset := token.NewFileSet()

	node, err := parser.ParseFile(fset, sourcePath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("generate: failed to parse %s: %w", sourcePath, err)
	}

	// Extract functions
	var functions []testtpl.FuncInfo

	ast.Inspect(node, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		if !fn.Name.IsExported() {
			return true
		}

		funcInfo := testtpl.FuncInfo{
			Name:       fn.Name.Name,
			IsExported: true,
		}

		// Get receiver type if method
		if fn.Recv != nil && len(fn.Recv.List) > 0 {
			if t, ok := fn.Recv.List[0].Type.(*ast.StarExpr); ok {
				if ident, ok := t.X.(*ast.Ident); ok {
					funcInfo.Receiver = ident.Name
				}
			} else if ident, ok := fn.Recv.List[0].Type.(*ast.Ident); ok {
				funcInfo.Receiver = ident.Name
			}
		}

		// Get parameters
		if fn.Type.Params != nil {
			for _, param := range fn.Type.Params.List {
				paramType := exprToString(param.Type)

				for _, name := range param.Names {
					funcInfo.Params = append(funcInfo.Params, testtpl.Param{
						Name: name.Name,
						Type: paramType,
					})
				}

				if len(param.Names) == 0 {
					funcInfo.Params = append(funcInfo.Params, testtpl.Param{
						Name: "arg",
						Type: paramType,
					})
				}
			}
		}

		// Get results
		if fn.Type.Results != nil {
			for _, result := range fn.Type.Results.List {
				funcInfo.Results = append(funcInfo.Results, exprToString(result.Type))
			}
		}

		functions = append(functions, funcInfo)

		return true
	})

	if len(functions) == 0 {
		return fmt.Errorf("generate: no exported functions found in %s", sourcePath)
	}

	// Prepare template data
	data := testtpl.TemplateData{
		Package:   node.Name.Name,
		Functions: functions,
		Parallel:  opts.Parallel,
		Mock:      opts.Mock,
		Benchmark: opts.Benchmark,
		Fuzz:      opts.Fuzz,
	}

	// Generate test file path
	testPath := strings.TrimSuffix(sourcePath, ".go") + "_test.go"

	// Select template
	var tpl string
	if opts.Table {
		tpl = testtpl.TableDrivenTestTemplate
	} else {
		tpl = testtpl.SimpleTestTemplate
	}

	// Add benchmark tests if requested
	if opts.Benchmark {
		tpl += testtpl.BenchmarkTestTemplate
	}

	// Write test file
	if err := writeTemplate(testPath, tpl, data); err != nil {
		return fmt.Errorf("generate: failed to create %s: %w", testPath, err)
	}

	if genOpts.JSON {
		var funcNames []string

		for _, f := range functions {
			funcNames = append(funcNames, f.Name)
		}

		result := TestResult{
			Status:     "created",
			SourceFile: sourcePath,
			TestFile:   testPath,
			Functions:  funcNames,
		}

		return json.NewEncoder(w).Encode(result)
	}

	_, _ = fmt.Fprintf(w, "Created test file: %s\n", testPath)
	_, _ = fmt.Fprintf(w, "Source file: %s\n", sourcePath)
	_, _ = fmt.Fprintf(w, "Package: %s\n", node.Name.Name)

	_, _ = fmt.Fprintln(w, "\nTests generated for:")

	for _, f := range functions {
		if f.Receiver != "" {
			_, _ = fmt.Fprintf(w, "  - %s.%s\n", f.Receiver, f.Name)
		} else {
			_, _ = fmt.Fprintf(w, "  - %s\n", f.Name)
		}
	}

	return nil
}

// exprToString converts an AST expression to string representation
func exprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + exprToString(t.X)
	case *ast.SelectorExpr:
		return exprToString(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		return "[]" + exprToString(t.Elt)
	case *ast.MapType:
		return "map[" + exprToString(t.Key) + "]" + exprToString(t.Value)
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.FuncType:
		return "func"
	case *ast.ChanType:
		return "chan " + exprToString(t.Value)
	case *ast.Ellipsis:
		return "..." + exprToString(t.Elt)
	default:
		return "any"
	}
}
