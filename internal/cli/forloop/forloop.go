package forloop

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

// Options configures the for command behavior
type Options struct {
	Delimiter string // delimiter for splitting input
	Variable  string // variable name to use (default: item)
	DryRun    bool   // print commands without executing
	Parallel  int    // number of parallel executions (0 = sequential)
}

// RunRange executes a command for each number in a range
// Usage: for range START END [STEP] -- COMMAND
func RunRange(w io.Writer, start, end, step int, command string, opts Options) error {
	if step == 0 {
		step = 1
		if start > end {
			step = -1
		}
	}

	varName := opts.Variable
	if varName == "" {
		varName = "i"
	}

	for i := start; ; {
		if step > 0 && i > end {
			break
		}

		if step < 0 && i < end {
			break
		}

		bindings := []envBinding{{name: varName, value: strconv.Itoa(i)}}

		if opts.DryRun {
			_, _ = fmt.Fprintf(w, "%s\n", renderDryRun(command, bindings))
		} else {
			if err := executeCommand(w, command, bindings); err != nil {
				return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("for range: command failed at %d: %s", i, err))
			}
		}

		i += step
	}

	return nil
}

// RunEach executes a command for each item in a list
// Usage: for each ITEM1 ITEM2 ... -- COMMAND
func RunEach(w io.Writer, items []string, command string, opts Options) error {
	varName := opts.Variable
	if varName == "" {
		varName = "item"
	}

	for _, item := range items {
		bindings := []envBinding{{name: varName, value: item}}

		if opts.DryRun {
			_, _ = fmt.Fprintf(w, "%s\n", renderDryRun(command, bindings))
		} else {
			if err := executeCommand(w, command, bindings); err != nil {
				return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("for each: command failed for %q: %s", item, err))
			}
		}
	}

	return nil
}

// RunLines executes a command for each line from stdin or file
// Usage: for lines [FILE] -- COMMAND
func RunLines(w io.Writer, r io.Reader, command string, opts Options) error {
	varName := opts.Variable
	if varName == "" {
		varName = "line"
	}

	scanner := bufio.NewScanner(r)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		bindings := []envBinding{
			{name: varName, value: line},
			{name: "n", value: strconv.Itoa(lineNum)},
		}

		if opts.DryRun {
			_, _ = fmt.Fprintf(w, "%s\n", renderDryRun(command, bindings))
		} else {
			if err := executeCommand(w, command, bindings); err != nil {
				return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("for lines: command failed at line %d: %s", lineNum, err))
			}
		}
	}

	return scanner.Err()
}

// RunSplit executes a command for each token split by delimiter
// Usage: for split DELIM INPUT -- COMMAND
func RunSplit(w io.Writer, input, delimiter, command string, opts Options) error {
	varName := opts.Variable
	if varName == "" {
		varName = "item"
	}

	items := strings.Split(input, delimiter)

	for i, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}

		bindings := []envBinding{
			{name: varName, value: item},
			{name: "i", value: strconv.Itoa(i)},
		}

		if opts.DryRun {
			_, _ = fmt.Fprintf(w, "%s\n", renderDryRun(command, bindings))
		} else {
			if err := executeCommand(w, command, bindings); err != nil {
				return fmt.Errorf("for split: command failed for %q: %w", item, err)
			}
		}
	}

	return nil
}

// RunGlob executes a command for each file matching a glob pattern
// Usage: for glob PATTERN -- COMMAND
func RunGlob(w io.Writer, pattern, command string, opts Options) error {
	varName := opts.Variable
	if varName == "" {
		varName = "file"
	}

	// Use filepath.Glob for pattern matching
	matches, err := globFiles(pattern)
	if err != nil {
		return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("for glob: %s", err))
	}

	for _, match := range matches {
		bindings := []envBinding{{name: varName, value: match}}

		if opts.DryRun {
			_, _ = fmt.Fprintf(w, "%s\n", renderDryRun(command, bindings))
		} else {
			if err := executeCommand(w, command, bindings); err != nil {
				return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("for glob: command failed for %q: %s", match, err))
			}
		}
	}

	return nil
}

// envBinding is a single loop-variable binding (name -> value) for one iteration.
//
// Security: the per-iteration value (range index, list item, line content, split
// token, or glob match) is NEVER string-concatenated into the shell command text.
// Doing so would let a value such as `$(rm -rf ~)` or `foo; rm -rf /` be reparsed
// by sh -c / cmd /C as command syntax (command injection). Instead the value is
// exported into the child process environment and the command's ${var}/$var
// references are rewritten to the shell's NATIVE expansion form ($var on POSIX,
// %var% on Windows cmd), so the shell expands the value from the environment as a
// plain string and never interprets it as code.
type envBinding struct {
	name  string
	value string
}

