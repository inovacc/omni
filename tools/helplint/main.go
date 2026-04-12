// Package main implements a cobra help docstring linter.
//
// By default, lints all *.go files in the target directory.
// Files may opt out by including the comment:
//
//	// helplint:ignore
//
// Rules enforced (exit 1 on violation):
//  1. Every cobra.Command with a non-empty Short field must also have a non-empty Long field.
//  2. If Long is present, it must contain "omni " (with a space) as a usage example.
//
// Exit codes:
//
//	0 — no violations
//	1 — one or more violations
//	2 — parse error
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

const ignoreDirective = "helplint:ignore"

func main() {
	dir := flag.String("dir", "cmd/", "Directory to scan for cobra commands")
	verbose := flag.Bool("verbose", false, "Print verbose output including passing commands")
	flag.Parse()

	violations, err := lintDir(*dir, *verbose)
	if err != nil {
		fmt.Fprintf(os.Stderr, "helplint: parse error: %v\n", err)
		os.Exit(2)
	}

	for _, v := range violations {
		fmt.Println(v)
	}

	if len(violations) > 0 {
		os.Exit(1)
	}
}

// lintDir walks a directory and lints all *.go files (unless marked helplint:ignore).
func lintDir(dir string, verbose bool) ([]string, error) {
	var violations []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading directory %q: %w", dir, err)
	}

	fset := token.NewFileSet()

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}

		filePath := filepath.Join(dir, entry.Name())
		fv, err := lintFile(fset, filePath, verbose)
		if err != nil {
			return nil, err
		}
		violations = append(violations, fv...)
	}

	return violations, nil
}

// lintFile parses a single Go file and checks all cobra.Command composite literals.
// Returns nil violations if the file contains a // helplint:ignore comment.
func lintFile(fset *token.FileSet, filePath string, verbose bool) ([]string, error) {
	src, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading %q: %w", filePath, err)
	}

	// Fast path: skip files with the ignore directive.
	if strings.Contains(string(src), ignoreDirective) {
		if verbose {
			fmt.Fprintf(os.Stderr, "helplint: %s: skipped (helplint:ignore)\n", filePath)
		}
		return nil, nil
	}

	f, err := parser.ParseFile(fset, filePath, src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parsing %q: %w", filePath, err)
	}

	var violations []string

	ast.Inspect(f, func(n ast.Node) bool {
		lit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}

		if !isCobraCommand(lit) {
			return true
		}

		short, long, hasLong := extractHelpFields(lit)
		hasShort := strings.TrimSpace(short) != ""

		if verbose {
			fmt.Fprintf(os.Stderr, "helplint: %s: Short=%q hasLong=%v Long=%q\n",
				filePath, short, hasLong, long)
		}

		// Rule 1: Short present but no Long field at all.
		if hasShort && !hasLong {
			violations = append(violations, fmt.Sprintf(
				"%s: Short or Long missing/invalid for a cobra.Command (Short=%q, Long field absent)",
				filePath, short,
			))
			return true
		}

		// Rule 2: Long present but lacks "omni " usage example.
		if hasLong && !strings.Contains(long, "omni ") {
			violations = append(violations, fmt.Sprintf(
				"%s: Short or Long missing/invalid for a cobra.Command (Short=%q, Long lacks 'omni ' usage example)",
				filePath, short,
			))
		}

		return true
	})

	return violations, nil
}

// isCobraCommand returns true if the composite literal is a cobra.Command.
func isCobraCommand(lit *ast.CompositeLit) bool {
	if lit.Type == nil {
		return false
	}

	// Match cobra.Command (SelectorExpr: cobra.Command)
	if sel, ok := lit.Type.(*ast.SelectorExpr); ok {
		if ident, ok := sel.X.(*ast.Ident); ok {
			return ident.Name == "cobra" && sel.Sel.Name == "Command"
		}
	}

	// Match Command (plain Ident, when already in cobra package)
	if ident, ok := lit.Type.(*ast.Ident); ok {
		return ident.Name == "Command"
	}

	return false
}

// extractHelpFields returns the string values of Short and Long fields from a cobra.Command literal,
// along with a boolean indicating whether Long was present as a non-empty literal.
func extractHelpFields(lit *ast.CompositeLit) (short, long string, hasLong bool) {
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		key, ok := kv.Key.(*ast.Ident)
		if !ok {
			continue
		}

		switch key.Name {
		case "Short":
			short = extractStringLiteral(kv.Value)
		case "Long":
			s := extractStringLiteral(kv.Value)
			if s != "" {
				hasLong = true
				long = s
			}
		}
	}
	return short, long, hasLong
}

// extractStringLiteral returns the unquoted string value of a basic string literal node.
// Returns empty string for non-literal expressions (variables, concatenations, etc.).
func extractStringLiteral(expr ast.Expr) string {
	switch v := expr.(type) {
	case *ast.BasicLit:
		if v.Kind == token.STRING {
			s := v.Value
			if len(s) >= 2 {
				if (s[0] == '"' && s[len(s)-1] == '"') ||
					(s[0] == '`' && s[len(s)-1] == '`') {
					return s[1 : len(s)-1]
				}
			}
		}
	}
	return ""
}
