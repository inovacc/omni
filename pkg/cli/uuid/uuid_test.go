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
