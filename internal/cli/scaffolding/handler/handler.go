package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/internal/cli/scaffolding"
	handlertpl "github.com/inovacc/omni/internal/cli/scaffolding/handler/templates"
)

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
func RunHandlerInit(w io.Writer, fs afero.Fs, name string, opts HandlerOptions, genOpts scaffolding.Options) error {
	if name == "" {
		return cmderr.Wrap(cmderr.ErrInvalidInput, "scaffold: handler name is required")
	}

	// Reject a name that could escape opts.Dir via path traversal (CWE-22): the
	// name is joined into the output path and may be templated from external
	// input (e.g. a CI job). It must be a bare identifier.
	if strings.ContainsAny(name, `/\`) || strings.Contains(name, "..") {
		return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("scaffold: invalid handler name %q: must not contain path separators or '..'", name))
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
	if err := fs.MkdirAll(opts.Dir, 0755); err != nil {
		return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("scaffold: failed to create directory %s: %v", opts.Dir, err))
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

	if err := scaffolding.WriteTemplate(fs, handlerPath, tpl, data); err != nil {
		return fmt.Errorf("scaffold: failed to create %s: %w", handlerPath, err)
	}

	filesCreated = append(filesCreated, handlerPath)

	// Generate test file (only for stdlib for now)
	if opts.Framework == "stdlib" {
		testPath := filepath.Join(opts.Dir, strings.ToLower(name)+"_test.go")

		if err := scaffolding.WriteTemplate(fs, testPath, handlertpl.HandlerTestTemplate, data); err != nil {
			return fmt.Errorf("scaffold: failed to create %s: %w", testPath, err)
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
