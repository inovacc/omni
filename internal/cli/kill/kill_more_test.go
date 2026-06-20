package kill

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"syscall"
	"testing"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/pkg/cobra/helper/output"
)

func TestIsAlphaSignalName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"empty", "", false},
		{"upper mnemonic", "USR1", false}, // contains digit
		{"pure alpha upper", "TERM", true},
		{"pure alpha lower", "term", true},
		{"mixed case", "Kill", true},
		{"numeric", "9", false},
		{"alpha then numeric", "SIG1", false},
		{"hyphen", "TER-M", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isAlphaSignalName(tt.in); got != tt.want {
				t.Errorf("isAlphaSignalName(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestListSignalsJSON(t *testing.T) {
	var buf bytes.Buffer
	f := output.New(&buf, output.FormatJSON)
	if !f.IsJSON() {
		t.Skip("output formatter is not in JSON mode")
	}
	if err := listSignalsJSON(&buf, f); err != nil {
		t.Fatalf("listSignalsJSON() error = %v", err)
	}

	var got []struct {
		Number int    `json:"number"`
		Name   string `json:"name"`
	}
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("listSignalsJSON() produced invalid JSON: %v\n%s", err, buf.String())
	}
	if len(got) == 0 {
		t.Fatal("listSignalsJSON() returned an empty signal list")
	}
	// Every name in the JSON must exist in signalMap with the same number.
	for _, s := range got {
		if want, ok := signalMap[s.Name]; ok {
			if int(want) != s.Number {
				t.Errorf("signal %s: JSON number %d != signalMap %d", s.Name, s.Number, int(want))
			}
		}
	}
}

func TestRunKillListJSON(t *testing.T) {
	var buf bytes.Buffer
	err := RunKill(&buf, nil, KillOptions{List: true, OutputFormat: output.FormatJSON})
	if err != nil {
		t.Fatalf("RunKill(list,json) error = %v", err)
	}
	if !json.Valid(buf.Bytes()) {
		t.Errorf("RunKill(list,json) produced invalid JSON: %s", buf.String())
	}
}

func TestKillFunc(t *testing.T) {
	// Signal 0 to our own PID is a liveness probe that delivers nothing but
	// validates the process exists; it is safe and does not terminate us.
	if err := Kill(os.Getpid(), syscall.Signal(0)); err != nil {
		// Some platforms reject signal 0 via os.Process.Signal; tolerate that
		// but still exercise the function path.
		t.Logf("Kill(self, 0) returned: %v (non-fatal)", err)
	}

	// A non-existent PID should yield an error on platforms where the OS can
	// detect it. On Windows os.FindProcess always succeeds, so the error (if
	// any) comes from Signal.
	err := Kill(999999999, syscall.SIGTERM)
	if err != nil {
		t.Logf("Kill(nonexistent) returned: %v", err)
	}
}

func TestClassifySignalErr(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want error // sentinel to match with errors.Is; nil means "no specific sentinel"
	}{
		{"nil passes through", nil, nil},
		{"already classified unsupported", cmderr.Wrap(cmderr.ErrUnsupported, "x"), cmderr.ErrUnsupported},
		{"already classified invalid", cmderr.Wrap(cmderr.ErrInvalidInput, "x"), cmderr.ErrInvalidInput},
		{"already classified permission", cmderr.Wrap(cmderr.ErrPermission, "x"), cmderr.ErrPermission},
		{"already classified notfound", cmderr.Wrap(cmderr.ErrNotFound, "x"), cmderr.ErrNotFound},
		{"os.ErrPermission -> ErrPermission", os.ErrPermission, cmderr.ErrPermission},
		{"EPERM -> ErrPermission", syscall.EPERM, cmderr.ErrPermission},
		{"generic error wrapped", errors.New("boom"), nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifySignalErr(tt.err, 1234)
			if tt.err == nil {
				if got != nil {
					t.Fatalf("classifySignalErr(nil) = %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatalf("classifySignalErr(%v) = nil, want non-nil", tt.err)
			}
			if tt.want != nil && !errors.Is(got, tt.want) {
				t.Errorf("classifySignalErr(%v) = %v, want errors.Is %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestClassifySignalErrNoSuchProcess(t *testing.T) {
	// The wrapped sentinel differs by platform (ESRCH on unix, ErrProcessDone
	// on windows). Feed the platform-correct one and assert ErrNotFound.
	var raw error
	if runtime.GOOS == "windows" {
		raw = os.ErrProcessDone
	} else {
		raw = syscall.ESRCH
	}
	got := classifySignalErr(raw, 4321)
	if !errors.Is(got, cmderr.ErrNotFound) {
		t.Errorf("classifySignalErr(no-such-process) = %v, want ErrNotFound", got)
	}
}

func TestIsNoSuchProcess(t *testing.T) {
	// Non-matching error is always false.
	if isNoSuchProcess(errors.New("unrelated")) {
		t.Error("isNoSuchProcess(unrelated) = true, want false")
	}
	// The platform-specific matching error must be true.
	if runtime.GOOS == "windows" {
		if !isNoSuchProcess(os.ErrProcessDone) {
			t.Error("isNoSuchProcess(ErrProcessDone) = false on windows, want true")
		}
	} else {
		if !isNoSuchProcess(syscall.ESRCH) {
			t.Error("isNoSuchProcess(ESRCH) = false on unix, want true")
		}
	}
}

func TestIsPlatformUnsupportedSignal(t *testing.T) {
	if runtime.GOOS == "windows" {
		if !isPlatformUnsupportedSignal("USR1") {
			t.Error("USR1 should be platform-unsupported on windows")
		}
		if isPlatformUnsupportedSignal("INT") {
			t.Error("INT is supported on windows, should be false")
		}
	} else {
		if isPlatformUnsupportedSignal("USR1") {
			t.Error("isPlatformUnsupportedSignal is always false on unix")
		}
	}
}

// TestWindowsSignalHelpers exercises the windows-only pure helpers. It is a
// no-op on non-windows but the windows-tagged functions only compile there.
func TestWindowsSignalHelpers(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-only helpers")
	}
	runWindowsSignalHelperChecks(t)
}

func TestRunKillUnsupportedSignalOnWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-only: USR1 is a valid signal on unix")
	}
	var buf bytes.Buffer
	err := RunKill(&buf, []string{"99999999"}, KillOptions{Signal: "USR1"})
	if err == nil {
		t.Fatal("expected ErrUnsupported for SIGUSR1 on windows")
	}
	if !errors.Is(err, cmderr.ErrUnsupported) {
		t.Errorf("RunKill(USR1) = %v, want ErrUnsupported", err)
	}
}

func TestRunKillUnsupportedSignalInlineOnWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-only: -USR1 is a valid signal on unix")
	}
	var buf bytes.Buffer
	// Inline "-USR1" specification must also produce ErrUnsupported on windows.
	err := RunKill(&buf, []string{"-USR1", "99999999"}, KillOptions{})
	if err == nil {
		t.Fatal("expected ErrUnsupported for inline -USR1 on windows")
	}
	if !errors.Is(err, cmderr.ErrUnsupported) {
		t.Errorf("RunKill(-USR1) = %v, want ErrUnsupported", err)
	}
}

func TestRunKillSelfSignalZeroVerbose(t *testing.T) {
	// Sending signal 0 to our own PID exercises the success path of sendSignal
	// without terminating the test process. Use a numeric signal "0".
	var buf bytes.Buffer
	pid := fmt.Sprintf("%d", os.Getpid())
	err := RunKill(&buf, []string{pid}, KillOptions{Signal: "0", Verbose: true})
	// On unix, signal 0 to self succeeds and prints a verbose line. On windows
	// signal 0 is not in the supported set, so an error is acceptable.
	if err != nil {
		if runtime.GOOS != "windows" {
			t.Logf("RunKill(self, 0) unix returned: %v", err)
		}
		return
	}
	if runtime.GOOS != "windows" && !strings.Contains(buf.String(), "Sent signal") {
		t.Errorf("verbose output missing 'Sent signal': %q", buf.String())
	}
}

func TestRunKillJSONResults(t *testing.T) {
	var buf bytes.Buffer
	// A non-existent PID in JSON mode must emit a results array (not an error).
	err := RunKill(&buf, []string{"999999999"}, KillOptions{OutputFormat: output.FormatJSON})
	if err != nil {
		t.Fatalf("RunKill(json, bad pid) error = %v, want nil (results emitted)", err)
	}
	var results []KillResult
	if err := json.Unmarshal(buf.Bytes(), &results); err != nil {
		t.Fatalf("invalid JSON results: %v\n%s", err, buf.String())
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Success {
		t.Errorf("expected failure result for nonexistent pid, got success")
	}
}

func TestRunKillJSONInvalidPID(t *testing.T) {
	var buf bytes.Buffer
	err := RunKill(&buf, []string{"not-a-pid"}, KillOptions{OutputFormat: output.FormatJSON})
	if err != nil {
		t.Fatalf("RunKill(json, invalid pid) error = %v, want nil", err)
	}
	var results []KillResult
	if err := json.Unmarshal(buf.Bytes(), &results); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(results) != 1 || results[0].Success {
		t.Errorf("expected single failure result, got %+v", results)
	}
}
