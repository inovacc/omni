package procutil

import (
	"os"
	"testing"
)

func TestRuntimeForName(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want Runtime
	}{
		{"go binary basename — not classified by name", "myapp", RuntimeUnknown},
		{"node lowercase", "node", RuntimeNode},
		{"node windows exe", "node.exe", RuntimeNode},
		{"nodejs alias", "nodejs", RuntimeNode},
		{"node uppercase via case-insensitive", "NODE", RuntimeNode},
		{"java", "java", RuntimeJava},
		{"java windows exe", "java.exe", RuntimeJava},
		{"javaw (no-console)", "javaw", RuntimeJava},
		{"javaw windows exe", "javaw.exe", RuntimeJava},
		{"python", "python", RuntimePython},
		{"python.exe", "python.exe", RuntimePython},
		{"python3", "python3", RuntimePython},
		{"python3.11", "python3.11", RuntimePython},
		{"pythonw", "pythonw", RuntimePython},
		{"pythonw.exe windows", "pythonw.exe", RuntimePython},
		{"unrelated binary", "nginx", RuntimeUnknown},
		{"empty string", "", RuntimeUnknown},
		// Defensive: must not classify "node-foo" or "javac" (compiler, not JVM) as runtime.
		{"node-prefix is not node", "node-foo", RuntimeUnknown},
		{"javac is the compiler, not jvm", "javac", RuntimeUnknown},
		{"python-prefixed name", "python-helper", RuntimeUnknown},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := runtimeForName(tc.in); got != tc.want {
				t.Errorf("runtimeForName(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestClassifyExe_SelfBinary(t *testing.T) {
	// Classify the running test binary; under go test it MUST be detected as Go.
	exe, err := os.Executable()
	if err != nil {
		t.Skipf("os.Executable failed: %v", err)
	}
	rt, info := classifyExe(exe)
	if rt != RuntimeGo {
		t.Fatalf("classifyExe(self) = %q, want %q", rt, RuntimeGo)
	}
	if info == nil {
		t.Fatal("classifyExe returned nil GoBinaryInfo for a Go binary")
	}
	if info.GoVersion == "" {
		t.Error("GoBinaryInfo.GoVersion should be populated for a Go binary")
	}
}
