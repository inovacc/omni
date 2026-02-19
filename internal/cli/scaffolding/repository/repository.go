package repository

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/inovacc/omni/internal/cli/scaffolding"
	repotpl "github.com/inovacc/omni/internal/cli/scaffolding/repository/templates"
)

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
func RunRepositoryInit(w io.Writer, name string, opts RepositoryOptions, genOpts scaffolding.Options) error {
	if name == "" {
		return fmt.Errorf("scaffold: repository name is required")
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
		return fmt.Errorf("scaffold: failed to create directory %s: %w", opts.Dir, err)
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

		if err := scaffolding.WriteTemplate(interfacePath, repotpl.InterfaceTemplate, data); err != nil {
			return fmt.Errorf("scaffold: failed to create %s: %w", interfacePath, err)
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

	if err := scaffolding.WriteTemplate(repoPath, tpl, data); err != nil {
		return fmt.Errorf("scaffold: failed to create %s: %w", repoPath, err)
	}

	filesCreated = append(filesCreated, repoPath)

	// Generate test file
	testPath := filepath.Join(opts.Dir, strings.ToLower(name)+"_test.go")

	if err := scaffolding.WriteTemplate(testPath, repotpl.RepositoryTestTemplate, data); err != nil {
		return fmt.Errorf("scaffold: failed to create %s: %w", testPath, err)
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
