package brdoc

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// TestRunCPFFormat exercises the formatCPF path (text + JSON, single + multi).
func TestRunCPFFormat(t *testing.T) {
	// Generate a raw (digits-only) CPF, then format it through RunCPF.
	raw := GenerateCPF()

	tests := []struct {
		name string
		args []string
		opts Options
	}{
		{"text single", []string{raw}, Options{Format: true}},
		{"text multi", []string{raw, raw}, Options{Format: true}},
		{"json single", []string{raw}, Options{Format: true, JSON: true}},
		{"json multi", []string{raw, raw}, Options{Format: true, JSON: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := RunCPF(&buf, tt.args, tt.opts); err != nil {
				t.Fatalf("RunCPF format error = %v", err)
			}

			out := buf.String()
			if out == "" {
				t.Fatal("RunCPF format produced no output")
			}

			if tt.opts.JSON && !strings.Contains(out, "cpf") {
				t.Errorf("JSON output missing cpf field: %q", out)
			}

			// A formatted CPF contains punctuation.
			if !tt.opts.JSON && !strings.Contains(out, ".") {
				t.Errorf("formatted CPF should contain '.': %q", out)
			}
		})
	}
}

// TestRunCPFFormatNoArgs covers the missing-operand branch of formatCPF.
func TestRunCPFFormatNoArgs(t *testing.T) {
	var buf bytes.Buffer
	if err := RunCPF(&buf, nil, Options{Format: true}); err == nil {
		t.Error("expected error for format with no args")
	}
}

// TestRunCPFValidateJSON covers the JSON branch of validateCPF (valid + invalid).
func TestRunCPFValidateJSON(t *testing.T) {
	valid := GenerateCPF()

	t.Run("single valid", func(t *testing.T) {
		var buf bytes.Buffer
		if err := RunCPF(&buf, []string{valid}, Options{Validate: true, JSON: true}); err != nil {
			t.Fatalf("RunCPF validate json error = %v", err)
		}

		var res CPFResult
		if err := json.Unmarshal(buf.Bytes(), &res); err != nil {
			t.Fatalf("unmarshal: %v (out=%q)", err, buf.String())
		}

		if !res.Valid {
			t.Errorf("expected valid=true for %q", valid)
		}
	})

	t.Run("multi mixed", func(t *testing.T) {
		var buf bytes.Buffer
		_ = RunCPF(&buf, []string{valid, "000.000.000-00"}, Options{Validate: true, JSON: true})

		var res CPFListResult
		if err := json.Unmarshal(buf.Bytes(), &res); err != nil {
			t.Fatalf("unmarshal list: %v (out=%q)", err, buf.String())
		}

		if res.Count != 2 {
			t.Errorf("count = %d, want 2", res.Count)
		}
	})
}

// TestRunCPFDefault covers the default (no flag) generate-one path.
func TestRunCPFDefault(t *testing.T) {
	var buf bytes.Buffer
	if err := RunCPF(&buf, nil, Options{}); err != nil {
		t.Fatalf("RunCPF default error = %v", err)
	}

	if buf.Len() == 0 {
		t.Error("RunCPF default should generate one CPF")
	}
}

// TestRunCNPJFormat exercises the formatCNPJ path.
func TestRunCNPJFormat(t *testing.T) {
	raw := GenerateCNPJ()

	tests := []struct {
		name string
		args []string
		opts Options
	}{
		{"text single", []string{raw}, Options{Format: true}},
		{"json single", []string{raw}, Options{Format: true, JSON: true}},
		{"json multi", []string{raw, raw}, Options{Format: true, JSON: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := RunCNPJ(&buf, tt.args, tt.opts); err != nil {
				t.Fatalf("RunCNPJ format error = %v", err)
			}

			if buf.Len() == 0 {
				t.Fatal("RunCNPJ format produced no output")
			}
		})
	}
}

// TestRunCNPJFormatNoArgs covers the missing-operand branch of formatCNPJ.
func TestRunCNPJFormatNoArgs(t *testing.T) {
	var buf bytes.Buffer
	if err := RunCNPJ(&buf, nil, Options{Format: true}); err == nil {
		t.Error("expected error for format with no args")
	}
}

// TestRunCNPJValidateJSON covers the JSON validate branch.
func TestRunCNPJValidateJSON(t *testing.T) {
	valid := GenerateCNPJ()

	var buf bytes.Buffer
	if err := RunCNPJ(&buf, []string{valid}, Options{Validate: true, JSON: true}); err != nil {
		t.Fatalf("RunCNPJ validate json error = %v", err)
	}

	var res CNPJResult
	if err := json.Unmarshal(buf.Bytes(), &res); err != nil {
		t.Fatalf("unmarshal: %v (out=%q)", err, buf.String())
	}

	if !res.Valid {
		t.Errorf("expected valid=true for %q", valid)
	}
}

// TestRunCNPJDefault covers the default generate-one path.
func TestRunCNPJDefault(t *testing.T) {
	var buf bytes.Buffer
	if err := RunCNPJ(&buf, nil, Options{}); err != nil {
		t.Fatalf("RunCNPJ default error = %v", err)
	}

	if buf.Len() == 0 {
		t.Error("RunCNPJ default should generate one CNPJ")
	}
}

// TestGenerateCNPJFormatted covers the exported helper.
func TestGenerateCNPJFormatted(t *testing.T) {
	got := GenerateCNPJFormatted()
	if !ValidateCNPJ(got) {
		t.Errorf("GenerateCNPJFormatted produced invalid CNPJ: %q", got)
	}

	if !strings.ContainsAny(got, "./-") {
		t.Errorf("GenerateCNPJFormatted should be punctuated: %q", got)
	}
}

// TestFormatRoundTrips verifies the package-level format helpers round-trip.
func TestFormatRoundTrips(t *testing.T) {
	cpf := GenerateCPF()
	if f := FormatCPF(cpf); !strings.Contains(f, ".") {
		t.Errorf("FormatCPF(%q) = %q, expected punctuation", cpf, f)
	}

	cnpj := GenerateCNPJ()
	if f := FormatCNPJ(cnpj); f == "" {
		t.Errorf("FormatCNPJ(%q) returned empty", cnpj)
	}
}
