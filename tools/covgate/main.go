// Package main implements a dual-rule coverage gate for omni.
// It parses Go coverage profiles and enforces a per-package floor
// and a weighted average minimum — with no subprocess invocations.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// pkgStats accumulates statement counts for one package.
type pkgStats struct {
	covered int
	total   int
}

func main() {
	profile := flag.String("profile", "coverage.out", "Go coverage profile path")
	pkgPrefix := flag.String("pkg-prefix", "", "Import-path prefix to include (e.g. github.com/inovacc/omni/pkg/)")
	avgMin := flag.Float64("avg-min", 75.0, "Weighted average coverage minimum (%)")
	floor := flag.Float64("floor", 40.0, "Per-package coverage floor (%)")
	exclude := flag.String("exclude", "private", "Comma-separated path substrings to exclude")
	flag.Parse()

	excludes := []string{}
	for _, s := range strings.Split(*exclude, ",") {
		s = strings.TrimSpace(s)
		if s != "" {
			excludes = append(excludes, s)
		}
	}

	pkgs, err := parseProfile(*profile, *pkgPrefix, excludes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "covgate: %v\n", err)
		os.Exit(2)
	}

	if len(pkgs) == 0 {
		fmt.Fprintf(os.Stderr, "covgate: no packages matched prefix %q\n", *pkgPrefix)
		os.Exit(2)
	}

	violations := false

	// Floor rule
	for name, s := range pkgs {
		if s.total == 0 {
			continue
		}
		pct := float64(s.covered) / float64(s.total) * 100.0
		if pct < *floor {
			fmt.Printf("FAIL %s: %.1f%% < %.1f%% floor\n", shortName(name, *pkgPrefix), pct, *floor)
			violations = true
		}
	}

	// Weighted average rule
	var totalCovered, totalStmts int
	for _, s := range pkgs {
		totalCovered += s.covered
		totalStmts += s.total
	}
	if totalStmts > 0 {
		wavg := float64(totalCovered) / float64(totalStmts) * 100.0
		if wavg < *avgMin {
			fmt.Printf("FAIL weighted avg %.1f%% < %.1f%% minimum\n", wavg, *avgMin)
			violations = true
		} else {
			fmt.Printf("OK   weighted avg %.1f%% >= %.1f%% minimum\n", wavg, *avgMin)
		}
	}

	if violations {
		os.Exit(1)
	}
}

// shortName strips the pkg-prefix from a package import path for readable output.
func shortName(importPath, prefix string) string {
	if strings.HasPrefix(importPath, prefix) {
		return importPath[len(prefix):]
	}
	return importPath
}

// parseProfile reads a Go coverage profile and returns per-package statement stats.
// Profile format (after the header line):
//
//	github.com/owner/repo/pkg/foo/foo.go:25.45,27.12 3 1
//	                                                  ^stmts ^count (0=not covered)
func parseProfile(path, prefix string, excludes []string) (map[string]*pkgStats, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open profile %q: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	pkgs := make(map[string]*pkgStats)
	scanner := bufio.NewScanner(f)
	first := true
	for scanner.Scan() {
		line := scanner.Text()
		if first {
			first = false
			// Skip "mode: set" header
			continue
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Split off the file path (everything before the last colon that precedes positions).
		// Format: <file>:<startLine>.<startCol>,<endLine>.<endCol> <stmts> <count>
		// We need the package portion of <file>.
		colonIdx := strings.LastIndex(line, ":")
		if colonIdx < 0 {
			continue
		}
		filePath := line[:colonIdx]
		rest := line[colonIdx+1:]

		// rest = "25.45,27.12 3 1"
		parts := strings.Fields(rest)
		if len(parts) < 2 {
			continue
		}
		stmts, err := strconv.Atoi(parts[0])
		if err != nil || stmts < 0 {
			// parts[0] contains position (e.g. "25.45,27.12"), actual stmts is parts[1]
			// Reparse: the colon split above may have cut off mid-position.
			// Re-approach: split on space to get the trailing two integers.
			fields := strings.Fields(line)
			if len(fields) < 3 {
				continue
			}
			stmts, err = strconv.Atoi(fields[len(fields)-2])
			if err != nil {
				continue
			}
			count, err := strconv.Atoi(fields[len(fields)-1])
			if err != nil {
				continue
			}
			// re-derive filePath from fields[0] (up to colon before line numbers)
			// fields[0] = "pkg/foo/foo.go:25.45,27.12"
			ci := strings.LastIndex(fields[0], ":")
			if ci < 0 {
				continue
			}
			filePath = fields[0][:ci]
			pkgPath := packageOf(filePath)
			if !matches(pkgPath, prefix, excludes) {
				continue
			}
			ps := getOrCreate(pkgs, pkgPath)
			ps.total += stmts
			if count > 0 {
				ps.covered += stmts
			}
			continue
		}
		count, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}

		pkgPath := packageOf(filePath)
		if !matches(pkgPath, prefix, excludes) {
			continue
		}
		ps := getOrCreate(pkgs, pkgPath)
		ps.total += stmts
		if count > 0 {
			ps.covered += stmts
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan profile: %w", err)
	}
	return pkgs, nil
}

// packageOf strips the file name from an import path, returning the package path.
// e.g. "github.com/inovacc/omni/pkg/foo/foo.go" → "github.com/inovacc/omni/pkg/foo"
func packageOf(filePath string) string {
	idx := strings.LastIndex(filePath, "/")
	if idx < 0 {
		return filePath
	}
	return filePath[:idx]
}

// matches returns true if pkgPath starts with prefix and contains none of the excludes.
func matches(pkgPath, prefix string, excludes []string) bool {
	if prefix != "" && !strings.HasPrefix(pkgPath, prefix) {
		return false
	}
	for _, ex := range excludes {
		if strings.Contains(pkgPath, ex) {
			return false
		}
	}
	return true
}

// getOrCreate returns the pkgStats for the given key, creating it if absent.
func getOrCreate(m map[string]*pkgStats, key string) *pkgStats {
	if ps, ok := m[key]; ok {
		return ps
	}
	ps := &pkgStats{}
	m[key] = ps
	return ps
}
