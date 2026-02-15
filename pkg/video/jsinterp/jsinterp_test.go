package jsinterp

import (
	"testing"
)

func TestExecute(t *testing.T) {
	interp := New()

	v, err := interp.Execute("1 + 2")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if v.ToInteger() != 3 {
		t.Errorf("got %d, want 3", v.ToInteger())
	}
}

func TestExecuteError(t *testing.T) {
	interp := New()

	_, err := interp.Execute("throw new Error('boom')")
	if err == nil {
		t.Error("expected error")
	}
}

func TestCallFunction(t *testing.T) {
	interp := New()

	code := `function add(a, b) { return a + b; }`
	v, err := interp.CallFunction(code, "add", 3, 4)
	if err != nil {
		t.Fatalf("CallFunction: %v", err)
	}

	if v.ToInteger() != 7 {
		t.Errorf("got %d, want 7", v.ToInteger())
	}
}

func TestCallFunctionNotFound(t *testing.T) {
	interp := New()

	_, err := interp.CallFunction("var x = 1;", "nonexistent")
	if err == nil {
		t.Error("expected error for missing function")
	}
}

func TestCallFunctionBadCode(t *testing.T) {
	interp := New()

	_, err := interp.CallFunction("invalid js {{{{", "fn")
	if err == nil {
		t.Error("expected error for bad code")
	}
}

func TestExtractFunction(t *testing.T) {
	interp := New()

	code := `function reverse(s) { return s.split('').reverse().join(''); }`
	fn, err := interp.ExtractFunction(code, "reverse")
	if err != nil {
		t.Fatalf("ExtractFunction: %v", err)
	}

	result, err := fn("hello")
	if err != nil {
		t.Fatalf("calling extracted function: %v", err)
	}

	if result != "olleh" {
		t.Errorf("got %q, want %q", result, "olleh")
	}
}

func TestExtractFunctionNotFound(t *testing.T) {
	interp := New()

	_, err := interp.ExtractFunction("var x = 1;", "nope")
	if err == nil {
		t.Error("expected error")
	}
}

func TestSetAndGetString(t *testing.T) {
	interp := New()

	if err := interp.Set("myVar", "hello"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	got := interp.GetString("myVar")
	if got != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}
}

func TestGetStringUndefined(t *testing.T) {
	interp := New()

	got := interp.GetString("nonexistent")
	if got != "" {
		t.Errorf("got %q, want empty string", got)
	}
}

func TestSetNumber(t *testing.T) {
	interp := New()

	_ = interp.Set("num", 42)
	v, _ := interp.Execute("num * 2")
	if v.ToInteger() != 84 {
		t.Errorf("got %d, want 84", v.ToInteger())
	}
}
