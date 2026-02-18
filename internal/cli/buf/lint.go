package buf

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/bufbuild/protocompile/ast"
	"github.com/bufbuild/protocompile/parser"
	"github.com/bufbuild/protocompile/reporter"
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

// RunLint runs lint on proto files using the built-in rule engine.
// The real buf lint engine (pkg/buf/pkg/bufapi) was removed;
// this uses the simplified linter with 28 built-in rules.
func RunLint(w io.Writer, dir string, opts LintOptions) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		absDir = dir
	}

	files, err := FindProtoFiles(absDir, nil)
	if err != nil {
		return fmt.Errorf("buf: %w", err)
	}

	if len(files) == 0 {
		_, _ = fmt.Fprintln(w, "No proto files found")
		return nil
	}

	config := LintConfig{Use: []string{CategoryStandard}}
	activeRules := getActiveRules(config)

	var totalIssues int
	for _, file := range files {
		protoFile, parseErr := parseProtoFileFromPath(file)
		if parseErr != nil {
			_, _ = fmt.Fprintf(w, "%s: parse error: %v\n", file, parseErr)
			continue
		}

		for _, rule := range activeRules {
			if shouldIgnoreRule(config, rule.ID, file) {
				continue
			}
			results := rule.Check(protoFile, file)
			for _, r := range results {
				_, _ = fmt.Fprintf(w, "%s:%d:%d: %s (%s)\n", r.File, r.Line, r.Column, r.Message, r.Rule)
				totalIssues++
			}
		}
	}

	if totalIssues > 0 {
		return fmt.Errorf("buf: lint found %d issue(s)", totalIssues)
	}
	return nil
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

// parseProtoFileFromPath reads a proto file and parses it using protocompile.
func parseProtoFileFromPath(path string) (*ProtoFile, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	handler := reporter.NewHandler(nil)
	fileNode, err := parser.Parse(filepath.Base(path), strings.NewReader(string(content)), handler)
	if err != nil || fileNode == nil {
		// Fallback to hand-written parser
		return ParseProtoFile(string(content))
	}

	return extractProtoFile(fileNode), nil
}

// extractProtoFile converts a protocompile AST into a ProtoFile.
func extractProtoFile(file *ast.FileNode) *ProtoFile {
	pf := &ProtoFile{}

	if syn := file.Syntax; syn != nil {
		pf.Syntax = syn.Syntax.AsString()
	}

	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.PackageNode:
			pf.Package = string(d.Name.AsIdentifier())
		case *ast.ImportNode:
			pi := ProtoImport{
				Path: d.Name.AsString(),
				Line: file.NodeInfo(d).Start().Line,
			}
			if d.Public != nil {
				pi.Public = true
			}
			if d.Weak != nil {
				pi.Weak = true
			}
			pf.Imports = append(pf.Imports, pi)
		case *ast.MessageNode:
			pf.Messages = append(pf.Messages, extractMessage(file, d))
		case *ast.EnumNode:
			pf.Enums = append(pf.Enums, extractEnum(file, d))
		case *ast.ServiceNode:
			pf.Services = append(pf.Services, extractService(file, d))
		case *ast.OptionNode:
			pf.Options = append(pf.Options, ProtoOption{
				Name:  extractOptionName(d),
				Value: extractOptionValue(d),
				Line:  file.NodeInfo(d).Start().Line,
			})
		}
	}

	return pf
}

