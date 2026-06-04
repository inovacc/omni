package task

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestShellCommandRunner_WindowsLeadingArgNoInjection guards the
// cmdinjection/task-shellrunner-windows-leading-element finding (HARDENING
// 2026-06-04, HIGH): on Windows a leading argv element containing a cmd.exe
// metacharacter (& | < > ^) must NOT be reparsed as command syntax. The
// supply-02 fix neutralized non-leading elements via delayed expansion but
// wrote args[0] verbatim into the cmd.exe line, so `x&echo ...` in the leading
// token injected a chained command.
//
// The test uses an unambiguous side effect — an `&`-chained output redirection
// that creates a marker FILE — rather than scanning stdout (cmd's
// "not recognized" error echoes the failed command name, which contains the
// payload text, so a substring check would false-positive).
func TestShellCommandRunner_WindowsLeadingArgNoInjection(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("cmd.exe metacharacter injection is Windows-specific")
	}

	dir := t.TempDir()
	marker := filepath.Join(dir, "PWNED.txt")

	// Leading element chains `& echo x > <marker>`. If the runner lets cmd.exe
	// reparse the leading token, the chained redirect creates the marker file.
	// With every argv element routed through delayed expansion, the `&` and `>`
	// are inert (the whole value is one literal program token) and no file is
	// created.
	args := []string{"nosuchprog_omni&echo x>" + marker, "tail"}

	var out bytes.Buffer
	r := NewShellCommandRunner("")
	// An error is expected (the bogus program is not found); we only assert that
	// the injected redirection did not execute.
	_ = r.Run(context.Background(), &out, args)

	if _, err := os.Stat(marker); err == nil {
		t.Fatalf("command injection via leading argv element: chained redirect created %s (cmd.exe reparsed the leading token)", marker)
	}
}
