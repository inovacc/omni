package idgen

import (
	"strings"
	"testing"
)

func TestGenerateUUID(t *testing.T) {
	t.Run("default v4", func(t *testing.T) {
		uuid, err := GenerateUUID()
		if err != nil {
			t.Fatalf("GenerateUUID() error = %v", err)
		}

		if !IsValidUUID(uuid) {
			t.Errorf("GenerateUUID() = %v, not a valid UUID", uuid)
		}

		parts := strings.Split(uuid, "-")
		if len(parts) != 5 {
			t.Fatalf("UUID format invalid: %v", uuid)
		}

		if parts[2][0] != '4' {
			t.Errorf("UUID version = %c, want 4", parts[2][0])
		}
	})

	t.Run("v7", func(t *testing.T) {
		uuid, err := GenerateUUID(WithUUIDVersion(V7))
		if err != nil {
			t.Fatalf("GenerateUUID(V7) error = %v", err)
		}

		if !IsValidUUID(uuid) {
			t.Errorf("GenerateUUID(V7) = %v, not valid", uuid)
		}

		parts := strings.Split(uuid, "-")
		if parts[2][0] != '7' {
			t.Errorf("UUID version = %c, want 7", parts[2][0])
		}
	})

	t.Run("uppercase", func(t *testing.T) {
		uuid, err := GenerateUUID(WithUppercase())
		if err != nil {
			t.Fatal(err)
		}

		if uuid != strings.ToUpper(uuid) {
			t.Errorf("expected uppercase, got %v", uuid)
		}
	})

	t.Run("no dashes", func(t *testing.T) {
		uuid, err := GenerateUUID(WithNoDashes())
		if err != nil {
			t.Fatal(err)
		}

		if strings.Contains(uuid, "-") {
			t.Errorf("expected no dashes, got %v", uuid)
		}

		if len(uuid) != 32 {
			t.Errorf("length = %d, want 32", len(uuid))
		}
	})

	t.Run("uppercase no dashes", func(t *testing.T) {
		uuid, err := GenerateUUID(WithUppercase(), WithNoDashes())
		if err != nil {
			t.Fatal(err)
		}

		if strings.Contains(uuid, "-") {
			t.Error("should have no dashes")
		}

		if uuid != strings.ToUpper(uuid) {
			t.Error("should be uppercase")
		}

		if len(uuid) != 32 {
			t.Errorf("length = %d, want 32", len(uuid))
		}
	})

	t.Run("unsupported version", func(t *testing.T) {
		_, err := GenerateUUID(WithUUIDVersion(3))
		if err == nil {
			t.Error("expected error for unsupported version")
		}
	})
}

func TestGenerateUUIDs(t *testing.T) {
	uuids, err := GenerateUUIDs(5)
	if err != nil {
		t.Fatalf("GenerateUUIDs() error = %v", err)
	}

	if len(uuids) != 5 {
		t.Errorf("got %d UUIDs, want 5", len(uuids))
	}

	for _, uuid := range uuids {
		if !IsValidUUID(uuid) {
			t.Errorf("invalid UUID: %v", uuid)
		}
	}
}

func TestIsValidUUID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{"valid with dashes", "550e8400-e29b-41d4-a716-446655440000", true},
		{"valid without dashes", "550e8400e29b41d4a716446655440000", true},
		{"valid uppercase", "550E8400-E29B-41D4-A716-446655440000", true},
		{"too short", "550e8400-e29b-41d4", false},
		{"invalid characters", "550e8400-e29b-41d4-a716-44665544000g", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidUUID(tt.input); got != tt.valid {
				t.Errorf("IsValidUUID(%v) = %v, want %v", tt.input, got, tt.valid)
			}
		})
	}
}

func TestUUIDUniqueness(t *testing.T) {
	seen := make(map[string]bool)

	for i := range 10000 {
		uuid, err := GenerateUUID()
		if err != nil {
			t.Fatal(err)
		}

		if seen[uuid] {
			t.Fatalf("collision at iteration %d", i)
		}

		seen[uuid] = true
	}
}

func TestGenerateULID(t *testing.T) {
	u, err := GenerateULID()
	if err != nil {
		t.Fatalf("GenerateULID() error = %v", err)
	}

	s := u.String()
	if len(s) != 26 {
		t.Errorf("ULID length = %d, want 26", len(s))
	}
}

