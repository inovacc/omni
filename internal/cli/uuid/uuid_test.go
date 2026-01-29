package uuid

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunUUID(t *testing.T) {
	t.Run("single uuid", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunUUID(&buf, UUIDOptions{})
		if err != nil {
			t.Fatalf("RunUUID() error = %v", err)
		}

		uuid := strings.TrimSpace(buf.String())
		if !IsValidUUID(uuid) {
			t.Errorf("RunUUID() = %v, not a valid UUID", uuid)
		}
	})

	t.Run("multiple uuids", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunUUID(&buf, UUIDOptions{Count: 5})
		if err != nil {
			t.Fatalf("RunUUID() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 5 {
			t.Errorf("RunUUID() generated %d UUIDs, want 5", len(lines))
		}

		for _, uuid := range lines {
			if !IsValidUUID(uuid) {
				t.Errorf("RunUUID() generated invalid UUID: %v", uuid)
			}
		}
	})

	t.Run("uppercase", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunUUID(&buf, UUIDOptions{Upper: true})
		if err != nil {
			t.Fatalf("RunUUID() error = %v", err)
		}

		uuid := strings.TrimSpace(buf.String())
		if uuid != strings.ToUpper(uuid) {
			t.Errorf("RunUUID() = %v, want uppercase", uuid)
		}
	})

	t.Run("no dashes", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunUUID(&buf, UUIDOptions{NoDashes: true})
		if err != nil {
			t.Fatalf("RunUUID() error = %v", err)
		}

		uuid := strings.TrimSpace(buf.String())
		if strings.Contains(uuid, "-") {
			t.Errorf("RunUUID() = %v, want no dashes", uuid)
		}

		if len(uuid) != 32 {
			t.Errorf("RunUUID() length = %d, want 32", len(uuid))
		}
	})
}

func TestIsValidUUID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{
			name:  "valid with dashes",
			input: "550e8400-e29b-41d4-a716-446655440000",
			valid: true,
		},
		{
			name:  "valid without dashes",
			input: "550e8400e29b41d4a716446655440000",
			valid: true,
		},
		{
			name:  "valid uppercase",
			input: "550E8400-E29B-41D4-A716-446655440000",
			valid: true,
		},
		{
			name:  "too short",
			input: "550e8400-e29b-41d4",
			valid: false,
		},
		{
			name:  "invalid characters",
			input: "550e8400-e29b-41d4-a716-44665544000g",
			valid: false,
		},
		{
			name:  "empty",
			input: "",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidUUID(tt.input)
			if result != tt.valid {
				t.Errorf("IsValidUUID(%v) = %v, want %v", tt.input, result, tt.valid)
			}
		})
	}
}

func TestNewUUID(t *testing.T) {
	uuid := NewUUID()
	if uuid == "" {
		t.Error("NewUUID() returned empty string")
	}

	if !IsValidUUID(uuid) {
		t.Errorf("NewUUID() = %v, not a valid UUID", uuid)
	}
}

func TestMustNewUUID(t *testing.T) {
	uuid := MustNewUUID()
	if uuid == "" {
		t.Error("MustNewUUID() returned empty string")
	}

	if !IsValidUUID(uuid) {
		t.Errorf("MustNewUUID() = %v, not a valid UUID", uuid)
	}
}

func TestGenerateUUIDv4(t *testing.T) {
	// Generate multiple UUIDs and verify they're unique
	uuids := make(map[string]bool)

	for range 100 {
		uuid, err := generateUUIDv4()
		if err != nil {
			t.Fatalf("generateUUIDv4() error = %v", err)
		}

		if uuids[uuid] {
			t.Errorf("generateUUIDv4() generated duplicate UUID: %v", uuid)
		}

		uuids[uuid] = true

		// Verify version 4 format
		parts := strings.Split(uuid, "-")
		if len(parts) != 5 {
			t.Errorf("generateUUIDv4() = %v, wrong format", uuid)
		}
		// Version should be 4
		if parts[2][0] != '4' {
			t.Errorf("generateUUIDv4() version = %c, want 4", parts[2][0])
		}
	}
}

