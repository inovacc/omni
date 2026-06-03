package collect

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/inovacc/omni/pkg/sbom/model"
)

// ModuleDir builds an SBOM from a Go module directory by parsing its go.mod
// require blocks. The module directive becomes the root component; every
// required module becomes a KindLibrary component. A missing go.mod returns an
// error wrapping os.ErrNotExist so the CLI maps it to cmderr.ErrNotFound.
//
// replace/exclude/retract directives are ignored for module-dir SBOMs in this
// phase (replace resolution is binary-only — see binary.go). Components are
// returned unsorted; callers should call (*model.SBOM).Normalize before
// emitting deterministic output.
func ModuleDir(dir string) (*model.SBOM, error) {
	path := filepath.Join(dir, "go.mod")
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("go.mod: %w", os.ErrNotExist)
	}
	defer func() { _ = f.Close() }()

	sb := &model.SBOM{}
	inRequireBlock := false
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		switch {
		case line == "":
			continue
		case strings.HasPrefix(line, "module "):
			modPath := strings.TrimSpace(strings.TrimPrefix(line, "module "))
			sb.Root = model.Component{Path: modPath, Kind: model.KindRoot}
			sb.Name = lastPathElement(modPath)
			continue
		case line == "require (":
			inRequireBlock = true
			continue
		case inRequireBlock && line == ")":
			inRequireBlock = false
			continue
		}

		if inRequireBlock {
			if c, ok := parseRequire(line); ok {
				sb.Components = append(sb.Components, c)
			}
			continue
		}

		if strings.HasPrefix(line, "require ") {
			if c, ok := parseRequire(strings.TrimPrefix(line, "require ")); ok {
				sb.Components = append(sb.Components, c)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan go.mod: %w", err)
	}
	return sb, nil
}

// parseRequire parses a single require line ("<path> <version> [// indirect]")
// into a KindLibrary component. It returns ok=false for lines that do not name
// a module path/version pair.
func parseRequire(line string) (model.Component, bool) {
	if i := strings.Index(line, "//"); i >= 0 {
		line = line[:i]
	}
	fields := strings.Fields(line)
	if len(fields) >= 2 && looksLikeModulePath(fields[0]) {
		return model.Component{Path: fields[0], Version: fields[1], Kind: model.KindLibrary}, true
	}
	return model.Component{}, false
}

// looksLikeModulePath reports whether s plausibly names a Go module path.
func looksLikeModulePath(s string) bool {
	return strings.ContainsAny(s, "./")
}

// lastPathElement returns the final '/'-separated element of a module path.
func lastPathElement(p string) string {
	if i := strings.LastIndexByte(p, '/'); i >= 0 {
		return p[i+1:]
	}
	return p
}