func TestULIDString(t *testing.T) {
	s := ULIDString()
	if s == "" {
		t.Error("ULIDString() returned empty")
	}

	if len(s) != 26 {
		t.Errorf("ULID length = %d, want 26", len(s))
	}
}

func TestULIDTimestamp(t *testing.T) {
	u, err := GenerateULID()
	if err != nil {
		t.Fatal(err)
	}

	ts := u.Timestamp()
	if ts.IsZero() {
		t.Error("ULID timestamp is zero")
	}
}

func TestULIDUniqueness(t *testing.T) {
	seen := make(map[string]bool)

	for range 1000 {
		s := ULIDString()
		if seen[s] {
			t.Fatal("ULID collision")
		}

		seen[s] = true
	}
}

func TestGenerateKSUID(t *testing.T) {
	k, err := GenerateKSUID()
	if err != nil {
		t.Fatalf("GenerateKSUID() error = %v", err)
	}

	s := k.String()
	if len(s) != 27 {
		t.Errorf("KSUID length = %d, want 27", len(s))
	}
}

func TestKSUIDString(t *testing.T) {
	s := KSUIDString()
	if s == "" {
		t.Error("KSUIDString() returned empty")
	}

	if len(s) != 27 {
		t.Errorf("KSUID length = %d, want 27", len(s))
	}
}

func TestKSUIDTimestamp(t *testing.T) {
	k, err := GenerateKSUID()
	if err != nil {
		t.Fatal(err)
	}

	ts := k.Timestamp()
	if ts.IsZero() {
		t.Error("KSUID timestamp is zero")
	}
}

func TestGenerateNanoid(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		n, err := GenerateNanoid()
		if err != nil {
			t.Fatalf("GenerateNanoid() error = %v", err)
		}

		if len(n) != 21 {
			t.Errorf("NanoID length = %d, want 21", len(n))
		}
	})

	t.Run("custom length", func(t *testing.T) {
		n, err := GenerateNanoid(WithNanoidLength(10))
		if err != nil {
			t.Fatal(err)
		}

		if len(n) != 10 {
			t.Errorf("NanoID length = %d, want 10", len(n))
		}
	})

	t.Run("custom alphabet", func(t *testing.T) {
		n, err := GenerateNanoid(WithNanoidAlphabet("abc"), WithNanoidLength(20))
		if err != nil {
			t.Fatal(err)
		}

		for _, c := range n {
			if c != 'a' && c != 'b' && c != 'c' {
				t.Errorf("unexpected character %c in NanoID", c)
			}
		}
	})
}

func TestNanoidString(t *testing.T) {
	s := NanoidString()
	if s == "" {
		t.Error("NanoidString() returned empty")
	}

	if len(s) != 21 {
		t.Errorf("NanoID length = %d, want 21", len(s))
	}
}

func TestSnowflakeGenerator(t *testing.T) {
	gen := NewSnowflakeGenerator(1)

	id, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if id <= 0 {
		t.Errorf("Snowflake ID = %d, want positive", id)
	}
}

func TestGenerateSnowflake(t *testing.T) {
	id, err := GenerateSnowflake()
	if err != nil {
		t.Fatalf("GenerateSnowflake() error = %v", err)
	}

	if id <= 0 {
		t.Errorf("Snowflake ID = %d, want positive", id)
	}
}

func TestSnowflakeString(t *testing.T) {
	s := SnowflakeString()
	if s == "" {
		t.Error("SnowflakeString() returned empty")
	}
}

func TestSnowflakeUniqueness(t *testing.T) {
	gen := NewSnowflakeGenerator(0)
	seen := make(map[int64]bool)

	for i := range 10000 {
		id, err := gen.Generate()
		if err != nil {
			t.Fatal(err)
		}

		if seen[id] {
			t.Fatalf("Snowflake collision at iteration %d", i)
		}

		seen[id] = true
	}
}

func TestParseSnowflake(t *testing.T) {
	gen := NewSnowflakeGenerator(42)

	id, err := gen.Generate()
	if err != nil {
		t.Fatal(err)
	}

	ts, workerID, seq := ParseSnowflake(id)
	if ts.IsZero() {
		t.Error("parsed timestamp is zero")
	}

	if workerID != 42 {
		t.Errorf("parsed workerID = %d, want 42", workerID)
	}

	if seq < 0 {
		t.Errorf("parsed sequence = %d, want >= 0", seq)
	}
}
