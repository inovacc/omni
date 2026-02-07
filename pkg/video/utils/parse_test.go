package utils

import (
	"math"
	"testing"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input string
		want  float64
		ok    bool
	}{
		{"", 0, false},
		{"90", 90, true},
		{"5400", 5400, true},
		{"1:30", 90, true},
		{"01:30", 90, true},
		{"1:30:00", 5400, true},
		{"01:30:00", 5400, true},
		{"PT1H30M", 5400, true},
		{"PT90S", 90, true},
		{"PT1M30S", 90, true},
		{"P1DT0S", 86400, true},
		{"1 hour 30 minutes", 5400, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := ParseDuration(tt.input)
			if ok != tt.ok {
				t.Fatalf("ParseDuration(%q): ok = %v, want %v", tt.input, ok, tt.ok)
			}

			if ok && math.Abs(got-tt.want) > 0.01 {
				t.Errorf("ParseDuration(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseFilesize(t *testing.T) {
	tests := []struct {
		input string
		want  int64
		ok    bool
	}{
		{"", 0, false},
		{"100", 100, true},
		{"1KB", 1000, true},
		{"1KiB", 1024, true},
		{"1MB", 1000000, true},
		{"1MiB", 1048576, true},
		{"1.5GB", 1500000000, true},
		{"1.5GiB", 1610612736, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := ParseFilesize(tt.input)
			if ok != tt.ok {
				t.Fatalf("ParseFilesize(%q): ok = %v, want %v", tt.input, ok, tt.ok)
			}

			if ok && got != tt.want {
				t.Errorf("ParseFilesize(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseISO8601(t *testing.T) {
	tests := []struct {
		input string
		year  int
		month int
		day   int
		ok    bool
	}{
		{"20240115", 2024, 1, 15, true},
		{"2024-01-15", 2024, 1, 15, true},
		{"2024-01-15T10:30:00Z", 2024, 1, 15, true},
		{"", 0, 0, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := ParseISO8601(tt.input)
			if ok != tt.ok {
				t.Fatalf("ParseISO8601(%q): ok = %v, want %v", tt.input, ok, tt.ok)
			}

			if ok {
				if got.Year() != tt.year || int(got.Month()) != tt.month || got.Day() != tt.day {
					t.Errorf("ParseISO8601(%q) = %v, want %d-%d-%d", tt.input, got, tt.year, tt.month, tt.day)
				}
			}
		})
	}
}

func TestIntOrNone(t *testing.T) {
	tests := []struct {
		input string
		want  *int64
	}{
		{"", nil},
		{"abc", nil},
		{"42", ptr(int64(42))},
		{"1,234,567", ptr(int64(1234567))},
	}

	for _, tt := range tests {
		got := IntOrNone(tt.input)
		if tt.want == nil && got != nil {
			t.Errorf("IntOrNone(%q) = %v, want nil", tt.input, *got)
		}

		if tt.want != nil && (got == nil || *got != *tt.want) {
			t.Errorf("IntOrNone(%q) = %v, want %v", tt.input, got, *tt.want)
		}
	}
}

func ptr(n int64) *int64 { return &n }
