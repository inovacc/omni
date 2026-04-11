// Command cmderr-cov parses a Go coverage profile and fails if the total
// statement coverage for the internal/cli/cmderr package is below a threshold.
//
// Per the omni "no exec" and "pure Go" aesthetic, this replaces the shell/awk
// one-liner the Taskfile research originally suggested. It is intentionally
// dependency-free (stdlib only) and cross-platform.
//
// Usage:
//
//	cmderr-cov -profile=cmderr-cov.out -min=90
//
// Exit codes:
//
//	0 — coverage >= min
//	1 — coverage <  min
//	2 — usage / I/O error
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	profile := flag.String("profile", "cmderr-cov.out", "path to go coverage profile")
	min := flag.Float64("min", 90.0, "minimum acceptable total coverage percentage")
	pkg := flag.String("pkg", "internal/cli/cmderr", "package substring to match in coverage profile")
	flag.Parse()

	if _, err := os.Stat(*profile); err != nil {
		fmt.Fprintf(os.Stderr, "cmderr-cov: cannot read profile %q: %v\n", *profile, err)
		os.Exit(2)
	}

	pct, err := totalCoverage(*profile, *pkg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cmderr-cov: %v\n", err)
		os.Exit(2)
	}

	if pct+1e-9 < *min {
		fmt.Fprintf(os.Stderr, "FAIL: %s coverage %.1f%% < %.1f%%\n", *pkg, pct, *min)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "OK: %s coverage %.1f%% >= %.1f%%\n", *pkg, pct, *min)
}

// totalCoverage shells out to `go tool cover -func=<profile>` and extracts the
// total percentage line. Using `go tool cover` keeps us aligned with whatever
// Go's coverage format evolves into without reimplementing profile parsing.
func totalCoverage(profile, pkg string) (float64, error) {
	// First: sanity-check the profile actually contains the requested package
	// so we don't silently pass on a mis-targeted profile.
	if err := assertProfileContains(profile, pkg); err != nil {
		return 0, err
	}

	cmd := exec.Command("go", "tool", "cover", "-func="+profile)
	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("go tool cover failed: %w", err)
	}

	re := regexp.MustCompile(`([0-9]+\.[0-9]+)%`)
	var totalLine string
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "total:") {
			totalLine = line
			break
		}
	}
	if totalLine == "" {
		return 0, fmt.Errorf("no total: line in coverage output")
	}
	m := re.FindStringSubmatch(totalLine)
	if len(m) != 2 {
		return 0, fmt.Errorf("could not parse percentage from %q", totalLine)
	}
	return strconv.ParseFloat(m[1], 64)
}

func assertProfileContains(profile, pkg string) error {
	f, err := os.Open(profile)
	if err != nil {
		return fmt.Errorf("open profile: %w", err)
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), pkg) {
			return nil
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan profile: %w", err)
	}
	return fmt.Errorf("profile %q contains no entries for package %q", profile, pkg)
}
