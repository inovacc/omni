package cli

import "strings"

func Grep(lines []string, pattern string) []string {
	var out []string
	for _, l := range lines {
		if strings.Contains(l, pattern) {
			out = append(out, l)
		}
	}
	return out
}
