package tomlutil

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/inovacc/omni/internal/cli/output"
)

func TestRunValidate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		opts    ValidateOptions
		wantErr bool
	}{
		{
			name:    "valid simple toml",
			input:   "name = \"test\"\nvalue = 123",
			opts:    ValidateOptions{},
			wantErr: false,
		},
		{
			name:    "valid with section",
			input:   "[database]\nhost = \"localhost\"\nport = 5432",
			opts:    ValidateOptions{},
			wantErr: false,
		},
		{
			name:    "valid array",
			input:   "items = [1, 2, 3]",
			opts:    ValidateOptions{},
			wantErr: false,
		},
		{
			name:    "valid nested table",
			input:   "[servers.alpha]\nip = \"10.0.0.1\"",
			opts:    ValidateOptions{},
			wantErr: false,
		},
		{
			name:    "invalid toml - unclosed string",
			input:   "name = \"unclosed",
			opts:    ValidateOptions{},
			wantErr: true,
		},
		{
			name:    "invalid toml - bad syntax",
			input:   "name = = value",
			opts:    ValidateOptions{},
			wantErr: true,
		},
		{
			name:    "empty input",
			input:   "",
			opts:    ValidateOptions{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := RunValidate(&buf, []string{tt.input}, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunValidate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunValidateJSON(t *testing.T) {
	var buf bytes.Buffer

	opts := ValidateOptions{OutputFormat: output.FormatJSON}

	err := RunValidate(&buf, []string{"name = \"test\"\nvalue = 123"}, opts)
	if err != nil {
		t.Fatalf("RunValidate() error = %v", err)
	}

	var result ValidateResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if !result.Valid {
		t.Errorf("Valid = false, want true")
	}
}

func TestRunValidateJSONInvalid(t *testing.T) {
	var buf bytes.Buffer

	opts := ValidateOptions{OutputFormat: output.FormatJSON}

	_ = RunValidate(&buf, []string{"name = \"unclosed"}, opts)

	var result ValidateResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if result.Valid {
		t.Errorf("Valid = true, want false")
	}

	if result.Error == "" {
		t.Errorf("Error should not be empty")
	}
}

func TestRunFormat(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		opts    FormatOptions
		wantErr bool
	}{
		{
			name:    "format simple toml",
			input:   "name = \"test\"\nvalue = 123",
			opts:    FormatOptions{},
			wantErr: false,
		},
		{
			name:    "format with section",
			input:   "[database]\nhost = \"localhost\"",
			opts:    FormatOptions{},
			wantErr: false,
		},
		{
			name:    "format with indent",
			input:   "[database]\nhost = \"localhost\"",
			opts:    FormatOptions{Indent: 4},
			wantErr: false,
		},
		{
			name:    "invalid toml",
			input:   "name = \"unclosed",
			opts:    FormatOptions{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := RunFormat(&buf, []string{tt.input}, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunFormat() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunFormatPreservesData(t *testing.T) {
	input := "[database]\nhost = \"localhost\"\nport = 5432\n\n[server]\nname = \"main\""

	var buf bytes.Buffer

	err := RunFormat(&buf, []string{input}, FormatOptions{})
	if err != nil {
		t.Fatalf("RunFormat() error = %v", err)
	}

	output := buf.String()

	// Check that key data is preserved
	if !strings.Contains(output, "host") {
		t.Errorf("Output should contain 'host'")
	}

	if !strings.Contains(output, "localhost") {
		t.Errorf("Output should contain 'localhost'")
	}

	if !strings.Contains(output, "5432") {
		t.Errorf("Output should contain '5432'")
	}
}
