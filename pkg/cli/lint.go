package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// LintOptions configures the lint command behavior.
type LintOptions struct {
	Format string // Output format: text, json
	Fix    bool   // Auto-fix issues where possible
	Strict bool   // Enable strict mode (more warnings)
	Quiet  bool   // Only show errors, not warnings
}

// LintSeverity represents the severity of a lint issue.
type LintSeverity string

const (
	LintError   LintSeverity = "error"
	LintWarning LintSeverity = "warning"
	LintInfo    LintSeverity = "info"
)

// LintIssue represents a single lint issue.
type LintIssue struct {
	File     string       `json:"file"`
	Line     int          `json:"line,omitempty"`
	Column   int          `json:"column,omitempty"`
	Severity LintSeverity `json:"severity"`
	Rule     string       `json:"rule"`
	Message  string       `json:"message"`
	Fix      string       `json:"fix,omitempty"`
}

// LintResult contains all lint issues for a file.
type LintResult struct {
	File       string      `json:"file"`
	Issues     []LintIssue `json:"issues"`
	ErrorCount int         `json:"error_count"`
	WarnCount  int         `json:"warning_count"`
}

// Common shell commands that should use omni
var shellCommands = map[string]string{
	"ls":        "omni ls",
	"cat":       "omni cat",
	"cp":        "omni cp",
	"mv":        "omni mv",
	"rm":        "omni rm",
	"mkdir":     "omni mkdir",
	"rmdir":     "omni rmdir",
	"touch":     "omni touch",
	"chmod":     "omni chmod",
	"chown":     "omni chown",
	"grep":      "omni grep",
	"sed":       "omni sed",
	"awk":       "omni awk",
	"head":      "omni head",
	"tail":      "omni tail",
	"sort":      "omni sort",
	"uniq":      "omni uniq",
	"wc":        "omni wc",
	"cut":       "omni cut",
	"tr":        "omni tr",
	"base64":    "omni base64",
	"tar":       "omni tar",
	"zip":       "omni zip",
	"unzip":     "omni unzip",
	"sha256sum": "omni sha256sum",
	"sha512sum": "omni sha512sum",
	"md5sum":    "omni md5sum",
	"uname":     "omni uname",
	"whoami":    "omni whoami",
	"pwd":       "omni pwd",
	"date":      "omni date",
	"env":       "omni env",
	"df":        "omni df",
	"du":        "omni du",
	"ps":        "omni ps",
	"kill":      "omni kill",
	"diff":      "omni diff",
	"stat":      "omni stat",
	"ln":        "omni ln",
	"readlink":  "omni readlink",
	"realpath":  "omni realpath",
	"dirname":   "omni dirname",
	"basename":  "omni basename",
	"nl":        "omni nl",
	"fold":      "omni fold",
	"join":      "omni join",
	"paste":     "omni paste",
	"column":    "omni column",
	"tac":       "omni tac",
	"yes":       "omni yes",
	"xargs":     "omni xargs",
}

// Non-portable commands that should be avoided
var nonPortableCommands = map[string]string{
	"find":      "Use 'omni ls -R' or Go filepath.Walk",
	"xdg-open":  "Not available on Windows/macOS",
	"open":      "macOS-specific, not available on Linux/Windows",
	"start":     "Windows-specific",
	"apt":       "Debian/Ubuntu-specific",
	"apt-get":   "Debian/Ubuntu-specific",
	"yum":       "RHEL/CentOS-specific",
	"dnf":       "Fedora-specific",
	"pacman":    "Arch-specific",
	"brew":      "macOS-specific",
	"choco":     "Windows-specific",
	"snap":      "Linux-specific",
	"systemctl": "Linux-specific",
	"service":   "Platform-specific init system",
	"launchctl": "macOS-specific",
	"sc":        "Windows-specific service control",
}

// Bash-specific syntax patterns
var bashSpecificPatterns = []struct {
	Pattern *regexp.Regexp
	Message string
	Fix     string
}{
	{
		Pattern: regexp.MustCompile(`\[\[.*\]\]`),
		Message: "Double brackets [[ ]] are bash-specific",
		Fix:     "Use single brackets [ ] for POSIX compatibility",
	},
	{
		Pattern: regexp.MustCompile(`\$\{[^}]+:[-+=?][^}]*\}`),
		Message: "Parameter expansion with :- := :+ :? is bash-specific",
		Fix:     "Use explicit conditionals or default values",
	},
	{
		Pattern: regexp.MustCompile(`\$\([^)]+\)`),
		Message: "Command substitution $() - prefer cross-platform alternatives",
		Fix:     "Use omni commands or Go code instead",
	},
	{
		Pattern: regexp.MustCompile(`\|&`),
		Message: "|& pipe is bash-specific",
		Fix:     "Use 2>&1 | for POSIX compatibility",
	},
	{
		Pattern: regexp.MustCompile(`<<<`),
		Message: "Here-strings <<< are bash-specific",
		Fix:     "Use echo | or here-documents",
	},
	{
		Pattern: regexp.MustCompile(`\bfunction\s+\w+`),
		Message: "function keyword is bash-specific",
		Fix:     "Use fname() { } syntax",
	},
	{
		Pattern: regexp.MustCompile(`\bsource\b`),
		Message: "source is bash-specific",
		Fix:     "Use . (dot) for POSIX compatibility",
	},
	{
		Pattern: regexp.MustCompile(`\bpushd\b|\bpopd\b`),
		Message: "pushd/popd are bash-specific",
		Fix:     "Use cd with saved paths",
	},
}

