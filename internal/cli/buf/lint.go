package buf

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

// LintRule represents a lint rule
type LintRule struct {
	ID       string
	Category string
	Check    func(*ProtoFile, string) []LintResult
}

// Lint rule categories
const (
	CategoryMinimal  = "MINIMAL"
	CategoryBasic    = "BASIC"
	CategoryStandard = "STANDARD"
	CategoryComments = "COMMENTS"
)

// All lint rules
var lintRules = []LintRule{
	// MINIMAL rules
	{ID: "DIRECTORY_SAME_PACKAGE", Category: CategoryMinimal, Check: checkDirectorySamePackage},
	{ID: "PACKAGE_DEFINED", Category: CategoryMinimal, Check: checkPackageDefined},
	{ID: "PACKAGE_DIRECTORY_MATCH", Category: CategoryMinimal, Check: checkPackageDirectoryMatch},
	{ID: "PACKAGE_SAME_DIRECTORY", Category: CategoryMinimal, Check: checkPackageSameDirectory},

	// BASIC rules
	{ID: "ENUM_FIRST_VALUE_ZERO", Category: CategoryBasic, Check: checkEnumFirstValueZero},
	{ID: "ENUM_NO_ALLOW_ALIAS", Category: CategoryBasic, Check: checkEnumNoAllowAlias},
	{ID: "ENUM_PASCAL_CASE", Category: CategoryBasic, Check: checkEnumPascalCase},
	{ID: "ENUM_VALUE_UPPER_SNAKE_CASE", Category: CategoryBasic, Check: checkEnumValueUpperSnakeCase},
	{ID: "FIELD_LOWER_SNAKE_CASE", Category: CategoryBasic, Check: checkFieldLowerSnakeCase},
	{ID: "IMPORT_NO_PUBLIC", Category: CategoryBasic, Check: checkImportNoPublic},
	{ID: "IMPORT_NO_WEAK", Category: CategoryBasic, Check: checkImportNoWeak},
	{ID: "MESSAGE_PASCAL_CASE", Category: CategoryBasic, Check: checkMessagePascalCase},
	{ID: "ONEOF_LOWER_SNAKE_CASE", Category: CategoryBasic, Check: checkOneofLowerSnakeCase},
	{ID: "PACKAGE_LOWER_SNAKE_CASE", Category: CategoryBasic, Check: checkPackageLowerSnakeCase},
	{ID: "RPC_PASCAL_CASE", Category: CategoryBasic, Check: checkRPCPascalCase},
	{ID: "SERVICE_PASCAL_CASE", Category: CategoryBasic, Check: checkServicePascalCase},

	// STANDARD rules
	{ID: "ENUM_VALUE_PREFIX", Category: CategoryStandard, Check: checkEnumValuePrefix},
	{ID: "ENUM_ZERO_VALUE_SUFFIX", Category: CategoryStandard, Check: checkEnumZeroValueSuffix},
	{ID: "FILE_LOWER_SNAKE_CASE", Category: CategoryStandard, Check: checkFileLowerSnakeCase},
	{ID: "RPC_REQUEST_RESPONSE_UNIQUE", Category: CategoryStandard, Check: checkRPCRequestResponseUnique},
	{ID: "RPC_REQUEST_STANDARD_NAME", Category: CategoryStandard, Check: checkRPCRequestStandardName},
	{ID: "RPC_RESPONSE_STANDARD_NAME", Category: CategoryStandard, Check: checkRPCResponseStandardName},
	{ID: "PACKAGE_VERSION_SUFFIX", Category: CategoryStandard, Check: checkPackageVersionSuffix},
	{ID: "SERVICE_SUFFIX", Category: CategoryStandard, Check: checkServiceSuffix},

	// COMMENTS rules
	{ID: "COMMENT_ENUM", Category: CategoryComments, Check: checkCommentEnum},
	{ID: "COMMENT_ENUM_VALUE", Category: CategoryComments, Check: checkCommentEnumValue},
	{ID: "COMMENT_FIELD", Category: CategoryComments, Check: checkCommentField},
	{ID: "COMMENT_MESSAGE", Category: CategoryComments, Check: checkCommentMessage},
	{ID: "COMMENT_RPC", Category: CategoryComments, Check: checkCommentRPC},
	{ID: "COMMENT_SERVICE", Category: CategoryComments, Check: checkCommentService},
}

