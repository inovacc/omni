package tagfixer

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"HelloWorld", "helloWorld"},
		{"hello_world", "helloWorld"},
		{"hello-world", "helloWorld"},
		{"HELLO_WORLD", "helloWorld"},
		{"ID", "id"},
		{"UserID", "userId"},
		{"user_id", "userId"},
		{"HTTPServer", "httpServer"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ToCamelCase(tt.input); got != tt.expected {
				t.Errorf("ToCamelCase(%s) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"helloWorld", "HelloWorld"},
		{"hello_world", "HelloWorld"},
		{"hello-world", "HelloWorld"},
		{"id", "Id"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ToPascalCase(tt.input); got != tt.expected {
				t.Errorf("ToPascalCase(%s) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"HelloWorld", "hello_world"},
		{"helloWorld", "hello_world"},
		{"hello-world", "hello_world"},
		{"ID", "id"},
		{"UserID", "user_id"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ToSnakeCase(tt.input); got != tt.expected {
				t.Errorf("ToSnakeCase(%s) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}

func TestToKebabCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"HelloWorld", "hello-world"},
		{"helloWorld", "hello-world"},
		{"hello_world", "hello-world"},
		{"ID", "id"},
		{"UserID", "user-id"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ToKebabCase(tt.input); got != tt.expected {
				t.Errorf("ToKebabCase(%s) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}

func TestDetectCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"helloWorld", "camelCase"},
		{"HelloWorld", "PascalCase"},
		{"hello_world", "snake_case"},
		{"hello-world", "kebab-case"},
		{"hello", "lowercase"},
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
		{"camel", CaseCamel, false},
		{"pascal", CasePascal, false},
		{"snake", CaseSnake, false},
		{"kebab", CaseKebab, false},
		{"CAMEL", CaseCamel, false},
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

func TestSplitIntoWords(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"helloWorld", []string{"hello", "World"}},
		{"HelloWorld", []string{"Hello", "World"}},
		{"hello_world", []string{"hello", "world"}},
		{"hello-world", []string{"hello", "world"}},
		{"ID", []string{"ID"}},
		{"UserID", []string{"User", "ID"}},
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

func TestExtractTagValue(t *testing.T) {
	tests := []struct {
		tag      string
		key      string
		expected string
	}{
		{`json:"name"`, "json", "name"},
		{`json:"name,omitempty"`, "json", "name,omitempty"},
		{`json:"name" yaml:"n"`, "json", "name"},
		{`json:"name" yaml:"n"`, "yaml", "n"},
		{`json:"-"`, "json", "-"},
		{`json:"name"`, "yaml", ""},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			if got := extractTagValue(tt.tag, tt.key); got != tt.expected {
				t.Errorf("extractTagValue(%s, %s) = %s, want %s", tt.tag, tt.key, got, tt.expected)
			}
		})
	}
}

func TestFixTag(t *testing.T) {
	tests := []struct {
		name       string
		tag        string
		fieldName  string
		tags       []string
		targetCase CaseType
		expected   string
		changes    int
	}{
		{
			name:       "camel to snake",
			tag:        "`json:\"userName\"`",
			fieldName:  "UserName",
			tags:       []string{"json"},
			targetCase: CaseSnake,
			expected:   "`json:\"user_name\"`",
			changes:    1,
		},
		{
			name:       "keep omitempty",
			tag:        "`json:\"userName,omitempty\"`",
			fieldName:  "UserName",
			tags:       []string{"json"},
			targetCase: CaseSnake,
			expected:   "`json:\"user_name,omitempty\"`",
			changes:    1,
		},
		{
			name:       "skip dash",
			tag:        "`json:\"-\"`",
			fieldName:  "UserName",
			tags:       []string{"json"},
			targetCase: CaseSnake,
			expected:   "`json:\"-\"`",
			changes:    0,
		},
		{
			name:       "multiple tags",
			tag:        "`json:\"userName\" yaml:\"UserName\"`",
			fieldName:  "UserName",
			tags:       []string{"json", "yaml"},
			targetCase: CaseSnake,
			expected:   "`json:\"user_name\" yaml:\"user_name\"`",
			changes:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, changes := fixTag(tt.tag, tt.fieldName, "TestStruct", tt.tags, tt.targetCase)
			if got != tt.expected {
				t.Errorf("fixTag() = %s, want %s", got, tt.expected)
			}
			if len(changes) != tt.changes {
				t.Errorf("fixTag() changes = %d, want %d", len(changes), tt.changes)
			}
		})
	}
}

func TestRunTagFixerDryRun(t *testing.T) {
	// Create temp directory with a test file
	tmpDir, err := os.MkdirTemp("", "tagfixer_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "test.go")
	content := `package test

type User struct {
	UserName string ` + "`json:\"UserName\"`" + `
	UserID   int    ` + "`json:\"UserID\"`" + `
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	opts := Options{
		Path:   tmpDir,
		Case:   CaseCamel,
		Tags:   []string{"json"},
		DryRun: true,
	}

	if err := RunTagFixer(&buf, opts); err != nil {
		t.Fatalf("RunTagFixer() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "UserName -> userName") {
		t.Errorf("Expected output to contain change, got: %s", output)
	}

	// Verify file wasn't modified
	data, _ := os.ReadFile(testFile)
	if !strings.Contains(string(data), `json:"UserName"`) {
		t.Error("File should not be modified in dry-run mode")
	}
}

func TestRunTagFixerAnalyze(t *testing.T) {
	// Create temp directory with a test file
	tmpDir, err := os.MkdirTemp("", "tagfixer_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "test.go")
	content := `package test

type User struct {
	UserName string ` + "`json:\"userName\"`" + `
	UserID   int    ` + "`json:\"userId\"`" + `
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	opts := Options{
		Path:    tmpDir,
		Tags:    []string{"json"},
		Analyze: true,
	}

	if err := RunTagFixer(&buf, opts); err != nil {
		t.Fatalf("RunTagFixer() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Analysis Report") {
		t.Errorf("Expected analysis report, got: %s", output)
	}
	if !strings.Contains(output, "json:") {
		t.Errorf("Expected json stats, got: %s", output)
	}
}

func TestRunTagFixerJSON(t *testing.T) {
	// Create temp directory with a test file
	tmpDir, err := os.MkdirTemp("", "tagfixer_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "test.go")
	content := `package test

type User struct {
	UserName string ` + "`json:\"UserName\"`" + `
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	opts := Options{
		Path:   tmpDir,
		Case:   CaseCamel,
		Tags:   []string{"json"},
		DryRun: true,
		JSON:   true,
	}

	if err := RunTagFixer(&buf, opts); err != nil {
		t.Fatalf("RunTagFixer() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"files"`) {
		t.Errorf("Expected JSON output, got: %s", output)
	}
}

func TestValidCaseTypes(t *testing.T) {
	types := ValidCaseTypes()
	if len(types) != 4 {
		t.Errorf("ValidCaseTypes() = %d types, want 4", len(types))
	}
}

func TestGetSortedTags(t *testing.T) {
	tags := GetSortedTags()
	if len(tags) == 0 {
		t.Error("GetSortedTags() returned empty")
	}
	if tags[0] != "json" {
		t.Errorf("GetSortedTags()[0] = %s, want json", tags[0])
	}
}

func TestCollectGoFiles(t *testing.T) {
	// Create temp directory structure
	tmpDir, err := os.MkdirTemp("", "tagfixer_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create files
	_ = os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "main_test.go"), []byte("package main"), 0644)
	_ = os.MkdirAll(filepath.Join(tmpDir, "sub"), 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "sub", "sub.go"), []byte("package sub"), 0644)

	// Test recursive
	files, err := collectGoFiles(tmpDir, true)
	if err != nil {
		t.Fatalf("collectGoFiles() error = %v", err)
	}

	if len(files) != 2 { // main.go and sub/sub.go (not _test.go)
		t.Errorf("collectGoFiles() = %d files, want 2", len(files))
	}

	// Test non-recursive
	files, err = collectGoFiles(tmpDir, false)
	if err != nil {
		t.Fatalf("collectGoFiles() error = %v", err)
	}

	if len(files) != 1 { // only main.go
		t.Errorf("collectGoFiles() non-recursive = %d files, want 1", len(files))
	}
}

func TestCalculateConsistency(t *testing.T) {
	tests := []struct {
		name       string
		caseStats  map[string]int
		minScore   float64
		maxScore   float64
		recommended CaseType
	}{
		{
			name:       "all camel",
			caseStats:  map[string]int{"camelCase": 10},
			minScore:   1.0,
			maxScore:   1.0,
			recommended: CaseCamel,
		},
		{
			name:       "mostly snake",
			caseStats:  map[string]int{"snake_case": 8, "camelCase": 2},
			minScore:   0.7,
			maxScore:   0.9,
			recommended: CaseSnake,
		},
		{
			name:       "empty",
			caseStats:  map[string]int{},
			minScore:   1.0,
			maxScore:   1.0,
			recommended: CaseCamel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, recommended := calculateConsistency(tt.caseStats)
			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("calculateConsistency() score = %f, want [%f, %f]", score, tt.minScore, tt.maxScore)
			}
			if recommended != tt.recommended {
				t.Errorf("calculateConsistency() recommended = %s, want %s", recommended, tt.recommended)
			}
		})
	}
}
