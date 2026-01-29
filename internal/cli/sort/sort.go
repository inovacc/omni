package sort

import (
	gosort "sort"
	"strings"
)

func Pipeline(
	lines []string,
	grepPattern string,
	doSort bool,
	doUniq bool,
) []string {
	out := lines

	if grepPattern != "" {
		out = grepLines(out, grepPattern)
	}

	if doSort {
		gosort.Strings(out)
	}

	if doUniq {
		out = uniqLines(out)
	}

	return out
}

func grepLines(lines []string, pattern string) []string {
	var result []string

	for _, l := range lines {
		if strings.Contains(l, pattern) {
			result = append(result, l)
		}
	}

	return result
}

func uniqLines(lines []string) []string {
	seen := make(map[string]bool)
	out := make([]string, 0, len(lines))

	for _, l := range lines {
		if !seen[l] {
			seen[l] = true
			out = append(out, l)
		}
	}

	return out
}
