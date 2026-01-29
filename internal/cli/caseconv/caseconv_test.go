package caseconv

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestToUpper(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "HELLO"},
		{"Hello World", "HELLO WORLD"},
		{"HELLO", "HELLO"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ToUpper(tt.input); got != tt.expected {
				t.Errorf("ToUpper(%s) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}

func TestToLower(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"HELLO", "hello"},
		{"Hello World", "hello world"},
		{"hello", "hello"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ToLower(tt.input); got != tt.expected {
				t.Errorf("ToLower(%s) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}

func TestToTitle(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello world", "Hello World"},
		{"HELLO WORLD", "Hello World"},
		{"hello", "Hello"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ToTitle(tt.input); got != tt.expected {
				t.Errorf("ToTitle(%s) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}

func TestToSentence(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello world", "Hello world"},
		{"HELLO WORLD", "Hello world"},
		{"hello", "Hello"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ToSentence(tt.input); got != tt.expected {
				t.Errorf("ToSentence(%s) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}

func TestToCamel(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello world", "helloWorld"},
		{"Hello World", "helloWorld"},
		{"hello_world", "helloWorld"},
		{"hello-world", "helloWorld"},
		{"HelloWorld", "helloWorld"},
		{"HELLO_WORLD", "helloWorld"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ToCamel(tt.input); got != tt.expected {
				t.Errorf("ToCamel(%s) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}

func TestToPascal(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello world", "HelloWorld"},
		{"hello_world", "HelloWorld"},
		{"hello-world", "HelloWorld"},
		{"helloWorld", "HelloWorld"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ToPascal(tt.input); got != tt.expected {
				t.Errorf("ToPascal(%s) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}

func TestToSnake(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello world", "hello_world"},
		{"HelloWorld", "hello_world"},
		{"helloWorld", "hello_world"},
		{"hello-world", "hello_world"},
		{"HELLO_WORLD", "hello_world"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ToSnake(tt.input); got != tt.expected {
				t.Errorf("ToSnake(%s) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}

func TestToKebab(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello world", "hello-world"},
		{"HelloWorld", "hello-world"},
		{"helloWorld", "hello-world"},
		{"hello_world", "hello-world"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ToKebab(tt.input); got != tt.expected {
				t.Errorf("ToKebab(%s) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}

func TestToConstant(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello world", "HELLO_WORLD"},
		{"helloWorld", "HELLO_WORLD"},
		{"HelloWorld", "HELLO_WORLD"},
		{"hello_world", "HELLO_WORLD"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ToConstant(tt.input); got != tt.expected {
				t.Errorf("ToConstant(%s) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}

func TestToDot(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello world", "hello.world"},
		{"helloWorld", "hello.world"},
		{"hello_world", "hello.world"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ToDot(tt.input); got != tt.expected {
				t.Errorf("ToDot(%s) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}

func TestToPath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello world", "hello/world"},
		{"helloWorld", "hello/world"},
		{"hello_world", "hello/world"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ToPath(tt.input); got != tt.expected {
				t.Errorf("ToPath(%s) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}

func TestToSwap(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", "hELLO wORLD"},
		{"HELLO", "hello"},
		{"hello", "HELLO"},
		{"HeLLo", "hEllO"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ToSwap(tt.input); got != tt.expected {
				t.Errorf("ToSwap(%s) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}

func TestToToggle(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "Hello"},
		{"Hello", "hello"},
		{"HELLO", "hELLO"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ToToggle(tt.input); got != tt.expected {
				t.Errorf("ToToggle(%s) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}

func TestDetectCase(t *testing.T) {
	tests := []struct {
		input    string
		expected CaseType
	}{
		{"helloWorld", CaseCamel},
		{"HelloWorld", CasePascal},
		{"hello_world", CaseSnake},
		{"hello-world", CaseKebab},
		{"HELLO_WORLD", CaseConstant},
		{"hello.world", CaseDot},
		{"hello/world", CasePath},
		{"HELLO", CaseUpper},
		{"hello", CaseLower},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := DetectCase(tt.input); got != tt.expected {
				t.Errorf("DetectCase(%s) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseCaseType(t *testing.T) {
	tests := []struct {
		input    string
		expected CaseType
		wantErr  bool
	}{
		{"upper", CaseUpper, false},
		{"lower", CaseLower, false},
		{"camel", CaseCamel, false},
		{"pascal", CasePascal, false},
		{"snake", CaseSnake, false},
		{"kebab", CaseKebab, false},
		{"constant", CaseConstant, false},
		{"UPPER", CaseUpper, false},
		{"invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseCaseType(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCaseType(%s) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.expected {
				t.Errorf("ParseCaseType(%s) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}

func TestConvert(t *testing.T) {
	input := "hello world"
	tests := []struct {
		caseType CaseType
		expected string
	}{
		{CaseUpper, "HELLO WORLD"},
		{CaseLower, "hello world"},
		{CaseCamel, "helloWorld"},
		{CasePascal, "HelloWorld"},
		{CaseSnake, "hello_world"},
		{CaseKebab, "hello-world"},
		{CaseConstant, "HELLO_WORLD"},
		{CaseDot, "hello.world"},
		{CasePath, "hello/world"},
	}

	for _, tt := range tests {
		t.Run(string(tt.caseType), func(t *testing.T) {
			if got := Convert(input, tt.caseType); got != tt.expected {
				t.Errorf("Convert(%s, %s) = %s, want %s", input, tt.caseType, got, tt.expected)
			}
		})
	}
}

func TestConvertAll(t *testing.T) {
	input := "hello world"
	result := ConvertAll(input)

	if len(result) != len(ValidCaseTypes()) {
		t.Errorf("ConvertAll() returned %d results, want %d", len(result), len(ValidCaseTypes()))
	}

	if result[CaseCamel] != "helloWorld" {
		t.Errorf("ConvertAll()[camel] = %s, want helloWorld", result[CaseCamel])
	}
}

func TestValidCaseTypes(t *testing.T) {
	types := ValidCaseTypes()
	if len(types) != 13 {
		t.Errorf("ValidCaseTypes() = %d types, want 13", len(types))
	}
}

func TestRunCase(t *testing.T) {
	var buf bytes.Buffer

	opts := Options{Case: CaseUpper}
	err := RunCase(&buf, []string{"hello world"}, opts)
	if err != nil {
		t.Fatalf("RunCase() error = %v", err)
	}

	expected := "HELLO WORLD\n"
	if buf.String() != expected {
		t.Errorf("RunCase() = %s, want %s", buf.String(), expected)
	}
}

func TestRunCaseJSON(t *testing.T) {
	var buf bytes.Buffer

	opts := Options{Case: CaseCamel, JSON: true}
	err := RunCase(&buf, []string{"hello world"}, opts)
	if err != nil {
		t.Fatalf("RunCase() error = %v", err)
	}

	var result ListResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("JSON unmarshal error = %v", err)
	}

	if result.Case != "camel" {
		t.Errorf("result.Case = %s, want camel", result.Case)
	}

	if len(result.Results) != 1 {
		t.Errorf("result.Results length = %d, want 1", len(result.Results))
	}

	if result.Results[0].Output != "helloWorld" {
		t.Errorf("result.Results[0].Output = %s, want helloWorld", result.Results[0].Output)
	}
}

func TestRunCaseMultiple(t *testing.T) {
	var buf bytes.Buffer

	opts := Options{Case: CaseUpper}
	err := RunCase(&buf, []string{"hello", "world"}, opts)
	if err != nil {
		t.Fatalf("RunCase() error = %v", err)
	}

	expected := "HELLO\nWORLD\n"
	if buf.String() != expected {
		t.Errorf("RunCase() = %s, want %s", buf.String(), expected)
	}
}

func TestSplitIntoWords(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"helloWorld", []string{"hello", "World"}},
		{"HelloWorld", []string{"Hello", "World"}},
		{"hello_world", []string{"hello", "world"}},
		{"hello-world", []string{"hello", "world"}},
		{"hello.world", []string{"hello", "world"}},
		{"hello/world", []string{"hello", "world"}},
		{"hello world", []string{"hello", "world"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := splitIntoWords(tt.input)
			if len(got) != len(tt.expected) {
				t.Errorf("splitIntoWords(%s) = %v, want %v", tt.input, got, tt.expected)
				return
			}
			for i, word := range got {
				if word != tt.expected[i] {
					t.Errorf("splitIntoWords(%s)[%d] = %s, want %s", tt.input, i, word, tt.expected[i])
				}
			}
		})
	}
}
