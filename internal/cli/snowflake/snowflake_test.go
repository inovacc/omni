package snowflake

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	id, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if id <= 0 {
		t.Errorf("Snowflake ID = %d, want positive", id)
	}
}

func TestSnowflakeUniqueness(t *testing.T) {
	seen := make(map[int64]bool)
	count := 10000

	for range count {
		id, err := New()
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}

		if seen[id] {
			t.Errorf("Duplicate Snowflake ID generated: %d", id)
		}

		seen[id] = true
	}
}

func TestSnowflakeSortable(t *testing.T) {
	ids := make([]int64, 0, 10)

	for range 10 {
		id, err := New()
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}

		ids = append(ids, id)

		time.Sleep(time.Millisecond)
	}

	// Verify IDs are in ascending order
	for i := 1; i < len(ids); i++ {
		if ids[i] <= ids[i-1] {
			t.Errorf("Snowflake IDs not sorted: %d <= %d", ids[i], ids[i-1])
		}
	}
}

func TestGeneratorWorkerID(t *testing.T) {
	gen1 := NewGenerator(1)
	gen2 := NewGenerator(2)

	id1, err := gen1.Generate()
	if err != nil {
		t.Fatalf("gen1.Generate() error = %v", err)
	}

	id2, err := gen2.Generate()
	if err != nil {
		t.Fatalf("gen2.Generate() error = %v", err)
	}

	// Extract worker IDs
	_, workerID1, _ := Parse(id1)
	_, workerID2, _ := Parse(id2)

	if workerID1 != 1 {
		t.Errorf("Worker ID = %d, want 1", workerID1)
	}

	if workerID2 != 2 {
		t.Errorf("Worker ID = %d, want 2", workerID2)
	}
}

func TestParse(t *testing.T) {
	before := time.Now()

	id, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	after := time.Now()

	ts, workerID, sequence := Parse(id)

	// Timestamp should be within range
	if ts.Before(before.Add(-time.Second)) || ts.After(after.Add(time.Second)) {
		t.Errorf("Timestamp %v not within expected range [%v, %v]", ts, before, after)
	}

	// Worker ID should be 0 (default)
	if workerID != 0 {
		t.Errorf("Worker ID = %d, want 0", workerID)
	}

	// Sequence should be valid
	if sequence < 0 || sequence > 4095 {
		t.Errorf("Sequence = %d, want [0, %d]", sequence, 4095)
	}
}

func TestRunSnowflake(t *testing.T) {
	var buf bytes.Buffer

	opts := Options{Count: 3}

	err := RunSnowflake(&buf, opts)
	if err != nil {
		t.Fatalf("RunSnowflake() error = %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Errorf("RunSnowflake() generated %d IDs, want 3", len(lines))
	}
}

func TestRunSnowflakeWorkerID(t *testing.T) {
	var buf bytes.Buffer

	opts := Options{Count: 1, WorkerID: 42}

	err := RunSnowflake(&buf, opts)
	if err != nil {
		t.Fatalf("RunSnowflake() error = %v", err)
	}
}

func TestRunSnowflakeInvalidWorkerID(t *testing.T) {
	var buf bytes.Buffer

	opts := Options{Count: 1, WorkerID: 2000}

	err := RunSnowflake(&buf, opts)
	if err == nil {
		t.Error("RunSnowflake() should error with invalid worker ID")
	}
}

func TestRunSnowflakeJSON(t *testing.T) {
	var buf bytes.Buffer

	opts := Options{Count: 2, JSON: true}

	err := RunSnowflake(&buf, opts)
	if err != nil {
		t.Fatalf("RunSnowflake() error = %v", err)
	}

	var result Result
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("JSON unmarshal error = %v", err)
	}

	if result.Count != 2 {
		t.Errorf("Count = %d, want 2", result.Count)
	}

	if len(result.Snowflakes) != 2 {
		t.Errorf("Snowflakes length = %d, want 2", len(result.Snowflakes))
	}
}

func TestNewString(t *testing.T) {
	str := NewString()
	if str == "" {
		t.Error("NewString() returned empty string")
	}
}
