package brdoc

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestGenerateCPF(t *testing.T) {
	cpf := GenerateCPF()
	if len(cpf) != 11 {
		t.Errorf("GenerateCPF() length = %d, want 11", len(cpf))
	}

	if !ValidateCPF(cpf) {
		t.Errorf("GenerateCPF() generated invalid CPF: %s", cpf)
	}
}

func TestGenerateCPFFormatted(t *testing.T) {
	cpf := GenerateCPFFormatted()
	if len(cpf) != 14 { // XXX.XXX.XXX-XX
		t.Errorf("GenerateCPFFormatted() length = %d, want 14", len(cpf))
	}

	if !strings.Contains(cpf, ".") || !strings.Contains(cpf, "-") {
		t.Errorf("GenerateCPFFormatted() not formatted: %s", cpf)
	}
}

func TestValidateCPF(t *testing.T) {
	tests := []struct {
		name  string
		cpf   string
		valid bool
	}{
		{"valid unformatted", "52998224725", true},
		{"valid formatted", "529.982.247-25", true},
		{"invalid checksum", "12345678900", false},
		{"all same digits", "11111111111", false},
		{"too short", "1234567890", false},
		{"too long", "123456789012", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateCPF(tt.cpf); got != tt.valid {
				t.Errorf("ValidateCPF(%s) = %v, want %v", tt.cpf, got, tt.valid)
			}
		})
	}
}