// RunLint runs lint on proto files
func RunLint(w io.Writer, dir string, opts LintOptions) error {
	// Load config
	config, err := LoadConfig(dir)
	if err != nil {
		return fmt.Errorf("buf: %w", err)
	}

	// Find proto files
	files, err := FindProtoFiles(dir, opts.ExcludePath)
	if err != nil {
		return fmt.Errorf("buf: %w", err)
	}

	if len(files) == 0 {
		_, _ = fmt.Fprintln(w, "No proto files found")

		return nil
	}

	// Get active rules based on config
	activeRules := getActiveRules(config.Lint)

	var allResults []LintResult

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			allResults = append(allResults, LintResult{
				File:    file,
				Line:    0,
				Column:  0,
				Rule:    "FILE_READ_ERROR",
				Message: err.Error(),
			})

			continue
		}

		protoFile, err := ParseProtoFile(string(content))
		if err != nil {
			allResults = append(allResults, LintResult{
				File:    file,
				Line:    0,
				Column:  0,
				Rule:    "PARSE_ERROR",
				Message: err.Error(),
			})

			continue
		}

		// Run each active rule
		relPath, _ := filepath.Rel(dir, file)

		for _, rule := range activeRules {
			// Check if rule should be ignored for this file
			if shouldIgnoreRule(config.Lint, rule.ID, relPath) {
				continue
			}

			results := rule.Check(protoFile, relPath)
			allResults = append(allResults, results...)
		}
	}

	if len(allResults) == 0 {
		return nil
	}

	// Output results
	return OutputResults(w, allResults, opts.ErrorFormat)
}

func getActiveRules(config LintConfig) []LintRule {
	// Default to STANDARD if no rules specified
	if len(config.Use) == 0 {
		config.Use = []string{CategoryStandard}
	}

	// Build set of categories to use
	categories := make(map[string]bool)
	for _, cat := range config.Use {
		categories[cat] = true
		// Include lower categories
		switch cat {
		case CategoryStandard:
			categories[CategoryBasic] = true
			categories[CategoryMinimal] = true
		case CategoryBasic:
			categories[CategoryMinimal] = true
		case CategoryComments:
			// Comments is standalone
		}
	}

	// Build set of rules to except
	exceptRules := make(map[string]bool)
	for _, rule := range config.Except {
		exceptRules[rule] = true
	}

	var activeRules []LintRule

	for _, rule := range lintRules {
		if categories[rule.Category] && !exceptRules[rule.ID] {
			activeRules = append(activeRules, rule)
		}
	}

	return activeRules
}

func shouldIgnoreRule(config LintConfig, ruleID, file string) bool {
	// Check global ignore
	for _, pattern := range config.Ignore {
		if matchPath(pattern, file) {
			return true
		}
	}

	// Check ignore_only
	if patterns, ok := config.IgnoreOnly[ruleID]; ok {
		for _, pattern := range patterns {
			if matchPath(pattern, file) {
				return true
			}
		}
	}

	return false
}

func matchPath(pattern, file string) bool {
	return strings.HasPrefix(file, pattern) || file == pattern
}

// Rule implementations

func checkDirectorySamePackage(_ *ProtoFile, _ string) []LintResult {
	// This requires checking multiple files in the same directory
	// Simplified: return no issues
	return nil
}

func checkPackageDefined(file *ProtoFile, path string) []LintResult {
	if file.Package == "" {
		return []LintResult{{
			File:    path,
			Line:    1,
			Column:  1,
			Rule:    "PACKAGE_DEFINED",
			Message: "Files must have a package defined",
		}}
	}

	return nil
}

func checkPackageDirectoryMatch(file *ProtoFile, path string) []LintResult {
	if file.Package == "" {
		return nil
	}

	// Get directory from path
	dir := filepath.Dir(path)
	dir = strings.ReplaceAll(dir, "\\", "/")

	// Convert package to expected directory
	expectedDir := strings.ReplaceAll(file.Package, ".", "/")

	if !strings.HasSuffix(dir, expectedDir) && dir != expectedDir {
		return []LintResult{{
			File:    path,
			Line:    1,
			Column:  1,
			Rule:    "PACKAGE_DIRECTORY_MATCH",
			Message: fmt.Sprintf("Package %q should match directory structure", file.Package),
		}}
	}

	return nil
}

func checkPackageSameDirectory(_ *ProtoFile, _ string) []LintResult {
	// This requires checking multiple files
	return nil
}

