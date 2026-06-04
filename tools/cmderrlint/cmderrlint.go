// Command cmderrlint is a stdlib-only static check that guards the cmderr
// error-classification contract: command code under internal/cli must not
// return a raw standard-library error sentinel (e.g. os.ErrNotExist, io.EOF)
// directly, because such errors bypass cmderr classification and fall through
// to exit code 1 instead of the correct class (see docs/EXIT-CODES.md).
//
// It is deliberately pure-Go (go/parser + go/ast, no golang.org/x/tools) to
// honor the lean-go.mod invariant, mirroring tools/freeze. The codebase is
// currently clean (zero violations), so this runs as a blocking regression
// guard: `go run ./tools/cmderrlint` (defaults to internal/cli) exits non-zero
// the moment a raw sentinel return is introduced.
//
// What it flags (in a `return` statement, as the directly-returned expression):
//   - a known stdlib error sentinel selector: os.ErrNotExist, os.ErrPermission, …
//   - fmt.Errorf("…: %w", <stdlib sentinel>) — %w-wrapping a raw stdlib sentinel
//
// What it does NOT flag (intentional, to keep false positives at zero):
//   - errors.Is/errors.As comparisons (the sentinel is an argument, not returned)
//   - returns of a pre-existing err variable, or cmderr-wrapped sentinels
//   - io.EOF and context.Canceled/DeadlineExceeded (dual-natured control flow)
//   - _test.go files, the cmderr package itself, and pkg/ libraries (which
//     legitimately use stdlib errors and do not depend on cmderr)
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// rawSentinels maps "pkg.Name" stdlib error sentinels to a short reason. These
// are values a command should classify via a cmderr sentinel instead of
// returning raw. The list is intentionally conservative (common cases only).
// io.EOF and context.Canceled/DeadlineExceeded are intentionally NOT listed:
// they are dual-natured (a real error AND idiomatic control flow — end-of-input,
// cancellation) and are routinely returned up to an in-package caller, so
// flagging them produces false positives. The rule targets only sentinels that,
// when returned from command code, should always be classified.
var rawSentinels = map[string]string{
	"os.ErrNotExist":      "classify as cmderr.ErrNotFound",
	"os.ErrExist":         "classify (e.g. cmderr.ErrConflict)",
	"os.ErrPermission":    "classify as cmderr.ErrPermission",
	"os.ErrClosed":        "classify as cmderr.ErrIO",
	"os.ErrInvalid":       "classify as cmderr.ErrInvalidInput",
	"io.ErrUnexpectedEOF": "classify as cmderr.ErrIO",
	"io.ErrClosedPipe":    "classify as cmderr.ErrIO",
	"io.ErrShortWrite":    "classify as cmderr.ErrIO",
	"sql.ErrNoRows":       "classify as cmderr.ErrNotFound",
	"sql.ErrConnDone":     "classify as cmderr.ErrIO",
}

// Finding is one flagged raw-sentinel return.
type Finding struct {
	Pos    token.Position
	Expr   string // the offending expression, e.g. "os.ErrNotExist"
	Reason string // suggested remediation
}

// String renders a Finding as a compiler-style diagnostic line.
func (f Finding) String() string {
	return fmt.Sprintf("%s: returns raw %s — %s (route through internal/cli/cmderr)", f.Pos, f.Expr, f.Reason)
}

// selectorName returns the "pkg.Name" form of a selector expression whose base
// is a bare identifier (e.g. os.ErrNotExist), or "" if expr is not such a
// selector.
func selectorName(expr ast.Expr) string {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return ""
	}

	pkg, ok := sel.X.(*ast.Ident)
	if !ok {
		return ""
	}

	return pkg.Name + "." + sel.Sel.Name
}

// isFmtErrorf reports whether call is a call to fmt.Errorf.
func isFmtErrorf(call *ast.CallExpr) bool {
	return selectorName(call.Fun) == "fmt.Errorf"
}

// checkReturned inspects a single directly-returned expression and appends a
// Finding if it is (or %w-wraps) a raw stdlib sentinel.
func checkReturned(fset *token.FileSet, expr ast.Expr, out *[]Finding) {
	// Case 1: the returned expression IS a raw sentinel selector.
	if name := selectorName(expr); name != "" {
		if reason, bad := rawSentinels[name]; bad {
			*out = append(*out, Finding{Pos: fset.Position(expr.Pos()), Expr: name, Reason: reason})
		}
		return
	}

	// Case 2: fmt.Errorf("…: %w", <sentinel>) — %w-wrapping a raw sentinel.
	call, ok := expr.(*ast.CallExpr)
	if !ok || !isFmtErrorf(call) {
		return
	}

	for _, arg := range call.Args {
		if name := selectorName(arg); name != "" {
			if reason, bad := rawSentinels[name]; bad {
				*out = append(*out, Finding{Pos: fset.Position(arg.Pos()), Expr: name, Reason: reason})
			}
		}
	}
}

// Check parses and scans a single Go file's AST for raw-sentinel returns.
func Check(fset *token.FileSet, file *ast.File) []Finding {
	var out []Finding

	ast.Inspect(file, func(n ast.Node) bool {
		ret, ok := n.(*ast.ReturnStmt)
		if !ok {
			return true
		}

		for _, res := range ret.Results {
			checkReturned(fset, res, &out)
		}

		return true
	})

	return out
}

// scanDir walks dir and checks every non-test .go file (skipping the cmderr
// package, which defines the sentinels, and pkg/ libraries).
func scanDir(dir string) ([]Finding, error) {
	var all []Finding
	fset := token.NewFileSet()

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// The cmderr package legitimately references the sentinels it wraps.
		if strings.Contains(filepath.ToSlash(path), "internal/cli/cmderr/") {
			return nil
		}

		f, perr := parser.ParseFile(fset, path, nil, 0)
		if perr != nil {
			return fmt.Errorf("parse %s: %w", path, perr)
		}

		all = append(all, Check(fset, f)...)
		return nil
	})

	return all, err
}

func main() {
	dirs := os.Args[1:]
	if len(dirs) == 0 {
		dirs = []string{"internal/cli"}
	}

	var findings []Finding

	for _, dir := range dirs {
		fs, err := scanDir(dir)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "cmderrlint: %v\n", err)
			os.Exit(2)
		}

		findings = append(findings, fs...)
	}

	if len(findings) == 0 {
		_, _ = fmt.Fprintf(os.Stdout, "cmderrlint: ok — no raw stdlib-error returns in %s\n", strings.Join(dirs, ", "))
		return
	}

	sort.Slice(findings, func(i, j int) bool {
		return findings[i].Pos.String() < findings[j].Pos.String()
	})

	for _, f := range findings {
		_, _ = fmt.Fprintln(os.Stderr, f.String())
	}

	_, _ = fmt.Fprintf(os.Stderr, "cmderrlint: %d raw stdlib-error return(s) bypass cmderr classification\n", len(findings))
	os.Exit(1)
}
