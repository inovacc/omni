package pipeline

import (
	"fmt"
	"strconv"
	"strings"
)

// Parse converts a CLI string like "grep -i error" into a Stage.
func Parse(cmdLine string) (Stage, error) {
	parts := parseCommandLine(cmdLine)
	if len(parts) == 0 {
		return nil, fmt.Errorf("pipeline: empty command")
	}

	cmd := parts[0]
	args := parts[1:]

	switch cmd {
	case "grep":
		return parseGrep(args, false)
	case "grep-v":
		return parseGrep(args, true)
	case "contains":
		return parseContains(args)
	case "replace":
		return parseReplace(args)
	case "head", "take":
		return parseHead(args)
	case "tail":
		return parseTail(args)
	case "skip":
		return parseSkip(args)
	case "sort":
		return parseSort(args)
	case "uniq":
		return parseUniq(args)
	case "cut":
		return parseCut(args)
	case "tr":
		return parseTr(args)
	case "sed":
		return parseSed(args)
	case "rev":
		return &Rev{}, nil
	case "nl":
		return parseNl(args)
	case "tee":
		return parseTee(args)
	case "tac":
		return &Tac{}, nil
	case "wc":
		return parseWc(args)
	default:
		return nil, fmt.Errorf("pipeline: unknown stage %q", cmd)
	}
}

// ParseAll parses multiple CLI strings into stages.
func ParseAll(cmdLines []string) ([]Stage, error) {
	stages := make([]Stage, 0, len(cmdLines))

	for _, line := range cmdLines {
		s, err := Parse(line)
		if err != nil {
			return nil, err
		}

		stages = append(stages, s)
	}

	return stages, nil
}

func parseGrep(args []string, invert bool) (Stage, error) {
	g := &Grep{Invert: invert}

	i := 0
	for i < len(args) {
		switch args[i] {
		case "-i":
			g.IgnoreCase = true
		case "-v":
			g.Invert = true
		default:
			g.Pattern = args[i]
		}

		i++
	}

	if g.Pattern == "" {
		return nil, fmt.Errorf("grep: missing pattern")
	}

	return g, nil
}

func parseContains(args []string) (Stage, error) {
	c := &Contains{}

	i := 0
	for i < len(args) {
		switch args[i] {
		case "-i":
			c.IgnoreCase = true
		default:
			c.Substr = args[i]
		}

		i++
	}

	if c.Substr == "" {
		return nil, fmt.Errorf("contains: missing substring")
	}

	return c, nil
}

func parseReplace(args []string) (Stage, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("replace: requires OLD NEW arguments")
	}

	return &Replace{Old: args[0], New: args[1]}, nil
}

func parseHead(args []string) (Stage, error) {
	h := &Head{N: 10}

	for i := 0; i < len(args); i++ {
		if args[i] == "-n" && i+1 < len(args) {
			n, err := strconv.Atoi(args[i+1])
			if err != nil {
				return nil, fmt.Errorf("head: invalid count %q", args[i+1])
			}

			h.N = n
			i++
		} else {
			n, err := strconv.Atoi(args[i])
			if err == nil {
				h.N = n
			}
		}
	}

	return h, nil
}

func parseTail(args []string) (Stage, error) {
	t := &Tail{N: 10}

	for i := 0; i < len(args); i++ {
		if args[i] == "-n" && i+1 < len(args) {
			n, err := strconv.Atoi(args[i+1])
			if err != nil {
				return nil, fmt.Errorf("tail: invalid count %q", args[i+1])
			}

			t.N = n
			i++
		} else {
			n, err := strconv.Atoi(args[i])
			if err == nil {
				t.N = n
			}
		}
	}

	return t, nil
}

func parseSkip(args []string) (Stage, error) {
	s := &Skip{}

	if len(args) > 0 {
		n, err := strconv.Atoi(args[0])
		if err != nil {
			return nil, fmt.Errorf("skip: invalid count %q", args[0])
		}

		s.N = n
	}

	return s, nil
}

func parseSort(args []string) (Stage, error) {
	s := &Sort{}

	for _, arg := range args {
		switch arg {
		case "-r", "--reverse":
			s.Reverse = true
		case "-n", "--numeric", "-rn", "-nr":
			s.Numeric = true
			if arg == "-rn" || arg == "-nr" {
				s.Reverse = true
			}
		}
	}

	return s, nil
}

