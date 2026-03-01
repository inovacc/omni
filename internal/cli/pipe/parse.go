package pipe

import (
	"slices"
	"strings"
)

// parseCommands splits args into individual commands
func parseCommands(args []string, separator string) []string {
	if separator == "" {
		separator = "|"
	}

	// Join all args for processing
	joined := strings.Join(args, " ")

	// Check for brace syntax: {cmd1}, {cmd2}, {cmd3}
	if strings.Contains(joined, "{") && strings.Contains(joined, "}") {
		return parseBraceCommands(joined)
	}

	// Check if single arg with separators
	if len(args) == 1 && strings.Contains(args[0], separator) {
		parts := strings.Split(args[0], separator)
		commands := make([]string, 0, len(parts))

		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				commands = append(commands, p)
			}
		}

		return commands
	}

	// Check if args contain separator as separate element
	hasSeparator := slices.Contains(args, separator)

	if hasSeparator {
		var (
			commands []string
			current  []string
		)

		for _, arg := range args {
			if arg == separator {
				if len(current) > 0 {
					commands = append(commands, strings.Join(current, " "))
					current = nil
				}
			} else {
				current = append(current, arg)
			}
		}

		if len(current) > 0 {
			commands = append(commands, strings.Join(current, " "))
		}

		return commands
	}

	// If multiple args and at least some have spaces, treat each as a separate command
	// This handles: omni pipe "cat file.txt" "grep pattern" "sort"
	if len(args) > 1 {
		someHaveSpaces := false

		for _, arg := range args {
			if strings.Contains(arg, " ") {
				someHaveSpaces = true

				break
			}
		}

		if someHaveSpaces {
			return args
		}
	}

	// Otherwise, join all args as a single command, then try to split by separator
	if strings.Contains(joined, separator) {
		parts := strings.Split(joined, separator)
		commands := make([]string, 0, len(parts))

		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				commands = append(commands, p)
			}
		}

		return commands
	}

	// Single command - join all args as one command
	return []string{strings.Join(args, " ")}
}

// parseBraceCommands parses commands in {cmd1}, {cmd2} format
func parseBraceCommands(input string) []string {
	var (
		commands []string
		current  strings.Builder
	)

	depth := 0

	for _, r := range input {
		switch r {
		case '{':
			depth++

			if depth == 1 {
				continue // Don't include opening brace
			}
		case '}':
			depth--

			if depth == 0 {
				cmd := strings.TrimSpace(current.String())
				if cmd != "" {
					commands = append(commands, cmd)
				}

				current.Reset()

				continue // Don't include closing brace
			}
		case ',':
			if depth == 0 {
				continue // Skip commas between commands
			}
		}

		if depth > 0 {
			current.WriteRune(r)
		}
	}

	// Handle any remaining content (in case of missing closing brace)
	if current.Len() > 0 {
		cmd := strings.TrimSpace(current.String())
		if cmd != "" {
			commands = append(commands, cmd)
		}
	}

	return commands
}

// parseCommandLine splits a command string into parts, respecting quotes
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