func checkEnumFirstValueZero(file *ProtoFile, path string) []LintResult {
	var results []LintResult

	for _, enum := range file.Enums {
		if len(enum.Values) > 0 && enum.Values[0].Number != 0 {
			results = append(results, LintResult{
				File:    path,
				Line:    enum.Values[0].Line,
				Column:  1,
				Rule:    "ENUM_FIRST_VALUE_ZERO",
				Message: fmt.Sprintf("Enum %q first value must be zero", enum.Name),
			})
		}
	}

	// Check nested enums in messages
	for _, msg := range file.Messages {
		results = append(results, checkEnumsInMessage(msg, path)...)
	}

	return results
}

func checkEnumsInMessage(msg ProtoMessage, path string) []LintResult {
	var results []LintResult

	for _, enum := range msg.Enums {
		if len(enum.Values) > 0 && enum.Values[0].Number != 0 {
			results = append(results, LintResult{
				File:    path,
				Line:    enum.Values[0].Line,
				Column:  1,
				Rule:    "ENUM_FIRST_VALUE_ZERO",
				Message: fmt.Sprintf("Enum %q first value must be zero", enum.Name),
			})
		}
	}

	for _, nested := range msg.Nested {
		results = append(results, checkEnumsInMessage(nested, path)...)
	}

	return results
}

func checkEnumNoAllowAlias(file *ProtoFile, path string) []LintResult {
	var results []LintResult

	for _, enum := range file.Enums {
		for _, opt := range enum.Options {
			if opt.Name == "allow_alias" && opt.Value == "true" {
				results = append(results, LintResult{
					File:    path,
					Line:    opt.Line,
					Column:  1,
					Rule:    "ENUM_NO_ALLOW_ALIAS",
					Message: fmt.Sprintf("Enum %q should not use allow_alias option", enum.Name),
				})
			}
		}
	}

	return results
}

func checkEnumPascalCase(file *ProtoFile, path string) []LintResult {
	var results []LintResult

	for _, enum := range file.Enums {
		if !isPascalCase(enum.Name) {
			results = append(results, LintResult{
				File:    path,
				Line:    enum.Line,
				Column:  1,
				Rule:    "ENUM_PASCAL_CASE",
				Message: fmt.Sprintf("Enum name %q should be PascalCase", enum.Name),
			})
		}
	}

	return results
}

func checkEnumValueUpperSnakeCase(file *ProtoFile, path string) []LintResult {
	var results []LintResult

	for _, enum := range file.Enums {
		for _, value := range enum.Values {
			if !isUpperSnakeCase(value.Name) {
				results = append(results, LintResult{
					File:    path,
					Line:    value.Line,
					Column:  1,
					Rule:    "ENUM_VALUE_UPPER_SNAKE_CASE",
					Message: fmt.Sprintf("Enum value %q should be UPPER_SNAKE_CASE", value.Name),
				})
			}
		}
	}

	return results
}

func checkFieldLowerSnakeCase(file *ProtoFile, path string) []LintResult {
	results := make([]LintResult, 0, len(file.Messages))

	for _, msg := range file.Messages {
		results = append(results, checkFieldsInMessage(msg, path)...)
	}

	return results
}

func checkFieldsInMessage(msg ProtoMessage, path string) []LintResult {
	var results []LintResult

	for _, field := range msg.Fields {
		if !isLowerSnakeCase(field.Name) {
			results = append(results, LintResult{
				File:    path,
				Line:    field.Line,
				Column:  1,
				Rule:    "FIELD_LOWER_SNAKE_CASE",
				Message: fmt.Sprintf("Field name %q should be lower_snake_case", field.Name),
			})
		}
	}

	for _, nested := range msg.Nested {
		results = append(results, checkFieldsInMessage(nested, path)...)
	}

	return results
}

func checkImportNoPublic(file *ProtoFile, path string) []LintResult {
	var results []LintResult

	for _, imp := range file.Imports {
		if imp.Public {
			results = append(results, LintResult{
				File:    path,
				Line:    imp.Line,
				Column:  1,
				Rule:    "IMPORT_NO_PUBLIC",
				Message: "Public imports are not recommended",
			})
		}
	}

	return results
}

func checkImportNoWeak(file *ProtoFile, path string) []LintResult {
	var results []LintResult

	for _, imp := range file.Imports {
		if imp.Weak {
			results = append(results, LintResult{
				File:    path,
				Line:    imp.Line,
				Column:  1,
				Rule:    "IMPORT_NO_WEAK",
				Message: "Weak imports are not recommended",
			})
		}
	}

	return results
}