// RunLint executes the lint command on Taskfiles.
func RunLint(w io.Writer, args []string, opts LintOptions) error {
	if len(args) == 0 {
		// Default to Taskfile.yml in current directory
		args = []string{"Taskfile.yml"}
	}

	var allResults []LintResult

	exitCode := 0

	for _, arg := range args {
		// Check if it's a directory
		info, err := os.Stat(arg)
		if err != nil {
			_, _ = fmt.Fprintf(w, "lint: %s: %v\n", arg, err)
			exitCode = 1

			continue
		}

		var files []string
		if info.IsDir() {
			// Find Taskfile.yml files
			files = findTaskfiles(arg)
		} else {
			files = []string{arg}
		}

		for _, file := range files {
			result, err := lintTaskfile(file, opts)
			if err != nil {
				_, _ = fmt.Fprintf(w, "lint: %s: %v\n", file, err)
				exitCode = 1

				continue
			}

			allResults = append(allResults, result)

			if result.ErrorCount > 0 {
				exitCode = 1
			}
		}
	}

	// Output results
	for _, result := range allResults {
		if len(result.Issues) == 0 {
			if !opts.Quiet {
				_, _ = fmt.Fprintf(w, "%s: OK\n", result.File)
			}

			continue
		}

		for _, issue := range result.Issues {
			if opts.Quiet && issue.Severity != LintError {
				continue
			}

			location := result.File
			if issue.Line > 0 {
				location = fmt.Sprintf("%s:%d", result.File, issue.Line)
			}

			severity := string(issue.Severity)
			switch issue.Severity {
			case LintError:
				severity = "\033[31merror\033[0m"
			case LintWarning:
				severity = "\033[33mwarning\033[0m"
			case LintInfo:
				severity = "\033[36minfo\033[0m"
			}

			_, _ = fmt.Fprintf(w, "%s: %s: [%s] %s\n", location, severity, issue.Rule, issue.Message)
			if issue.Fix != "" && !opts.Quiet {
				_, _ = fmt.Fprintf(w, "  fix: %s\n", issue.Fix)
			}
		}

		_, _ = fmt.Fprintf(w, "\n%s: %d error(s), %d warning(s)\n\n",
			result.File, result.ErrorCount, result.WarnCount)
	}

	if exitCode != 0 {
		return fmt.Errorf("lint found issues")
	}

	return nil
}

func findTaskfiles(dir string) []string {
	var files []string

	patterns := []string{"Taskfile.yml", "Taskfile.yaml", "taskfile.yml", "taskfile.yaml"}

	for _, pattern := range patterns {
		path := filepath.Join(dir, pattern)
		if _, err := os.Stat(path); err == nil {
			files = append(files, path)
		}
	}

	// Also check subdirectories for included taskfiles
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil //nolint:nilerr // Intentionally skip errors
		}

		if info.IsDir() {
			return nil
		}

		base := filepath.Base(path)
		for _, pattern := range patterns {
			if base == pattern {
				// Avoid duplicates
				found := slices.Contains(files, path)

				if !found {
					files = append(files, path)
				}
			}
		}

		return nil
	})

	return files
}

