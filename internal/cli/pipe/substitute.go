package pipe

import "strings"

// substituteVariables replaces variable placeholders in command string with actual values
// Supports:
//   - $VAR or ${VAR} - single value substitution (uses last line of output)
//   - [$VAR...] - iteration over all lines
func substituteVariables(cmdStr, output, varName string) ([]string, bool) {
	if varName == "" {
		varName = "OUT"
	}

	// Check for iteration pattern: [$VAR...]
	iterPattern := "[" + "$" + varName + "...]"
	iterPatternBrace := "[" + "${" + varName + "}...]"

	if strings.Contains(cmdStr, iterPattern) || strings.Contains(cmdStr, iterPatternBrace) {
		// Split output into lines and create command for each
		lines := strings.Split(strings.TrimSpace(output), "\n")
		commands := make([]string, 0, len(lines))

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			cmd := strings.ReplaceAll(cmdStr, iterPattern, line)
			cmd = strings.ReplaceAll(cmd, iterPatternBrace, line)
			commands = append(commands, cmd)
		}

		return commands, true
	}

	// Single value substitution: $VAR or ${VAR}
	singlePattern := "$" + varName
	singlePatternBrace := "${" + varName + "}"

	if strings.Contains(cmdStr, singlePattern) || strings.Contains(cmdStr, singlePatternBrace) {
		// Use last non-empty line as value
		lines := strings.Split(strings.TrimSpace(output), "\n")
		value := ""

		for i := len(lines) - 1; i >= 0; i-- {
			if strings.TrimSpace(lines[i]) != "" {
				value = strings.TrimSpace(lines[i])

				break
			}
		}

		cmd := strings.ReplaceAll(cmdStr, singlePatternBrace, value)
		cmd = strings.ReplaceAll(cmd, singlePattern, value)

		return []string{cmd}, false
	}

	return []string{cmdStr}, false
}