func parseUniq(args []string) (Stage, error) {
	u := &Uniq{}

	for _, arg := range args {
		if arg == "-i" {
			u.IgnoreCase = true
		}
	}

	return u, nil
}

func parseCut(args []string) (Stage, error) {
	c := &Cut{Delimiter: "\t"}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-d":
			if i+1 < len(args) {
				c.Delimiter = args[i+1]
				i++
			}
		case "-f":
			if i+1 < len(args) {
				fields, err := parseFieldSpec(args[i+1])
				if err != nil {
					return nil, fmt.Errorf("cut: %w", err)
				}

				c.Fields = fields
				i++
			}
		default:
			// Handle -d" " style (delimiter attached to flag)
			if strings.HasPrefix(args[i], "-d") {
				c.Delimiter = args[i][2:]
			} else if strings.HasPrefix(args[i], "-f") {
				fields, err := parseFieldSpec(args[i][2:])
				if err != nil {
					return nil, fmt.Errorf("cut: %w", err)
				}

				c.Fields = fields
			}
		}
	}

	if len(c.Fields) == 0 {
		return nil, fmt.Errorf("cut: missing field specification (-f)")
	}

	return c, nil
}

func parseFieldSpec(spec string) ([]int, error) {
	var fields []int

	for part := range strings.SplitSeq(spec, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		n, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid field %q", part)
		}

		fields = append(fields, n)
	}

	return fields, nil
}

func parseTr(args []string) (Stage, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("tr: requires FROM TO arguments")
	}

	return &Tr{From: args[0], To: args[1]}, nil
}

func parseSed(args []string) (Stage, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("sed: missing expression")
	}

	// Parse s/pattern/replacement/flags format
	expr := args[0]
	if strings.HasPrefix(expr, "s") && len(expr) > 1 {
		delim := rune(expr[1])
		parts := splitSedExpr(expr[2:], delim)

		if len(parts) >= 2 {
			s := &Sed{
				Pattern:     parts[0],
				Replacement: parts[1],
			}

			if len(parts) >= 3 && strings.Contains(parts[2], "g") {
				s.Global = true
			}

			return s, nil
		}
	}

	// Fallback: pattern replacement
	if len(args) >= 2 {
		return &Sed{Pattern: args[0], Replacement: args[1], Global: true}, nil
	}

	return nil, fmt.Errorf("sed: invalid expression %q", expr)
}

func splitSedExpr(s string, delim rune) []string {
	var (
		parts   []string
		current strings.Builder
		escaped bool
	)

	for _, r := range s {
		if escaped {
			current.WriteRune(r)

			escaped = false

			continue
		}

		if r == '\\' {
			escaped = true

			current.WriteRune(r)

			continue
		}

		if r == delim {
			parts = append(parts, current.String())
			current.Reset()

			continue
		}

		current.WriteRune(r)
	}

	parts = append(parts, current.String())

	return parts
}

func parseNl(args []string) (Stage, error) {
	nl := &Nl{Start: 1}

	for i := 0; i < len(args); i++ {
		if args[i] == "-s" && i+1 < len(args) {
			n, err := strconv.Atoi(args[i+1])
			if err != nil {
				return nil, fmt.Errorf("nl: invalid start %q", args[i+1])
			}

			nl.Start = n
			i++
		}
	}

	return nl, nil
}

func parseTee(args []string) (Stage, error) {
	t := &Tee{}

	if len(args) > 0 {
		t.Path = args[0]
	}

	return t, nil
}

func parseWc(args []string) (Stage, error) {
	w := &Wc{}

	for _, arg := range args {
		switch arg {
		case "-l":
			w.Lines = true
		case "-w":
			w.Words = true
		case "-c", "-m":
			w.Chars = true
		}
	}

	return w, nil
}

// parseCommandLine splits a command string into parts, respecting quotes.
func parseCommandLine(cmdLine string) []string {
	var (
		parts   []string
		current strings.Builder
		inQuote rune
		escaped bool
	)

	for _, r := range cmdLine {
		if escaped {
			current.WriteRune(r)

			escaped = false

			continue
		}

		if r == '\\' {
			escaped = true
			continue
		}

		if inQuote != 0 {
			if r == inQuote {
				inQuote = 0
			} else {
				current.WriteRune(r)
			}

			continue
		}

		if r == '"' || r == '\'' {
			inQuote = r
			continue
		}

		if r == ' ' || r == '\t' {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}

			continue
		}

		current.WriteRune(r)
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}
