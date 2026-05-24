package obfuscate

import (
	"os"
	"runtime"
	"testing"
)

func TestDetect_SelfBinaryIsClean(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Skipf("os.Executable: %v", err)
	}
	v, err := Detect(exe)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if !v.BuildInfoFound {
		t.Error("self test binary should have build info")
	}
	if v.Verdict != VerdictClean {
		t.Errorf("self test binary verdict = %q, want %q", v.Verdict, VerdictClean)
	}
	// On non-Windows platforms we also expect SymbolsFound=true.
	if runtime.GOOS != "windows" && !v.SymbolsFound {
		t.Error("non-Windows self binary should have Go symbol sections")
	}
}

func TestDetect_NotAGoBinary(t *testing.T) {
	// Write a fake "binary" that contains no Go markers at all.
	f, err := os.CreateTemp(t.TempDir(), "not-go-*")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write([]byte("this is not a binary file at all")); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	v, err := Detect(f.Name())
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if v.Verdict != VerdictNotGo {
		t.Errorf("non-binary verdict = %q, want %q", v.Verdict, VerdictNotGo)
	}
	if v.BuildInfoFound {
		t.Error("non-binary should not have buildinfo")
	}
}

func TestDetect_MissingPath(t *testing.T) {
	_, err := Detect("/no/such/path/definitely-does-not-exist-xyz")
	if err == nil {
		t.Error("Detect on missing path should error")
	}
}

func TestGarbleNameRE_RecognizesMangledPaths(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"github.com/x/y", false},
		{"_abc123/main", true},                  // leading mangled segment
		{"github.com/_abc123/y", true},          // mid-path mangled segment
		{"github.com/x/y._abc123", true},        // dot-prefixed mangled identifier
		{"_short", false},                       // < 6 chars after underscore — not mangled
		{"_AaAaAa", false},                      // garble uses lowercase-alnum only
		{"github.com/_aaaaaa/y", true},          // exactly 6 chars
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			if got := garbleNameRE.MatchString(tc.in); got != tc.want {
				t.Errorf("garbleNameRE.MatchString(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}
