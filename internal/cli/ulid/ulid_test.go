package ulid

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	ulid, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	str := ulid.String()
	if len(str) != 26 {
		t.Errorf("ULID length = %d, want 26", len(str))
	}
}

func TestULIDUniqueness(t *testing.T) {
	seen := make(map[string]bool)
	count := 1000

	for range count {
		ulid, err := New()
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}

		str := ulid.String()
		if seen[str] {
			t.Errorf("Duplicate ULID generated: %s", str)
		}

		seen[str] = true
	}
}

func TestULIDTimestamp(t *testing.T) {
	before := time.Now()

	ulid, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	after := time.Now()

	ts := ulid.Timestamp()
	if ts.Before(before.Add(-time.Second)) || ts.After(after.Add(time.Second)) {
		t.Errorf("ULID timestamp %v not within expected range [%v, %v]", ts, before, after)
	}
}

func TestULIDSortable(t *testing.T) {
	ulids := make([]string, 0, 10)

	for range 10 {
		ulid, err := New()
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}

		ulids = append(ulids, ulid.String())

		time.Sleep(time.Millisecond)
	}

	// Verify ULIDs are in ascending order
	for i := 1; i < len(ulids); i++ {
		if ulids[i] < ulids[i-1] {
			t.Errorf("ULIDs not sorted: %s < %s", ulids[i], ulids[i-1])
		}
	}
}

func TestRunULID(t *testing.T) {
	var buf bytes.Buffer

	opts := Options{Count: 3}

	err := RunULID(&buf, opts)
	if err != nil {
		t.Fatalf("RunULID() error = %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Errorf("RunULID() generated %d ULIDs, want 3", len(lines))
	}

	for i, line := range lines {
		if len(line) != 26 {
			t.Errorf("ULID[%d] length = %d, want 26", i, len(line))
		}
	}
}

func TestRunULIDLower(t *testing.T) {
	var buf bytes.Buffer

	opts := Options{Count: 1, Lower: true}

	err := RunULID(&buf, opts)
	if err != nil {
		t.Fatalf("RunULID() error = %v", err)
	}

	ulid := strings.TrimSpace(buf.String())
	if ulid != strings.ToLower(ulid) {
		t.Errorf("ULID not lowercase: %s", ulid)
	}
}

func TestRunULIDJSON(t *testing.T) {
	var buf bytes.Buffer

	opts := Options{Count: 2, JSON: true}

	err := RunULID(&buf, opts)
	if err != nil {
		t.Fatalf("RunULID() error = %v", err)
	}

	var result Result
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("JSON unmarshal error = %v", err)
	}

	if result.Count != 2 {
		t.Errorf("Count = %d, want 2", result.Count)
	}

	if len(result.ULIDs) != 2 {
		t.Errorf("ULIDs length = %d, want 2", len(result.ULIDs))
	}
}

func TestNewString(t *testing.T) {
	str := NewString()
	if len(str) != 26 {
		t.Errorf("NewString() length = %d, want 26", len(str))
	}
}