func TestFormatCPF(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"52998224725", "529.982.247-25"},
		{"529.982.247-25", "529.982.247-25"},
		{"529982247-25", "529.982.247-25"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := FormatCPF(tt.input); got != tt.expected {
				t.Errorf("FormatCPF(%s) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}

func TestCPFState(t *testing.T) {
	// CPF with fiscal digit 0 should be from RS
	state := CPFState("52998224705")
	if state == "" {
		t.Error("CPFState() returned empty string")
	}
}

func TestGenerateCNPJ(t *testing.T) {
	cnpj := GenerateCNPJ()
	if len(cnpj) != 14 {
		t.Errorf("GenerateCNPJ() length = %d, want 14", len(cnpj))
	}

	if !ValidateCNPJ(cnpj) {
		t.Errorf("GenerateCNPJ() generated invalid CNPJ: %s", cnpj)
	}
}

func TestGenerateCNPJLegacy(t *testing.T) {
	cnpj := GenerateCNPJLegacy()
	if len(cnpj) != 14 {
		t.Errorf("GenerateCNPJLegacy() length = %d, want 14", len(cnpj))
	}

	// Legacy should be all digits
	for _, c := range cnpj {
		if c < '0' || c > '9' {
			t.Errorf("GenerateCNPJLegacy() contains non-digit: %s", cnpj)
			break
		}
	}

	if !ValidateCNPJ(cnpj) {
		t.Errorf("GenerateCNPJLegacy() generated invalid CNPJ: %s", cnpj)
	}
}

func TestValidateCNPJ(t *testing.T) {
	tests := []struct {
		name  string
		cnpj  string
		valid bool
	}{
		{"valid unformatted", "11222333000181", true},
		{"valid formatted", "11.222.333/0001-81", true},
		{"invalid checksum", "11222333000100", false},
		{"all same digits", "11111111111111", false},
		{"too short", "1122233300018", false},
		{"too long", "112223330001811", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateCNPJ(tt.cnpj); got != tt.valid {
				t.Errorf("ValidateCNPJ(%s) = %v, want %v", tt.cnpj, got, tt.valid)
			}
		})
	}
}

func TestFormatCNPJ(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"11222333000181", "11.222.333/0001-81"},
		{"11.222.333/0001-81", "11.222.333/0001-81"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := FormatCNPJ(tt.input); got != tt.expected {
				t.Errorf("FormatCNPJ(%s) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}

func TestRunCPFGenerate(t *testing.T) {
	var buf bytes.Buffer

	opts := Options{Generate: true, Count: 3}

	err := RunCPF(&buf, nil, opts)
	if err != nil {
		t.Fatalf("RunCPF() error = %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Errorf("RunCPF() generated %d CPFs, want 3", len(lines))
	}

	for _, line := range lines {
		if len(line) != 14 { // formatted length
			t.Errorf("Generated CPF has wrong length: %s", line)
		}
	}
}

func TestRunCPFGenerateJSON(t *testing.T) {
	var buf bytes.Buffer

	opts := Options{Generate: true, Count: 2, JSON: true}

	err := RunCPF(&buf, nil, opts)
	if err != nil {
		t.Fatalf("RunCPF() error = %v", err)
	}

	var result CPFListResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("JSON unmarshal error = %v", err)
	}

	if result.Count != 2 {
		t.Errorf("Count = %d, want 2", result.Count)
	}

	if len(result.CPFs) != 2 {
		t.Errorf("CPFs length = %d, want 2", len(result.CPFs))
	}
}

func TestRunCPFValidate(t *testing.T) {
	var buf bytes.Buffer

	opts := Options{Validate: true}

	err := RunCPF(&buf, []string{"529.982.247-25"}, opts)
	if err != nil {
		t.Fatalf("RunCPF() error = %v", err)
	}

	if !strings.Contains(buf.String(), "valid") {
		t.Errorf("RunCPF() output should contain 'valid': %s", buf.String())
	}
}

func TestRunCPFValidateInvalid(t *testing.T) {
	var buf bytes.Buffer

	opts := Options{Validate: true}

	err := RunCPF(&buf, []string{"12345678900"}, opts)
	if err == nil {
		t.Error("RunCPF() should return error for invalid CPF")
	}

	if !strings.Contains(buf.String(), "invalid") {
		t.Errorf("RunCPF() output should contain 'invalid': %s", buf.String())
	}
}

func TestRunCNPJGenerate(t *testing.T) {
	var buf bytes.Buffer

	opts := Options{Generate: true, Count: 3}

	err := RunCNPJ(&buf, nil, opts)
	if err != nil {
		t.Fatalf("RunCNPJ() error = %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Errorf("RunCNPJ() generated %d CNPJs, want 3", len(lines))
	}
}

func TestRunCNPJGenerateLegacy(t *testing.T) {
	var buf bytes.Buffer

	opts := Options{Generate: true, Count: 1, Legacy: true}

	err := RunCNPJ(&buf, nil, opts)
	if err != nil {
		t.Fatalf("RunCNPJ() error = %v", err)
	}

	output := strings.TrimSpace(buf.String())
	// Legacy CNPJ formatted: XX.XXX.XXX/XXXX-XX (18 chars)
	if len(output) != 18 {
		t.Errorf("Legacy CNPJ length = %d, want 18", len(output))
	}
}

func TestRunCNPJGenerateJSON(t *testing.T) {
	var buf bytes.Buffer

	opts := Options{Generate: true, Count: 2, JSON: true}

	err := RunCNPJ(&buf, nil, opts)
	if err != nil {
		t.Fatalf("RunCNPJ() error = %v", err)
	}

	var result CNPJListResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("JSON unmarshal error = %v", err)
	}

	if result.Count != 2 {
		t.Errorf("Count = %d, want 2", result.Count)
	}

	if len(result.CNPJs) != 2 {
		t.Errorf("CNPJs length = %d, want 2", len(result.CNPJs))
	}
}

func TestRunCNPJValidate(t *testing.T) {
	var buf bytes.Buffer

	opts := Options{Validate: true}

	err := RunCNPJ(&buf, []string{"11.222.333/0001-81"}, opts)
	if err != nil {
		t.Fatalf("RunCNPJ() error = %v", err)
	}

	if !strings.Contains(buf.String(), "valid") {
		t.Errorf("RunCNPJ() output should contain 'valid': %s", buf.String())
	}
}

func TestRunCNPJValidateInvalid(t *testing.T) {
	var buf bytes.Buffer

	opts := Options{Validate: true}

	err := RunCNPJ(&buf, []string{"11222333000100"}, opts)
	if err == nil {
		t.Error("RunCNPJ() should return error for invalid CNPJ")
	}

	if !strings.Contains(buf.String(), "invalid") {
		t.Errorf("RunCNPJ() output should contain 'invalid': %s", buf.String())
	}
}

func TestCleanDoc(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"123.456.789-09", "12345678909"},
		{"11.222.333/0001-81", "11222333000181"},
		{"12345678909", "12345678909"},
		{"123 456 789", "123456789"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := cleanDoc(tt.input); got != tt.expected {
				t.Errorf("cleanDoc(%s) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}

func TestRunCPFNoArgs(t *testing.T) {
	var buf bytes.Buffer

	opts := Options{Validate: true}

	err := RunCPF(&buf, nil, opts)
	if err == nil {
		t.Error("RunCPF() validate without args should error")
	}
}

func TestRunCNPJNoArgs(t *testing.T) {
	var buf bytes.Buffer

	opts := Options{Validate: true}

	err := RunCNPJ(&buf, nil, opts)
	if err == nil {
		t.Error("RunCNPJ() validate without args should error")
	}
}

func TestCPFUniqueness(t *testing.T) {
	seen := make(map[string]bool)
	count := 100

	for range count {
		cpf := GenerateCPF()
		if seen[cpf] {
			t.Errorf("Duplicate CPF generated: %s", cpf)
		}

		seen[cpf] = true
	}
}

func TestCNPJUniqueness(t *testing.T) {
	seen := make(map[string]bool)
	count := 100

	for range count {
		cnpj := GenerateCNPJ()
		if seen[cnpj] {
			t.Errorf("Duplicate CNPJ generated: %s", cnpj)
		}

		seen[cnpj] = true
	}
}
