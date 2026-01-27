package cli

import (
	"sort"
	"strings"
)

func Sort(lines []string) {
	sort.Strings(lines)
}

func Uniq(lines []string) []string {
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

func TrimLines(lines []string) []string {
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		out = append(out, strings.TrimSpace(l))
	}
	return out
}
