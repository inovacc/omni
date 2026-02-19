package testgen

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"strings"

	"github.com/inovacc/omni/internal/cli/scaffolding"
	testtpl "github.com/inovacc/omni/internal/cli/scaffolding/testgen/templates"
)

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
	Status     string   `json:"status"`
	SourceFile string   `json:"source_file"`
	TestFile   string   `json:"test_file"`
	Functions  []string `json:"functions"`
}

// RunTestInit generates tests for a Go source file
func RunTestInit(w io.Writer, sourcePath string, opts TestOptions, genOpts scaffolding.Options) error {
	if sourcePath == "" {
		return fmt.Errorf("scaffold: source file path is required")
	}

	// Check if file exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("scaffold: file not found: %s", sourcePath)
	}

	// Parse the Go source file
	fset := token.NewFileSet()

	node, err := parser.ParseFile(fset, sourcePath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("scaffold: failed to parse %s: %w", sourcePath, err)
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
		return fmt.Errorf("scaffold: no exported functions found in %s", sourcePath)
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
	if err := scaffolding.WriteTemplate(testPath, tpl, data); err != nil {
		return fmt.Errorf("scaffold: failed to create %s: %w", testPath, err)
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
