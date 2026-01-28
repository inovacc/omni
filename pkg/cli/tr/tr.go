package tr

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
)

// TrOptions configures the tr command behavior
type TrOptions struct {
	Delete     bool // -d: delete characters in SET1
	Squeeze    bool // -s: squeeze repeated characters in last SET
	Complement bool // -c: use the complement of SET1
	Truncate   bool // -t: truncate SET1 to length of SET2
}

// RunTr executes the tr command
func RunTr(w io.Writer, r io.Reader, set1, set2 string, opts TrOptions) error {
	// Expand character classes and ranges
	expandedSet1 := expandCharSet(set1)
	expandedSet2 := expandCharSet(set2)

	// Build translation map
	var (
		transMap   map[rune]rune
		deleteSet  map[rune]bool
		squeezeSet map[rune]bool
	)

	if opts.Complement {
		expandedSet1 = complementSet(expandedSet1)
	}

	if opts.Delete {
		deleteSet = make(map[rune]bool)
		for _, r := range expandedSet1 {
			deleteSet[r] = true
		}
	} else {
		transMap = buildTransMap(expandedSet1, expandedSet2, opts.Truncate)
	}

	if opts.Squeeze {
		squeezeSet = make(map[rune]bool)

		targetSet := expandedSet2
		if opts.Delete || expandedSet2 == "" {
			targetSet = expandedSet1
		}

		for _, r := range targetSet {
			squeezeSet[r] = true
		}
	}

	reader := bufio.NewReader(r)

	var (
		lastRune           rune
		lastWasSqueezeChar bool
	)

	for {
		ru, _, err := reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				break
			}

			return err
		}

		// Delete mode
		if opts.Delete && deleteSet[ru] {
			continue
		}

		// Translate mode
		if transMap != nil {
			if replacement, ok := transMap[ru]; ok {
				ru = replacement
			}
		}

		// Squeeze mode
		if opts.Squeeze && squeezeSet[ru] {
			if lastWasSqueezeChar && ru == lastRune {
				continue
			}

			lastWasSqueezeChar = true
		} else {
			lastWasSqueezeChar = false
		}

		lastRune = ru
		_, _ = fmt.Fprint(w, string(ru))
	}

	return nil
}

// expandCharSet expands character classes and ranges
func expandCharSet(set string) string {
	var result strings.Builder

	runes := []rune(set)

	for i := 0; i < len(runes); i++ {
		// Check for character classes
		if i+2 < len(runes) && runes[i] == '[' && runes[i+1] == ':' {
			// Find closing :]
			end := strings.Index(string(runes[i:]), ":]")
			if end > 0 {
				class := string(runes[i+2 : i+end])
				result.WriteString(expandClass(class))

				i += end + 1

				continue
			}
		}

		// Check for range (a-z)
		if i+2 < len(runes) && runes[i+1] == '-' && runes[i+2] != '-' {
			start := runes[i]

			endR := runes[i+2]
			if start <= endR {
				for c := start; c <= endR; c++ {
					result.WriteRune(c)
				}
			}

			i += 2

			continue
		}

		// Check for escape sequences
		if runes[i] == '\\' && i+1 < len(runes) {
			switch runes[i+1] {
			case 'n':
				result.WriteRune('\n')
			case 't':
				result.WriteRune('\t')
			case 'r':
				result.WriteRune('\r')
			case '\\':
				result.WriteRune('\\')
			default:
				result.WriteRune(runes[i+1])
			}

			i++

			continue
		}

		result.WriteRune(runes[i])
	}

	return result.String()
}

// expandClass expands POSIX character classes
func expandClass(class string) string {
	var result strings.Builder

	switch class {
	case "alnum":
		for c := 'a'; c <= 'z'; c++ {
			result.WriteRune(c)
		}

		for c := 'A'; c <= 'Z'; c++ {
			result.WriteRune(c)
		}

		for c := '0'; c <= '9'; c++ {
			result.WriteRune(c)
		}
	case "alpha":
		for c := 'a'; c <= 'z'; c++ {
			result.WriteRune(c)
		}

		for c := 'A'; c <= 'Z'; c++ {
			result.WriteRune(c)
		}
	case "digit":
		for c := '0'; c <= '9'; c++ {
			result.WriteRune(c)
		}
	case "lower":
		for c := 'a'; c <= 'z'; c++ {
			result.WriteRune(c)
		}
	case "upper":
		for c := 'A'; c <= 'Z'; c++ {
			result.WriteRune(c)
		}
	case "space":
		result.WriteString(" \t\n\r\f\v")
	case "blank":
		result.WriteString(" \t")
	case "punct":
		result.WriteString("!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~")
	case "graph":
		// printable characters except space
		for c := '!'; c <= '~'; c++ {
			result.WriteRune(c)
		}
	case "print":
		// printable characters including space
		for c := ' '; c <= '~'; c++ {
			result.WriteRune(c)
		}
	case "xdigit":
		result.WriteString("0123456789ABCDEFabcdef")
	}

	return result.String()
}

// complementSet returns all printable ASCII characters NOT in the set
func complementSet(set string) string {
	inSet := make(map[rune]bool)
	for _, r := range set {
		inSet[r] = true
	}

	var result strings.Builder

	for c := range rune(256) {
		if !inSet[c] && unicode.IsPrint(c) {
			result.WriteRune(c)
		}
	}

	return result.String()
}

// buildTransMap creates the translation map from set1 to set2
func buildTransMap(set1, set2 string, truncate bool) map[rune]rune {
	transMap := make(map[rune]rune)
	runes1 := []rune(set1)
	runes2 := []rune(set2)

	if len(runes2) == 0 {
		return transMap
	}

	// If not truncating, extend set2 with its last character
	if !truncate && len(runes2) < len(runes1) {
		lastChar := runes2[len(runes2)-1]
		for len(runes2) < len(runes1) {
			runes2 = append(runes2, lastChar)
		}
	}

	// Build the map
	for i, r := range runes1 {
		if truncate && i >= len(runes2) {
			break
		}

		if i < len(runes2) {
			transMap[r] = runes2[i]
		}
	}

	return transMap
}

// RunTrFromStdin is a convenience function that reads from stdin
func RunTrFromStdin(w io.Writer, set1, set2 string, opts TrOptions) error {
	return RunTr(w, os.Stdin, set1, set2, opts)
}
