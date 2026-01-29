package ksuid

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	ksuid, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	str := ksuid.String()
	if len(str) != encodedSize {
		t.Errorf("KSUID length = %d, want %d", len(str), encodedSize)
	}
}

func TestKSUIDUniqueness(t *testing.T) {
	seen := make(map[string]bool)
	count := 1000

	for i := 0; i < count; i++ {
		ksuid, err := New()
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}

		str := ksuid.String()
		if seen[str] {
			t.Errorf("Duplicate KSUID generated: %s", str)
		}
		seen[str] = true
	}
}

func TestKSUIDTimestamp(t *testing.T) {
	before := time.Now()
	ksuid, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	after := time.Now()

	ts := ksuid.Timestamp()
	if ts.Before(before.Add(-time.Second)) || ts.After(after.Add(time.Second)) {
		t.Errorf("KSUID timestamp %v not within expected range [%v, %v]", ts, before, after)
	}
}

func TestRunKSUID(t *testing.T) {
	var buf bytes.Buffer

	opts := Options{Count: 3}
	err := RunKSUID(&buf, opts)
	if err != nil {
		t.Fatalf("RunKSUID() error = %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Errorf("RunKSUID() generated %d KSUIDs, want 3", len(lines))
	}

	for i, line := range lines {
		if len(line) != encodedSize {
			t.Errorf("KSUID[%d] length = %d, want %d", i, len(line), encodedSize)
		}
	}
}

func TestRunKSUIDJSON(t *testing.T) {
	var buf bytes.Buffer

	opts := Options{Count: 2, JSON: true}
	err := RunKSUID(&buf, opts)
	if err != nil {
		t.Fatalf("RunKSUID() error = %v", err)
	}

	var result Result
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("JSON unmarshal error = %v", err)
	}

	if result.Count != 2 {
		t.Errorf("Count = %d, want 2", result.Count)
	}

	if len(result.KSUIDs) != 2 {
		t.Errorf("KSUIDs length = %d, want 2", len(result.KSUIDs))
	}
}

func TestNewString(t *testing.T) {
	str := NewString()
	if len(str) != encodedSize {
		t.Errorf("NewString() length = %d, want %d", len(str), encodedSize)
	}
}
