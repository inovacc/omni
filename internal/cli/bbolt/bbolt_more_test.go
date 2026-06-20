package bbolt

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestRunInfo(t *testing.T) {
	dbPath, cleanup := createTestDB(t)
	defer cleanup()

	t.Run("text", func(t *testing.T) {
		var buf bytes.Buffer

		if err := RunInfo(&buf, dbPath, Options{}); err != nil {
			t.Fatalf("RunInfo() error = %v", err)
		}

		if !strings.Contains(buf.String(), "Page Size:") {
			t.Errorf("RunInfo() text output should contain page size, got %q", buf.String())
		}
	})

	t.Run("json", func(t *testing.T) {
		var buf bytes.Buffer

		if err := RunInfo(&buf, dbPath, Options{JSON: true}); err != nil {
			t.Fatalf("RunInfo() json error = %v", err)
		}

		var res InfoResult
		if err := json.Unmarshal(buf.Bytes(), &res); err != nil {
			t.Fatalf("RunInfo() json invalid: %v", err)
		}

		if res.PageSize <= 0 {
			t.Errorf("RunInfo() json page size = %d, want > 0", res.PageSize)
		}

		if res.Path == "" {
			t.Error("RunInfo() json path should not be empty")
		}
	})

	t.Run("missing db", func(t *testing.T) {
		var buf bytes.Buffer

		if err := RunInfo(&buf, dbPath+".nope", Options{}); err == nil {
			t.Error("RunInfo() should error on missing db")
		}
	})
}

func TestRunPages(t *testing.T) {
	dbPath, cleanup := createTestDB(t)
	defer cleanup()

	t.Run("text", func(t *testing.T) {
		var buf bytes.Buffer

		if err := RunPages(&buf, dbPath, Options{}); err != nil {
			t.Fatalf("RunPages() error = %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "TYPE") || !strings.Contains(out, "OVERFLOW") {
			t.Errorf("RunPages() text output should contain header, got %q", out)
		}
	})

	t.Run("json", func(t *testing.T) {
		var buf bytes.Buffer

		if err := RunPages(&buf, dbPath, Options{JSON: true}); err != nil {
			t.Fatalf("RunPages() json error = %v", err)
		}

		var pages []PageInfo
		if err := json.Unmarshal(buf.Bytes(), &pages); err != nil {
			t.Fatalf("RunPages() json invalid: %v", err)
		}

		if len(pages) == 0 {
			t.Error("RunPages() json should report at least one page")
		}
	})

	t.Run("missing db", func(t *testing.T) {
		var buf bytes.Buffer

		if err := RunPages(&buf, dbPath+".nope", Options{}); err == nil {
			t.Error("RunPages() should error on missing db")
		}
	})
}

func TestRunPageDump(t *testing.T) {
	dbPath, cleanup := createTestDB(t)
	defer cleanup()

	t.Run("text page 0", func(t *testing.T) {
		var buf bytes.Buffer

		if err := RunPageDump(&buf, dbPath, 0, Options{}); err != nil {
			t.Fatalf("RunPageDump() error = %v", err)
		}

		if !strings.Contains(buf.String(), "Page 0") {
			t.Errorf("RunPageDump() should print page header, got %q", buf.String())
		}
	})

	t.Run("json page 0", func(t *testing.T) {
		var buf bytes.Buffer

		if err := RunPageDump(&buf, dbPath, 0, Options{JSON: true}); err != nil {
			t.Fatalf("RunPageDump() json error = %v", err)
		}

		var res map[string]any
		if err := json.Unmarshal(buf.Bytes(), &res); err != nil {
			t.Fatalf("RunPageDump() json invalid: %v", err)
		}

		if _, ok := res["hex"]; !ok {
			t.Error("RunPageDump() json should include hex field")
		}
	})

	t.Run("negative page id", func(t *testing.T) {
		var buf bytes.Buffer

		if err := RunPageDump(&buf, dbPath, -1, Options{}); err == nil {
			t.Error("RunPageDump() should reject negative page id")
		}
	})

	t.Run("out of range page id", func(t *testing.T) {
		var buf bytes.Buffer

		if err := RunPageDump(&buf, dbPath, 1<<20, Options{}); err == nil {
			t.Error("RunPageDump() should reject out-of-range page id")
		}
	})

	t.Run("missing db", func(t *testing.T) {
		var buf bytes.Buffer

		if err := RunPageDump(&buf, dbPath+".nope", 0, Options{}); err == nil {
			t.Error("RunPageDump() should error on missing db")
		}
	})
}

func TestIsValidPageSize(t *testing.T) {
	tests := []struct {
		name string
		size int
		want bool
	}{
		{"too small", 256, false},
		{"min valid", 512, true},
		{"4096 valid", 4096, true},
		{"max valid", 65536, true},
		{"too large", 131072, false},
		{"not power of two", 4095, false},
		{"zero", 0, false},
		{"negative", -1, false},
		{"odd in range", 1000, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidPageSize(tt.size); got != tt.want {
				t.Errorf("isValidPageSize(%d) = %v, want %v", tt.size, got, tt.want)
			}
		})
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name  string
		value []byte
		asHex bool
		want  string
	}{
		{"plain string", []byte("hello"), false, "hello"},
		{"hex encoded", []byte{0x00, 0xff, 0x10}, true, "00ff10"},
		{"empty plain", []byte{}, false, ""},
		{"empty hex", []byte{}, true, ""},
		{"binary as string", []byte{0x41, 0x42}, false, "AB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatValue(tt.value, tt.asHex); got != tt.want {
				t.Errorf("formatValue(%v, %v) = %q, want %q", tt.value, tt.asHex, got, tt.want)
			}
		})
	}
}
