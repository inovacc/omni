package cron

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		wantErr bool
		checkFn func(*Schedule) bool
	}{
		{
			name: "every minute",
			expr: "* * * * *",
			checkFn: func(s *Schedule) bool {
				return len(s.Minutes) == 60 && len(s.Hours) == 24
			},
		},
		{
			name: "specific time",
			expr: "30 9 * * *",
			checkFn: func(s *Schedule) bool {
				return len(s.Minutes) == 1 && s.Minutes[0] == 30 &&
					len(s.Hours) == 1 && s.Hours[0] == 9
			},
		},
		{
			name: "every 15 minutes",
			expr: "*/15 * * * *",
			checkFn: func(s *Schedule) bool {
				return len(s.Minutes) == 4 &&
					s.Minutes[0] == 0 && s.Minutes[1] == 15 &&
					s.Minutes[2] == 30 && s.Minutes[3] == 45
			},
		},
		{
			name: "weekdays only",
			expr: "0 9 * * 1-5",
			checkFn: func(s *Schedule) bool {
				return len(s.Weekdays) == 5 &&
					s.Weekdays[0] == 1 && s.Weekdays[4] == 5
			},
		},
		{
			name: "specific months",
			expr: "0 0 1 1,6,12 *",
			checkFn: func(s *Schedule) bool {
				return len(s.Months) == 3 &&
					s.Months[0] == 1 && s.Months[1] == 6 && s.Months[2] == 12
			},
		},
		{
			name: "month names",
			expr: "0 0 1 jan,jul *",
			checkFn: func(s *Schedule) bool {
				return len(s.Months) == 2 &&
					s.Months[0] == 1 && s.Months[1] == 7
			},
		},
		{
			name: "weekday names",
			expr: "0 0 * * mon,wed,fri",
			checkFn: func(s *Schedule) bool {
				return len(s.Weekdays) == 3 &&
					s.Weekdays[0] == 1 && s.Weekdays[1] == 3 && s.Weekdays[2] == 5
			},
		},
		{
			name: "sunday as 0",
			expr: "0 0 * * 0",
			checkFn: func(s *Schedule) bool {
				return len(s.Weekdays) == 1 && s.Weekdays[0] == 0
			},
		},
		{
			name: "sunday as 7",
			expr: "0 0 * * 7",
			checkFn: func(s *Schedule) bool {
				return len(s.Weekdays) == 1 && s.Weekdays[0] == 0
			},
		},
		{
			name:    "invalid - too few fields",
			expr:    "* * * *",
			wantErr: true,
		},
		{
			name:    "invalid - too many fields",
			expr:    "* * * * * *",
			wantErr: true,
		},
		{
			name:    "invalid - bad minute",
			expr:    "60 * * * *",
			wantErr: true,
		},
		{
			name:    "invalid - bad hour",
			expr:    "* 24 * * *",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schedule, err := Parse(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.expr, err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFn != nil && !tt.checkFn(schedule) {
				t.Errorf("Parse(%q) check failed", tt.expr)
			}
		})
	}
}

func TestExpandAlias(t *testing.T) {
	tests := []struct {
		alias string
		want  string
	}{
		{"@yearly", "0 0 1 1 *"},
		{"@annually", "0 0 1 1 *"},
		{"@monthly", "0 0 1 * *"},
		{"@weekly", "0 0 * * 0"},
		{"@daily", "0 0 * * *"},
		{"@midnight", "0 0 * * *"},
		{"@hourly", "0 * * * *"},
		{"* * * * *", "* * * * *"}, // Not an alias
	}

	for _, tt := range tests {
		t.Run(tt.alias, func(t *testing.T) {
			got := expandAlias(tt.alias)
			if got != tt.want {
				t.Errorf("expandAlias(%q) = %q, want %q", tt.alias, got, tt.want)
			}
		})
	}
}

func TestRun(t *testing.T) {
	var buf bytes.Buffer

	err := Run(&buf, []string{"0", "9", "*", "*", "1-5"}, Options{})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "Expression:") {
		t.Error("Output should contain 'Expression:'")
	}

	if !strings.Contains(output, "Schedule:") {
		t.Error("Output should contain 'Schedule:'")
	}
}

func TestRunJSON(t *testing.T) {
	var buf bytes.Buffer

	err := Run(&buf, []string{"*/15 * * * *"}, Options{JSON: true})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	var schedule Schedule
	if err := json.Unmarshal(buf.Bytes(), &schedule); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if len(schedule.Minutes) != 4 {
		t.Errorf("Expected 4 minutes, got %d", len(schedule.Minutes))
	}
}

func TestRunValidate(t *testing.T) {
	var buf bytes.Buffer

	err := Run(&buf, []string{"0 9 * * *"}, Options{Validate: true})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !strings.Contains(buf.String(), "Valid") {
		t.Error("Output should indicate valid expression")
	}
}

func TestRunValidateJSON(t *testing.T) {
	var buf bytes.Buffer

	err := Run(&buf, []string{"0 9 * * *"}, Options{Validate: true, JSON: true})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if valid, ok := result["valid"].(bool); !ok || !valid {
		t.Error("JSON should show valid: true")
	}
}

func TestRunNextRuns(t *testing.T) {
	var buf bytes.Buffer

	err := Run(&buf, []string{"0 * * * *"}, Options{Next: 3})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "Next runs:") {
		t.Error("Output should contain 'Next runs:'")
	}
}

func TestRunError(t *testing.T) {
	var buf bytes.Buffer

	err := Run(&buf, []string{}, Options{})
	if err == nil {
		t.Error("Run() expected error for empty args")
	}
}

func TestRunInvalidExpression(t *testing.T) {
	var buf bytes.Buffer

	err := Run(&buf, []string{"invalid"}, Options{})
	if err == nil {
		t.Error("Run() expected error for invalid expression")
	}
}

func TestParseField(t *testing.T) {
	tests := []struct {
		field string
		min   int
		max   int
		want  []int
	}{
		{"*", 0, 5, []int{0, 1, 2, 3, 4, 5}},
		{"*/2", 0, 5, []int{0, 2, 4}},
		{"1,3,5", 0, 5, []int{1, 3, 5}},
		{"1-3", 0, 5, []int{1, 2, 3}},
		{"1-4/2", 0, 5, []int{1, 3}},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			got, err := parseField(tt.field, tt.min, tt.max)
			if err != nil {
				t.Errorf("parseField(%q) error = %v", tt.field, err)
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("parseField(%q) = %v, want %v", tt.field, got, tt.want)
				return
			}

			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("parseField(%q) = %v, want %v", tt.field, got, tt.want)
					break
				}
			}
		})
	}
}

func TestHumanDescription(t *testing.T) {
	schedule, err := Parse("0 9 * * 1-5")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if schedule.Human == "" {
		t.Error("Human description should not be empty")
	}

	// Check that it mentions the time
	if !strings.Contains(schedule.Human, "09:00") {
		t.Errorf("Human description should mention time, got: %s", schedule.Human)
	}
}

func TestContains(t *testing.T) {
	vals := []int{1, 3, 5, 7}

	if !contains(vals, 3) {
		t.Error("contains(vals, 3) should be true")
	}

	if contains(vals, 2) {
		t.Error("contains(vals, 2) should be false")
	}
}
