package procutil

import "testing"

// highUnusedPID is a PID large enough that no real process should own it, so
// Kill/sendImpl exercise the "open process" path and fail with not-found —
// without ever terminating a real process.
const highUnusedPID = 0x7FFFFFF0

// TestKillSendImplErrorPaths drives Kill (pid > 0) into sendImpl for every
// signal. On Windows TERM/KILL hit OpenProcess (which fails for a non-existent
// PID), and INT/HUP hit the unsupported-signal branch. We assert only that an
// error is returned (the PID does not exist), never that a process died.
func TestKillSendImplErrorPaths(t *testing.T) {
	cases := []struct {
		name string
		sig  Signal
	}{
		{"term", SigTerm},
		{"kill", SigKill},
		{"int", SigInt},
		{"hup", SigHup},
		{"unsupported", Signal("BOGUS")},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := Kill(highUnusedPID, c.sig)
			// A non-existent PID must never succeed; every branch of sendImpl
			// returns a non-nil error for this PID (open fails on Windows;
			// unsupported/INT/HUP error on every OS that maps them as such).
			if err == nil {
				t.Errorf("Kill(%d, %s) = nil, want error for non-existent PID", highUnusedPID, c.sig)
			}
		})
	}
}

// TestKillRejectsNonPositivePID covers the pid<=0 guard ahead of sendImpl.
func TestKillRejectsNonPositivePID(t *testing.T) {
	for _, pid := range []int{0, -1, -100} {
		if err := Kill(pid, SigTerm); err == nil {
			t.Errorf("Kill(%d) = nil, want invalid-pid error", pid)
		}
	}
}