func checkMessagePascalCase(file *ProtoFile, path string) []LintResult {
	results := make([]LintResult, 0, len(file.Messages))

	for _, msg := range file.Messages {
		results = append(results, checkMessagePascalCaseRecursive(msg, path)...)
	}

	return results
}

func checkMessagePascalCaseRecursive(msg ProtoMessage, path string) []LintResult {
	var results []LintResult

	if !isPascalCase(msg.Name) {
		results = append(results, LintResult{
			File:    path,
			Line:    msg.Line,
			Column:  1,
			Rule:    "MESSAGE_PASCAL_CASE",
			Message: fmt.Sprintf("Message name %q should be PascalCase", msg.Name),
		})
	}

	for _, nested := range msg.Nested {
		results = append(results, checkMessagePascalCaseRecursive(nested, path)...)
	}

	return results
}

func checkOneofLowerSnakeCase(_ *ProtoFile, _ string) []LintResult {
	// Simplified: oneof names are parsed as part of messages
	return nil
}

func checkPackageLowerSnakeCase(file *ProtoFile, path string) []LintResult {
	if file.Package == "" {
		return nil
	}

	parts := strings.SplitSeq(file.Package, ".")
	for part := range parts {
		if !isLowerSnakeCase(part) && !isLowerCase(part) {
			return []LintResult{{
				File:    path,
				Line:    1,
				Column:  1,
				Rule:    "PACKAGE_LOWER_SNAKE_CASE",
				Message: fmt.Sprintf("Package name %q should be lower_snake_case", file.Package),
			}}
		}
	}

	return nil
}

func checkRPCPascalCase(file *ProtoFile, path string) []LintResult {
	var results []LintResult

	for _, svc := range file.Services {
		for _, method := range svc.Methods {
			if !isPascalCase(method.Name) {
				results = append(results, LintResult{
					File:    path,
					Line:    method.Line,
					Column:  1,
					Rule:    "RPC_PASCAL_CASE",
					Message: fmt.Sprintf("RPC name %q should be PascalCase", method.Name),
				})
			}
		}
	}

	return results
}

func checkServicePascalCase(file *ProtoFile, path string) []LintResult {
	var results []LintResult

	for _, svc := range file.Services {
		if !isPascalCase(svc.Name) {
			results = append(results, LintResult{
				File:    path,
				Line:    svc.Line,
				Column:  1,
				Rule:    "SERVICE_PASCAL_CASE",
				Message: fmt.Sprintf("Service name %q should be PascalCase", svc.Name),
			})
		}
	}

	return results
}

func checkEnumValuePrefix(file *ProtoFile, path string) []LintResult {
	var results []LintResult

	for _, enum := range file.Enums {
		prefix := toUpperSnakeCase(enum.Name)

		for _, value := range enum.Values {
			if !strings.HasPrefix(value.Name, prefix+"_") {
				results = append(results, LintResult{
					File:    path,
					Line:    value.Line,
					Column:  1,
					Rule:    "ENUM_VALUE_PREFIX",
					Message: fmt.Sprintf("Enum value %q should be prefixed with %q", value.Name, prefix+"_"),
				})
			}
		}
	}

	return results
}

func checkEnumZeroValueSuffix(file *ProtoFile, path string) []LintResult {
	var results []LintResult

	for _, enum := range file.Enums {
		if len(enum.Values) > 0 {
			zeroValue := enum.Values[0]
			if zeroValue.Number == 0 && !strings.HasSuffix(zeroValue.Name, "_UNSPECIFIED") {
				results = append(results, LintResult{
					File:    path,
					Line:    zeroValue.Line,
					Column:  1,
					Rule:    "ENUM_ZERO_VALUE_SUFFIX",
					Message: fmt.Sprintf("Zero value %q should end with _UNSPECIFIED", zeroValue.Name),
				})
			}
		}
	}

	return results
}

func checkFileLowerSnakeCase(_ *ProtoFile, path string) []LintResult {
	filename := filepath.Base(path)
	name := strings.TrimSuffix(filename, ".proto")

	if !isLowerSnakeCase(name) {
		return []LintResult{{
			File:    path,
			Line:    1,
			Column:  1,
			Rule:    "FILE_LOWER_SNAKE_CASE",
			Message: fmt.Sprintf("File name %q should be lower_snake_case", filename),
		}}
	}

	return nil
}

