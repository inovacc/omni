package gopsagent

import "testing"

// TestPrivileged_OpGCIsGated proves that OpGC (forced process-global
// stop-the-world runtime.GC()) is treated as a privileged opcode and therefore
// requires a configured AuthKey. Without this gate, any local process can spam
// the default no-AuthKey loopback agent with forced GCs (CWE-400 DoS), exactly
// as OpSetGCPercent is already gated.
func TestPrivileged_OpGCIsGated(t *testing.T) {
	if !privileged(OpGC) {
		t.Errorf("privileged(OpGC) = false, want true: forced GC must require an AuthKey")
	}
}

// TestPrivileged_StateChangingOpcodesAreGated guards the full set of
// state-changing / disclosing opcodes so a future edit cannot silently drop
// one back into the unauthenticated surface.
func TestPrivileged_StateChangingOpcodesAreGated(t *testing.T) {
	gated := []struct {
		name string
		op   byte
	}{
		{"OpShutdown", OpShutdown},
		{"OpSetGCPercent", OpSetGCPercent},
		{"OpCPUProfile", OpCPUProfile},
		{"OpTrace", OpTrace},
		{"OpHeapProfile", OpHeapProfile},
		{"OpStack", OpStack},
		{"OpGC", OpGC},
	}
	for _, tc := range gated {
		if !privileged(tc.op) {
			t.Errorf("privileged(%s) = false, want true", tc.name)
		}
	}

	// Read-only introspection opcodes must remain unauthenticated.
	readonly := []struct {
		name string
		op   byte
	}{
		{"OpVersion", OpVersion},
		{"OpMemStats", OpMemStats},
		{"OpStats", OpStats},
		{"OpRuntimeSnapshot", OpRuntimeSnapshot},
	}
	for _, tc := range readonly {
		if privileged(tc.op) {
			t.Errorf("privileged(%s) = true, want false: read-only opcode must stay unauthenticated", tc.name)
		}
	}
}
