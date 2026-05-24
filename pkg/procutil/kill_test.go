package procutil

import (
	"strings"
	"testing"
)

func TestParseSignal(t *testing.T) {
	cases := []struct {
		in      string
		want    Signal
		wantErr bool
	}{
		{"", SigTerm, false},
		{"TERM", SigTerm, false},
		{"term", SigTerm, false},
		{"  TERM  ", SigTerm, false},
		{"KILL", SigKill, false},
		{"kill", SigKill, false},
		{"INT", SigInt, false},
		{"HUP", SigHup, false},
		{"SIGTERM", "", true},
		{"9", "", true},
		{"bogus", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := ParseSignal(tc.in)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ParseSignal(%q) err = %v, wantErr = %v", tc.in, err, tc.wantErr)
			}
			if !tc.wantErr && got != tc.want {
				t.Errorf("ParseSignal(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestKill_InvalidPID(t *testing.T) {
	// pid <= 0 must error before any syscall — exercises the early-return guard.
	if err := Kill(0, SigTerm); err == nil {
		t.Error("Kill(0, ...) should error")
	}
	if err := Kill(-5, SigTerm); err == nil {
		t.Error("Kill(-5, ...) should error")
	}
}

func TestKillAllMatching_EmptyTarget(t *testing.T) {
	if _, err := KillAllMatching(nil, "", SigTerm, RuntimeGo); err == nil || !strings.Contains(err.Error(), "target") {
		t.Errorf("KillAllMatching('') should error mentioning 'target'; got %v", err)
	}
}