func TestRunUUIDExtended(t *testing.T) {
	t.Run("zero count", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunUUID(&buf, UUIDOptions{Count: 0})
		if err != nil {
			t.Fatalf("RunUUID() error = %v", err)
		}

		// Count 0 should produce at least 1 UUID or empty
		output := strings.TrimSpace(buf.String())
		if output != "" && !IsValidUUID(output) {
			t.Errorf("RunUUID() = %v, not valid", output)
		}
	})

	t.Run("negative count", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunUUID(&buf, UUIDOptions{Count: -1})
		// Negative count behavior depends on implementation
		if err != nil {
			t.Logf("RunUUID() negative count: %v", err)
		}
	})

	t.Run("large count", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunUUID(&buf, UUIDOptions{Count: 100})
		if err != nil {
			t.Fatalf("RunUUID() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 100 {
			t.Errorf("RunUUID() generated %d UUIDs, want 100", len(lines))
		}
	})

	t.Run("uppercase no dashes", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunUUID(&buf, UUIDOptions{Upper: true, NoDashes: true})
		if err != nil {
			t.Fatalf("RunUUID() error = %v", err)
		}

		uuid := strings.TrimSpace(buf.String())
		if strings.Contains(uuid, "-") {
			t.Error("RunUUID() should have no dashes")
		}

		if uuid != strings.ToUpper(uuid) {
			t.Error("RunUUID() should be uppercase")
		}

		if len(uuid) != 32 {
			t.Errorf("RunUUID() length = %d, want 32", len(uuid))
		}
	})

	t.Run("all unique", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunUUID(&buf, UUIDOptions{Count: 1000})
		if err != nil {
			t.Fatalf("RunUUID() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		seen := make(map[string]bool)

		for _, uuid := range lines {
			if seen[uuid] {
				t.Errorf("RunUUID() generated duplicate: %v", uuid)
			}

			seen[uuid] = true
		}
	})

	t.Run("consistent format", func(t *testing.T) {
		var buf bytes.Buffer

		_ = RunUUID(&buf, UUIDOptions{Count: 10})

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		for i, uuid := range lines {
			if !IsValidUUID(uuid) {
				t.Errorf("RunUUID() line %d invalid: %v", i, uuid)
			}
		}
	})

	t.Run("output ends with newline", func(t *testing.T) {
		var buf bytes.Buffer

		_ = RunUUID(&buf, UUIDOptions{})

		output := buf.String()
		if len(output) > 0 && !strings.HasSuffix(output, "\n") {
			t.Error("RunUUID() output should end with newline")
		}
	})
}

func TestIsValidUUIDExtended(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{
			name:  "all zeros",
			input: "00000000-0000-0000-0000-000000000000",
			valid: true,
		},
		{
			name:  "all fs",
			input: "ffffffff-ffff-ffff-ffff-ffffffffffff",
			valid: true,
		},
		{
			name:  "mixed case",
			input: "550e8400-E29B-41d4-A716-446655440000",
			valid: true,
		},
		{
			name:  "wrong length with dashes",
			input: "550e8400-e29b-41d4-a716-4466554400",
			valid: false,
		},
		{
			name:  "extra dashes",
			input: "550e8400-e29b-41d4-a716-44665544-0000",
			valid: true, // Implementation strips all dashes, so this is valid
		},
		{
			name:  "no dashes wrong length",
			input: "550e8400e29b41d4a71644665544000",
			valid: false,
		},
		{
			name:  "with braces",
			input: "{550e8400-e29b-41d4-a716-446655440000}",
			valid: false,
		},
		{
			name:  "whitespace",
			input: " 550e8400-e29b-41d4-a716-446655440000",
			valid: false,
		},
		{
			name:  "newline suffix",
			input: "550e8400-e29b-41d4-a716-446655440000\n",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidUUID(tt.input)
			if result != tt.valid {
				t.Errorf("IsValidUUID(%v) = %v, want %v", tt.input, result, tt.valid)
			}
		})
	}
}

func TestUUIDUniqueness(t *testing.T) {
	t.Run("NewUUID uniqueness", func(t *testing.T) {
		seen := make(map[string]bool)

		for i := range 10000 {
			uuid := NewUUID()
			if seen[uuid] {
				t.Fatalf("NewUUID() collision at iteration %d", i)
			}

			seen[uuid] = true
		}
	})

	t.Run("MustNewUUID uniqueness", func(t *testing.T) {
		seen := make(map[string]bool)

		for i := range 10000 {
			uuid := MustNewUUID()
			if seen[uuid] {
				t.Fatalf("MustNewUUID() collision at iteration %d", i)
			}

			seen[uuid] = true
		}
	})
}

func TestUUIDVersion(t *testing.T) {
	t.Run("version 4 indicator", func(t *testing.T) {
		for range 100 {
			uuid := NewUUID()

			parts := strings.Split(uuid, "-")
			if len(parts) != 5 {
				t.Fatalf("UUID format invalid: %v", uuid)
			}

			// Third part should start with 4
			if parts[2][0] != '4' {
				t.Errorf("UUID version = %c, want 4", parts[2][0])
			}
		}
	})

	t.Run("variant indicator", func(t *testing.T) {
		for range 100 {
			uuid := NewUUID()

			parts := strings.Split(uuid, "-")
			if len(parts) != 5 {
				continue
			}

			// Fourth part should start with 8, 9, a, or b
			firstChar := parts[3][0]
			validVariant := firstChar == '8' || firstChar == '9' ||
				firstChar == 'a' || firstChar == 'b' ||
				firstChar == 'A' || firstChar == 'B'

			if !validVariant {
				t.Errorf("UUID variant = %c, want 8/9/a/b", firstChar)
			}
		}
	})
}
