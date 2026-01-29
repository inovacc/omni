package nanoid

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	nanoid, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if len(nanoid) != defaultLength {
		t.Errorf("NanoID length = %d, want %d", len(nanoid), defaultLength)
	}
}

func TestNanoIDUniqueness(t *testing.T) {
	seen := make(map[string]bool)
	count := 1000

	for i := 0; i < count; i++ {
		nanoid, err := New()
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}

		if seen[nanoid] {
			t.Errorf("Duplicate NanoID generated: %s", nanoid)
		}
		seen[nanoid] = true
	}
}

func TestGenerateCustomLength(t *testing.T) {
	lengths := []int{5, 10, 15, 30, 50}

	for _, length := range lengths {
		nanoid, err := Generate(defaultAlphabet, length)
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}

		if len(nanoid) != length {
			t.Errorf("NanoID length = %d, want %d", len(nanoid), length)
		}
	}
}

func TestGenerateCustomAlphabet(t *testing.T) {
	alphabet := "abc123"
	nanoid, err := Generate(alphabet, 100)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	for _, c := range nanoid {
		if !strings.ContainsRune(alphabet, c) {
			t.Errorf("NanoID contains character '%c' not in alphabet '%s'", c, alphabet)
		}
	}
}

func TestRunNanoID(t *testing.T) {
	var buf bytes.Buffer

	opts := Options{Count: 3}
	err := RunNanoID(&buf, opts)
	if err != nil {
		t.Fatalf("RunNanoID() error = %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Errorf("RunNanoID() generated %d NanoIDs, want 3", len(lines))
	}

	for i, line := range lines {
		if len(line) != defaultLength {
			t.Errorf("NanoID[%d] length = %d, want %d", i, len(line), defaultLength)
		}
	}
}

func TestRunNanoIDCustomLength(t *testing.T) {
	var buf bytes.Buffer

	opts := Options{Count: 1, Length: 10}
	err := RunNanoID(&buf, opts)
	if err != nil {
		t.Fatalf("RunNanoID() error = %v", err)
	}

	nanoid := strings.TrimSpace(buf.String())
	if len(nanoid) != 10 {
		t.Errorf("NanoID length = %d, want 10", len(nanoid))
	}
}

func TestRunNanoIDJSON(t *testing.T) {
	var buf bytes.Buffer

	opts := Options{Count: 2, JSON: true}
	err := RunNanoID(&buf, opts)
	if err != nil {
		t.Fatalf("RunNanoID() error = %v", err)
	}

	var result Result
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("JSON unmarshal error = %v", err)
	}

	if result.Count != 2 {
		t.Errorf("Count = %d, want 2", result.Count)
	}

	if len(result.NanoIDs) != 2 {
		t.Errorf("NanoIDs length = %d, want 2", len(result.NanoIDs))
	}
}

func TestNewString(t *testing.T) {
	str := NewString()
	if len(str) != defaultLength {
		t.Errorf("NewString() length = %d, want %d", len(str), defaultLength)
	}
}

func TestMustNew(t *testing.T) {
	// Should not panic
	nanoid := MustNew()
	if len(nanoid) != defaultLength {
		t.Errorf("MustNew() length = %d, want %d", len(nanoid), defaultLength)
	}
}
