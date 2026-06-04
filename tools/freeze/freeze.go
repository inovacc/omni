//go:build ignore

// Command freeze enumerates the frozen (non-Experimental) public API of pkg/*.
// Output is deterministic (sorted) so CI can diff it against docs/API-FREEZE.md.
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	root := "pkg"
	var lines []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return err
		}
		fset := token.NewFileSet()
		pkgs, perr := parser.ParseDir(fset, path, func(fi os.FileInfo) bool {
			return strings.HasSuffix(fi.Name(), ".go") && !strings.HasSuffix(fi.Name(), "_test.go")
		}, parser.ParseComments)
		if perr != nil || len(pkgs) == 0 {
			return nil
		}
		for name, pkg := range pkgs {
			if experimental(pkg) {
				continue
			}
			rel := filepath.ToSlash(path)
			for _, sym := range exportedSymbols(pkg) {
				lines = append(lines, rel+" "+name+"."+sym)
			}
		}
		return nil
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	sort.Strings(lines)
	fmt.Println("# API Freeze — frozen pkg/* public surface (regenerate: task freeze:check)")
	for _, l := range lines {
		fmt.Println(l)
	}
}

// experimental reports whether any file's package doc marks the package Experimental.
func experimental(pkg *ast.Package) bool {
	for _, f := range pkg.Files {
		if f.Doc != nil && strings.Contains(f.Doc.Text(), "Experimental:") {
			return true
		}
	}
	return false
}

// exportedSymbols returns sorted exported top-level identifiers (+ exported
// methods and struct fields) declared in pkg.
func exportedSymbols(pkg *ast.Package) []string {
	seen := map[string]struct{}{}
	for _, f := range pkg.Files {
		for _, decl := range f.Decls {
			switch dd := decl.(type) {
			case *ast.FuncDecl:
				if !dd.Name.IsExported() {
					continue
				}
				if dd.Recv != nil { // method
					recv := recvType(dd.Recv)
					if ast.IsExported(recv) {
						seen[recv+"."+dd.Name.Name+"()"] = struct{}{}
					}
					continue
				}
				seen[dd.Name.Name+"()"] = struct{}{}
			case *ast.GenDecl:
				for _, spec := range dd.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						if s.Name.IsExported() {
							seen[s.Name.Name] = struct{}{}
							if st, ok := s.Type.(*ast.StructType); ok {
								for _, fld := range st.Fields.List {
									for _, n := range fld.Names {
										if n.IsExported() {
											seen[s.Name.Name+"#"+n.Name] = struct{}{}
										}
									}
								}
							}
						}
					case *ast.ValueSpec:
						for _, n := range s.Names {
							if n.IsExported() {
								seen[n.Name] = struct{}{}
							}
						}
					}
				}
			}
		}
	}
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func recvType(fl *ast.FieldList) string {
	if len(fl.List) == 0 {
		return ""
	}
	switch t := fl.List[0].Type.(type) {
	case *ast.StarExpr:
		if id, ok := t.X.(*ast.Ident); ok {
			return id.Name
		}
	case *ast.Ident:
		return t.Name
	}
	return ""
}