func extractMessage(file *ast.FileNode, msg *ast.MessageNode) ProtoMessage {
	pm := ProtoMessage{
		Name: string(msg.Name.AsIdentifier()),
		Line: file.NodeInfo(msg).Start().Line,
	}

	for _, decl := range msg.Decls {
		switch d := decl.(type) {
		case *ast.FieldNode:
			pf := ProtoField{
				Name:   string(d.Name.AsIdentifier()),
				Type:   extractFieldType(d),
				Number: extractTagNumber(d.Tag),
				Line:   file.NodeInfo(d).Start().Line,
			}
			if d.Label.KeywordNode != nil {
				pf.Label = d.Label.KeywordNode.Val
			}
			pm.Fields = append(pm.Fields, pf)
		case *ast.MapFieldNode:
			pm.Fields = append(pm.Fields, ProtoField{
				Name:   string(d.Name.AsIdentifier()),
				Type:   "map",
				Number: extractTagNumber(d.Tag),
				Line:   file.NodeInfo(d).Start().Line,
			})
		case *ast.MessageNode:
			pm.Nested = append(pm.Nested, extractMessage(file, d))
		case *ast.EnumNode:
			pm.Enums = append(pm.Enums, extractEnum(file, d))
		case *ast.OptionNode:
			pm.Options = append(pm.Options, ProtoOption{
				Name:  extractOptionName(d),
				Value: extractOptionValue(d),
				Line:  file.NodeInfo(d).Start().Line,
			})
		case *ast.OneofNode:
			for _, odecl := range d.Decls {
				if f, ok := odecl.(*ast.FieldNode); ok {
					pm.Fields = append(pm.Fields, ProtoField{
						Name:   string(f.Name.AsIdentifier()),
						Type:   extractFieldType(f),
						Number: extractTagNumber(f.Tag),
						Label:  "oneof",
						Line:   file.NodeInfo(f).Start().Line,
					})
				}
			}
		}
	}

	return pm
}

func extractEnum(file *ast.FileNode, enum *ast.EnumNode) ProtoEnum {
	pe := ProtoEnum{
		Name: string(enum.Name.AsIdentifier()),
		Line: file.NodeInfo(enum).Start().Line,
	}

	for _, decl := range enum.Decls {
		switch d := decl.(type) {
		case *ast.EnumValueNode:
			num, _ := d.Number.AsInt64()
			pe.Values = append(pe.Values, ProtoEnumValue{
				Name:   string(d.Name.AsIdentifier()),
				Number: int(num),
				Line:   file.NodeInfo(d).Start().Line,
			})
		case *ast.OptionNode:
			pe.Options = append(pe.Options, ProtoOption{
				Name:  extractOptionName(d),
				Value: extractOptionValue(d),
				Line:  file.NodeInfo(d).Start().Line,
			})
		}
	}

	return pe
}

func extractService(file *ast.FileNode, svc *ast.ServiceNode) ProtoService {
	ps := ProtoService{
		Name: string(svc.Name.AsIdentifier()),
		Line: file.NodeInfo(svc).Start().Line,
	}

	for _, decl := range svc.Decls {
		switch d := decl.(type) {
		case *ast.RPCNode:
			method := ProtoMethod{
				Name:       string(d.Name.AsIdentifier()),
				InputType:  extractRPCType(d.Input),
				OutputType: extractRPCType(d.Output),
				Line:       file.NodeInfo(d).Start().Line,
			}
			if d.Input.Stream != nil {
				method.ClientStreaming = true
			}
			if d.Output.Stream != nil {
				method.ServerStreaming = true
			}
			ps.Methods = append(ps.Methods, method)
		case *ast.OptionNode:
			ps.Options = append(ps.Options, ProtoOption{
				Name:  extractOptionName(d),
				Value: extractOptionValue(d),
				Line:  file.NodeInfo(d).Start().Line,
			})
		}
	}

	return ps
}

func extractFieldType(f *ast.FieldNode) string {
	if f.FldType != nil {
		return string(f.FldType.AsIdentifier())
	}
	return ""
}

func extractRPCType(t *ast.RPCTypeNode) string {
	if t.MessageType != nil {
		return string(t.MessageType.AsIdentifier())
	}
	return ""
}

func extractTagNumber(tag *ast.UintLiteralNode) int {
	if tag == nil {
		return 0
	}
	return int(tag.Val)
}

func extractOptionName(opt *ast.OptionNode) string {
	if opt.Name != nil && len(opt.Name.Parts) > 0 {
		return string(opt.Name.Parts[0].Name.AsIdentifier())
	}
	return ""
}

func extractOptionValue(opt *ast.OptionNode) string {
	if opt.Val == nil {
		return ""
	}
	// Use the Value() interface which returns the Go representation
	v := opt.Val.Value()
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
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