func lintTaskfile(filename string, opts LintOptions) (LintResult, error) {
	result := LintResult{
		File:   filename,
		Issues: []LintIssue{},
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return result, err
	}

	// Parse YAML
	var taskfile map[string]any
	if err := yaml.Unmarshal(data, &taskfile); err != nil {
		result.Issues = append(result.Issues, LintIssue{
			File:     filename,
			Severity: LintError,
			Rule:     "yaml-parse",
			Message:  fmt.Sprintf("Invalid YAML: %v", err),
		})
		result.ErrorCount++

		return result, nil
	}

	// Read file for line-by-line analysis
	lines, _ := readFileLines(filename)

	// Check for version
	if _, ok := taskfile["version"]; !ok {
		result.Issues = append(result.Issues, LintIssue{
			File:     filename,
			Severity: LintWarning,
			Rule:     "missing-version",
			Message:  "Taskfile should specify version",
			Fix:      "Add 'version: \"3\"' at the top",
		})
		result.WarnCount++
	}

	// Check tasks
	tasks, ok := taskfile["tasks"].(map[string]any)
	if !ok {
		result.Issues = append(result.Issues, LintIssue{
			File:     filename,
			Severity: LintError,
			Rule:     "missing-tasks",
			Message:  "No tasks defined",
		})
		result.ErrorCount++

		return result, nil
	}

	for taskName, taskDef := range tasks {
		taskMap, ok := taskDef.(map[string]any)
		if !ok {
			continue
		}

		// Check commands
		cmds, ok := taskMap["cmds"].([]any)
		if !ok {
			continue
		}

		for _, cmdDef := range cmds {
			var cmd string

			switch c := cmdDef.(type) {
			case string:
				cmd = c
			case map[string]any:
				if cmdStr, ok := c["cmd"].(string); ok {
					cmd = cmdStr
				}
			}

			if cmd == "" {
				continue
			}

			// Find line number for this command
			lineNum := findLineNumber(lines, cmd)

			// Check for shell commands that should use omni
			issues := checkCommand(filename, taskName, cmd, lineNum, opts)
			for _, issue := range issues {
				result.Issues = append(result.Issues, issue)
				switch issue.Severity {
				case LintError:
					result.ErrorCount++
				case LintWarning:
					result.WarnCount++
				case LintInfo:
					// Info level doesn't count towards error/warning totals
				}
			}
		}
	}

	return result, nil
}

func readFileLines(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = file.Close()
	}()

	var lines []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

func findLineNumber(lines []string, search string) int {
	// Escape special characters for search
	search = strings.TrimSpace(search)
	if search == "" {
		return 0
	}

	for i, line := range lines {
		if strings.Contains(line, search) || strings.Contains(line, strings.Split(search, " ")[0]) {
			return i + 1
		}
	}

	return 0
}

func checkCommand(filename, taskName, cmd string, lineNum int, opts LintOptions) []LintIssue {
	var issues []LintIssue

	// Tokenize command
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return issues
	}

	baseCmd := filepath.Base(parts[0])

	// Check for shell commands that should use omni
	if replacement, ok := shellCommands[baseCmd]; ok {
		// Don't flag if already using omni
		if !strings.HasPrefix(parts[0], "omni") {
			issues = append(issues, LintIssue{
				File:     filename,
				Line:     lineNum,
				Severity: LintWarning,
				Rule:     "use-omni",
				Message:  fmt.Sprintf("Task '%s': '%s' could be replaced with '%s' for portability", taskName, baseCmd, replacement),
				Fix:      fmt.Sprintf("Replace '%s' with '%s'", baseCmd, replacement),
			})
		}
	}

	// Check for non-portable commands
	if msg, ok := nonPortableCommands[baseCmd]; ok {
		issues = append(issues, LintIssue{
			File:     filename,
			Line:     lineNum,
			Severity: LintError,
			Rule:     "non-portable",
			Message:  fmt.Sprintf("Task '%s': '%s' is non-portable: %s", taskName, baseCmd, msg),
		})
	}

	// Check for bash-specific syntax
	for _, pattern := range bashSpecificPatterns {
		if pattern.Pattern.MatchString(cmd) {
			severity := LintWarning
			if opts.Strict {
				severity = LintError
			}

			issues = append(issues, LintIssue{
				File:     filename,
				Line:     lineNum,
				Severity: severity,
				Rule:     "bash-specific",
				Message:  fmt.Sprintf("Task '%s': %s", taskName, pattern.Message),
				Fix:      pattern.Fix,
			})
		}
	}

	// Check for pipe chains that could fail silently
	if strings.Contains(cmd, "|") && !strings.Contains(cmd, "set -o pipefail") {
		if opts.Strict {
			issues = append(issues, LintIssue{
				File:     filename,
				Line:     lineNum,
				Severity: LintInfo,
				Rule:     "pipe-safety",
				Message:  fmt.Sprintf("Task '%s': Pipe chain without pipefail may hide errors", taskName),
				Fix:      "Use 'set -o pipefail' or omni commands",
			})
		}
	}

	// Check for hardcoded paths
	if strings.Contains(cmd, "/usr/") || strings.Contains(cmd, "/bin/") || strings.Contains(cmd, "/etc/") {
		issues = append(issues, LintIssue{
			File:     filename,
			Line:     lineNum,
			Severity: LintWarning,
			Rule:     "hardcoded-path",
			Message:  fmt.Sprintf("Task '%s': Hardcoded Unix path may not work on Windows", taskName),
			Fix:      "Use relative paths or environment variables",
		})
	}

	// Check for Windows-incompatible path separators in strings
	if strings.Contains(cmd, ":/") && !strings.Contains(cmd, "://") {
		// Likely a Windows path issue (but ignore URLs)
		issues = append(issues, LintIssue{
			File:     filename,
			Line:     lineNum,
			Severity: LintInfo,
			Rule:     "path-separator",
			Message:  fmt.Sprintf("Task '%s': Consider using filepath-safe path construction", taskName),
		})
	}

	return issues
}