func checkRPCRequestResponseUnique(_ *ProtoFile, _ string) []LintResult {
	// This requires cross-service checking
	return nil
}

func checkRPCRequestStandardName(file *ProtoFile, path string) []LintResult {
	var results []LintResult

	for _, svc := range file.Services {
		for _, method := range svc.Methods {
			expected := method.Name + "Request"
			// Strip package prefix if present
			inputType := method.InputType
			if idx := strings.LastIndex(inputType, "."); idx >= 0 {
				inputType = inputType[idx+1:]
			}

			if inputType != expected {
				results = append(results, LintResult{
					File:    path,
					Line:    method.Line,
					Column:  1,
					Rule:    "RPC_REQUEST_STANDARD_NAME",
					Message: fmt.Sprintf("RPC %q request type should be named %q", method.Name, expected),
				})
			}
		}
	}

	return results
}

func checkRPCResponseStandardName(file *ProtoFile, path string) []LintResult {
	var results []LintResult

	for _, svc := range file.Services {
		for _, method := range svc.Methods {
			expected := method.Name + "Response"
			outputType := method.OutputType

			if idx := strings.LastIndex(outputType, "."); idx >= 0 {
				outputType = outputType[idx+1:]
			}

			if outputType != expected {
				results = append(results, LintResult{
					File:    path,
					Line:    method.Line,
					Column:  1,
					Rule:    "RPC_RESPONSE_STANDARD_NAME",
					Message: fmt.Sprintf("RPC %q response type should be named %q", method.Name, expected),
				})
			}
		}
	}

	return results
}

func checkPackageVersionSuffix(file *ProtoFile, path string) []LintResult {
	if file.Package == "" {
		return nil
	}

	versionRegex := regexp.MustCompile(`v\d+$`)
	parts := strings.Split(file.Package, ".")
	lastPart := parts[len(parts)-1]

	if !versionRegex.MatchString(lastPart) {
		return []LintResult{{
			File:    path,
			Line:    1,
			Column:  1,
			Rule:    "PACKAGE_VERSION_SUFFIX",
			Message: fmt.Sprintf("Package %q should end with a version like v1", file.Package),
		}}
	}

	return nil
}

func checkServiceSuffix(file *ProtoFile, path string) []LintResult {
	var results []LintResult

	for _, svc := range file.Services {
		if !strings.HasSuffix(svc.Name, "Service") {
			results = append(results, LintResult{
				File:    path,
				Line:    svc.Line,
				Column:  1,
				Rule:    "SERVICE_SUFFIX",
				Message: fmt.Sprintf("Service name %q should end with Service", svc.Name),
			})
		}
	}

	return results
}

// Comment rules - simplified implementations
func checkCommentEnum(_ *ProtoFile, _ string) []LintResult      { return nil }
func checkCommentEnumValue(_ *ProtoFile, _ string) []LintResult { return nil }
func checkCommentField(_ *ProtoFile, _ string) []LintResult     { return nil }
func checkCommentMessage(_ *ProtoFile, _ string) []LintResult   { return nil }
func checkCommentRPC(_ *ProtoFile, _ string) []LintResult       { return nil }
func checkCommentService(_ *ProtoFile, _ string) []LintResult   { return nil }

// Helper functions

func isPascalCase(s string) bool {
	if s == "" {
		return false
	}

	if !unicode.IsUpper(rune(s[0])) {
		return false
	}

	for i, r := range s {
		if i == 0 {
			continue
		}

		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}

	return true
}

func isLowerSnakeCase(s string) bool {
	if s == "" {
		return false
	}

	for i, r := range s {
		if unicode.IsUpper(r) {
			return false
		}

		if r == '_' {
			if i == 0 || i == len(s)-1 {
				return false
			}
		} else if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}

	return true
}

func isUpperSnakeCase(s string) bool {
	if s == "" {
		return false
	}

	for i, r := range s {
		if unicode.IsLower(r) {
			return false
		}

		if r == '_' {
			if i == 0 || i == len(s)-1 {
				return false
			}
		} else if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}

	return true
}

func isLowerCase(s string) bool {
	for _, r := range s {
		if unicode.IsUpper(r) {
			return false
		}
	}

	return true
}

func toUpperSnakeCase(s string) string {
	var result strings.Builder

	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			result.WriteRune('_')
		}

		result.WriteRune(unicode.ToUpper(r))
	}

	return result.String()
}