// replaceVariable replaces ${var} and $var with value.
//
// This is used ONLY for dry-run rendering (a human-readable preview), never to
// build the command that is actually executed. See executeCommand / shellRewrite
// for the injection-safe execution path.
func replaceVariable(cmd, varName, value string) string {
	// Replace ${var}
	cmd = strings.ReplaceAll(cmd, "${"+varName+"}", value)
	// Replace $var (word boundary aware)
	cmd = replaceWordVariable(cmd, "$"+varName, value)

	return cmd
}

// renderDryRun produces the resolved-command preview shown by --dry-run. It mirrors
// the literal values for human inspection only; the real execution path keeps the
// values out of the command string entirely (passed via the environment instead).
func renderDryRun(command string, bindings []envBinding) string {
	out := command
	for _, b := range bindings {
		out = replaceVariable(out, b.name, b.value)
	}

	return out
}

// replaceWordVariable replaces $var considering word boundaries
func replaceWordVariable(s, old, new string) string {
	result := s

	for {
		idx := strings.Index(result, old)
		if idx == -1 {
			break
		}

		// Check if next char is alphanumeric (part of variable name)
		endIdx := idx + len(old)
		if endIdx < len(result) {
			next := result[endIdx]
			if isAlphaNumeric(next) {
				// Skip this, it's part of a longer variable name
				// Find after this occurrence
				result = result[:idx] + "\x00SKIP\x00" + result[idx+1:]
				continue
			}
		}

		result = result[:idx] + new + result[endIdx:]
	}

	// Restore skipped $
	result = strings.ReplaceAll(result, "\x00SKIP\x00", "$")

	return result
}

func isAlphaNumeric(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_'
}

// executeCommand runs a shell command with the loop bindings supplied via the
// process environment rather than concatenated into the command string.
//
// Injection safety: bindings[i].value is exported as an environment variable and
// the command's variable references are rewritten to the shell's native expansion
// form (see shellRewrite). The shell expands those references to the literal value
// at runtime, so even a value containing shell metacharacters (e.g. `;`, `$()`,
// backticks, `&&`) is treated as inert data, not as additional commands.
//
// supply-01 (Windows): cmd.exe is invoked with `/V:ON` to enable DELAYED
// environment-variable expansion for this invocation only. The rewritten template
// references values as `!var!` rather than `%var%`. The distinction is the entire
// security boundary on Windows:
//   - `%var%` is substituted at PARSE TIME, before tokenization. A value like
//     `x & type nul > MARKER` would be spliced into the command line and its `&`
//     reparsed as a command separator (injection).
//   - `!var!` is substituted at RUNTIME, after the line is tokenized, so the value
//     lands as a single inert token; its metacharacters are not reparsed as
//     command syntax.
//
// Known limitation: delayed expansion treats `!` as special, so a loop VALUE (or
// the template) containing a literal `!` may be altered (the `!...!` span is
// interpreted as a variable reference). This is an accepted tradeoff: injection
// safety outweighs faithful handling of literal `!` in untrusted loop values.
func executeCommand(w io.Writer, command string, bindings []envBinding) error {
	var cmd *exec.Cmd

	// Use shell to execute. Rewrite the (trusted, user-provided) command template
	// so its $var/${var} references resolve from the environment in the native
	// syntax of the target shell.
	if isWindows() {
		cmd = exec.Command("cmd", "/V:ON", "/C", shellRewrite(command, bindings, true))
	} else {
		cmd = exec.Command("sh", "-c", shellRewrite(command, bindings, false))
	}

	cmd.Env = os.Environ()
	for _, b := range bindings {
		cmd.Env = append(cmd.Env, b.name+"="+b.value)
	}

	cmd.Stdout = w
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// shellRewrite rewrites the command template's ${var}/$var loop-variable references
// into the native expansion syntax of the target shell so the shell — not omni —
// substitutes the value from the exported environment variable.
//
//   - POSIX sh: $var / ${var} are already native; left unchanged.
//   - Windows cmd: $var / ${var} are rewritten to !var! (DELAYED expansion form,
//     enabled via `cmd /V:ON`). See supply-01 in executeCommand: %var% expands at
//     parse time and is injectable, whereas !var! expands after tokenization and
//     is inert.
//
// Only the known loop variable names are rewritten; any other text (including
// unrelated $ usage) is left untouched.
func shellRewrite(command string, bindings []envBinding, windows bool) string {
	if !windows {
		// POSIX shells expand $var/${var} from the environment natively.
		return command
	}

	out := command
	for _, b := range bindings {
		out = strings.ReplaceAll(out, "${"+b.name+"}", "!"+b.name+"!")
		out = replaceWordVariable(out, "$"+b.name, "!"+b.name+"!")
	}

	return out
}

func isWindows() bool {
	return os.PathSeparator == '\\'
}

// globFiles returns files matching the pattern
func globFiles(pattern string) ([]string, error) {
	// Simple glob using filepath
	return filepathGlob(pattern)
}
