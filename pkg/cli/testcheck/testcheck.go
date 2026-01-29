package testcheck

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// PackageStatus represents the test status of a package.
type PackageStatus struct {
	Name      string   `json:"name"`
	Path      string   `json:"path"`
	HasTests  bool     `json:"has_tests"`
	TestFiles []string `json:"test_files,omitempty"`
	GoFiles   int      `json:"go_files"`
}

// Result holds the overall test check results.
type Result struct {
	Total     int             `json:"total"`
	WithTests int             `json:"with_tests"`
	NoTests   int             `json:"no_tests"`
	Coverage  float64         `json:"coverage_percent"`
	Packages  []PackageStatus `json:"packages"`
}

// Options configures the testcheck behavior.
type Options struct {
	JSON    bool // Output as JSON
	ShowAll bool // Show all packages (default shows only missing)
	Summary bool // Show only summary
	Verbose bool // Show test file names
}

// Run performs the test check on the given directory.
func Run(w io.Writer, dir string, opts Options) error {
	result, err := Check(dir)
	if err != nil {
		return err
	}

	return printResult(w, result, opts)
}

// Check scans the directory and returns test status for all packages.
func Check(dir string) (*Result, error) {
	packages := make(map[string]*PackageStatus)

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and vendor
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") || name == "vendor" || name == "testdata" {
				return filepath.SkipDir
			}

			return nil
		}

		// Only process .go files
		if !strings.HasSuffix(d.Name(), ".go") {
			return nil
		}

		// Get the package directory
		pkgDir := filepath.Dir(path)

		relDir, err := filepath.Rel(dir, pkgDir)
		if err != nil {
			relDir = pkgDir
		}

		// Initialize package if not seen
		if _, exists := packages[pkgDir]; !exists {
			packages[pkgDir] = &PackageStatus{
				Name: filepath.Base(pkgDir),
				Path: relDir,
			}
		}

		pkg := packages[pkgDir]

		// Check if it's a test file
		if strings.HasSuffix(d.Name(), "_test.go") {
			pkg.HasTests = true
			pkg.TestFiles = append(pkg.TestFiles, d.Name())
		} else {
			pkg.GoFiles++
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	// Convert map to sorted slice
	var pkgList []PackageStatus

	for _, pkg := range packages {
		// Only include packages with Go source files
		if pkg.GoFiles > 0 {
			pkgList = append(pkgList, *pkg)
		}
	}

	sort.Slice(pkgList, func(i, j int) bool {
		return pkgList[i].Path < pkgList[j].Path
	})

	// Calculate totals
	result := &Result{
		Total:    len(pkgList),
		Packages: pkgList,
	}

	for _, pkg := range pkgList {
		if pkg.HasTests {
			result.WithTests++
		} else {
			result.NoTests++
		}
	}

	if result.Total > 0 {
		result.Coverage = float64(result.WithTests) / float64(result.Total) * 100
	}

	return result, nil
}

func printResult(w io.Writer, result *Result, opts Options) error {
	if opts.JSON {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")

		return enc.Encode(result)
	}

	if opts.Summary {
		_, _ = fmt.Fprintf(w, "Total: %d | With Tests: %d | No Tests: %d | Coverage: %.1f%%\n",
			result.Total, result.WithTests, result.NoTests, result.Coverage)

		return nil
	}

	// Print header
	_, _ = fmt.Fprintf(w, "Test Coverage Check\n")
	_, _ = fmt.Fprintf(w, "%s\n\n", strings.Repeat("=", 50))

	// Print packages
	for _, pkg := range result.Packages {
		if !opts.ShowAll && pkg.HasTests {
			continue
		}

		status := "NO TEST"
		if pkg.HasTests {
			status = "HAS TEST"
		}

		_, _ = fmt.Fprintf(w, "%-40s %s\n", pkg.Path, status)

		if opts.Verbose && pkg.HasTests && len(pkg.TestFiles) > 0 {
			for _, f := range pkg.TestFiles {
				_, _ = fmt.Fprintf(w, "  - %s\n", f)
			}
		}
	}

	// Print summary
	_, _ = fmt.Fprintf(w, "\n%s\n", strings.Repeat("-", 50))
	_, _ = fmt.Fprintf(w, "Total: %d | With Tests: %d | No Tests: %d | Coverage: %.1f%%\n",
		result.Total, result.WithTests, result.NoTests, result.Coverage)

	return nil
}
