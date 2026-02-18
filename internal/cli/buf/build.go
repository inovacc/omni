package buf

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bufbuild/protocompile"
	"github.com/bufbuild/protocompile/linker"
	"github.com/bufbuild/protocompile/reporter"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/types/descriptorpb"
)

// RunBuild compiles proto files and produces a FileDescriptorSet.
func RunBuild(w io.Writer, dir string, opts BuildOptions) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		absDir = dir
	}

	files, err := FindProtoFiles(absDir, opts.ExcludePath)
	if err != nil {
		return fmt.Errorf("buf build: %w", err)
	}

	if len(files) == 0 {
		_, _ = fmt.Fprintln(w, "No proto files found")
		return nil
	}

	// Convert to relative paths for the compiler
	relFiles := make([]string, len(files))
	for i, f := range files {
		rel, relErr := filepath.Rel(absDir, f)
		if relErr != nil {
			rel = f
		}
		relFiles[i] = filepath.ToSlash(rel)
	}

	// Compile proto files
	compiled, compileErrs := compileProtos(absDir, relFiles)
	if compileErrs != nil {
		return fmt.Errorf("buf build: %w", compileErrs)
	}

	// Build FileDescriptorSet
	fds := buildFileDescriptorSet(compiled)

	if opts.Output != "" {
		if err := writeFileDescriptorSet(fds, opts.Output); err != nil {
			return fmt.Errorf("buf build: %w", err)
		}
		_, _ = fmt.Fprintf(w, "Built %d file(s) to %s\n", len(fds.File), opts.Output)
	} else {
		// Write JSON to stdout
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		if err := enc.Encode(descriptorSetToMap(fds)); err != nil {
			return fmt.Errorf("buf build: %w", err)
		}
	}

	return nil
}

// RunBreaking checks for breaking changes between dir and opts.Against.
func RunBreaking(w io.Writer, dir string, opts BreakingOptions) error {
	if opts.Against == "" {
		return fmt.Errorf("buf: --against is required")
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		absDir = dir
	}
	absAgainst, err := filepath.Abs(opts.Against)
	if err != nil {
		absAgainst = opts.Against
	}

	// Compile current
	currentFiles, err := FindProtoFiles(absDir, opts.ExcludePath)
	if err != nil {
		return fmt.Errorf("buf breaking: %w", err)
	}
	currentRel := toRelSlash(absDir, currentFiles)
	currentLinked, compileErr := compileProtos(absDir, currentRel)
	if compileErr != nil {
		return fmt.Errorf("buf breaking: current: %w", compileErr)
	}

	// Compile against
	againstFiles, err := FindProtoFiles(absAgainst, opts.ExcludePath)
	if err != nil {
		return fmt.Errorf("buf breaking: %w", err)
	}
	againstRel := toRelSlash(absAgainst, againstFiles)
	againstLinked, compileErr := compileProtos(absAgainst, againstRel)
	if compileErr != nil {
		return fmt.Errorf("buf breaking: against: %w", compileErr)
	}

	currentFDS := buildFileDescriptorSet(currentLinked)
	againstFDS := buildFileDescriptorSet(againstLinked)

	// Detect breaking changes
	issues := detectBreakingChanges(currentFDS, againstFDS, opts.ExcludeImports)

	if len(issues) == 0 {
		return nil
	}

	for _, issue := range issues {
		_, _ = fmt.Fprintf(w, "%s:%d:%d: %s (%s)\n",
			issue.File, issue.Line, issue.Column, issue.Message, issue.Rule)
	}

	return fmt.Errorf("buf breaking: found %d breaking change(s)", len(issues))
}

// compileProtos compiles proto files in the given directory using protocompile.
func compileProtos(dir string, relFiles []string) (linker.Files, error) {
	var errs []reporter.ErrorWithPos

	errHandler := func(err reporter.ErrorWithPos) error {
		errs = append(errs, err)
		return nil // collect all errors
	}

	rep := reporter.NewReporter(errHandler, nil)

	compiler := protocompile.Compiler{
		Resolver: protocompile.WithStandardImports(
			&protocompile.SourceResolver{
				ImportPaths: []string{dir},
			},
		),
		Reporter: rep,
	}

	linked, err := compiler.Compile(context.Background(), relFiles...)
	if err != nil {
		if len(errs) > 0 {
			var msgs []string
			for _, e := range errs {
				msgs = append(msgs, e.Error())
			}
			return nil, fmt.Errorf("compilation failed:\n%s", strings.Join(msgs, "\n"))
		}
		return nil, err
	}

	return linked, nil
}

// buildFileDescriptorSet converts compiled linker.Files to a FileDescriptorSet.
func buildFileDescriptorSet(linked linker.Files) *descriptorpb.FileDescriptorSet {
	fds := &descriptorpb.FileDescriptorSet{}
	seen := make(map[string]bool)

	for _, fd := range linked {
		addFileAndDeps(fds, protodesc.ToFileDescriptorProto(fd), seen)
	}

	return fds
}

// addFileAndDeps recursively adds a file descriptor and its dependencies.
func addFileAndDeps(fds *descriptorpb.FileDescriptorSet, fd *descriptorpb.FileDescriptorProto, seen map[string]bool) {
	name := fd.GetName()
	if seen[name] {
		return
	}
	seen[name] = true
	fds.File = append(fds.File, fd)
}

// writeFileDescriptorSet writes a FileDescriptorSet to a file.
func writeFileDescriptorSet(fds *descriptorpb.FileDescriptorSet, path string) error {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".json":
		data, err := json.MarshalIndent(descriptorSetToMap(fds), "", "  ")
		if err != nil {
			return fmt.Errorf("marshal json: %w", err)
		}
		return os.WriteFile(path, data, 0644)

	default: // .bin, .binpb, or any other extension → binary protobuf
		data, err := proto.Marshal(fds)
		if err != nil {
			return fmt.Errorf("marshal binary: %w", err)
		}
		return os.WriteFile(path, data, 0644)
	}
}

// descriptorSetToMap converts a FileDescriptorSet to a JSON-friendly map.
func descriptorSetToMap(fds *descriptorpb.FileDescriptorSet) map[string]interface{} {
	filesArr := make([]map[string]interface{}, 0, len(fds.File))

	for _, fd := range fds.File {
		fileMap := map[string]interface{}{
			"name": fd.GetName(),
		}
		if fd.Package != nil {
			fileMap["package"] = fd.GetPackage()
		}
		if fd.Syntax != nil {
			fileMap["syntax"] = fd.GetSyntax()
		}
		if len(fd.Dependency) > 0 {
			fileMap["dependency"] = fd.Dependency
		}

		if len(fd.MessageType) > 0 {
			msgs := make([]map[string]interface{}, 0, len(fd.MessageType))
			for _, msg := range fd.MessageType {
				msgs = append(msgs, messageToMap(msg))
			}
			fileMap["messageType"] = msgs
		}

		if len(fd.EnumType) > 0 {
			enums := make([]map[string]interface{}, 0, len(fd.EnumType))
			for _, enum := range fd.EnumType {
				enums = append(enums, enumToMap(enum))
			}
			fileMap["enumType"] = enums
		}

		if len(fd.Service) > 0 {
			svcs := make([]map[string]interface{}, 0, len(fd.Service))
			for _, svc := range fd.Service {
				svcs = append(svcs, serviceToMap(svc))
			}
			fileMap["service"] = svcs
		}

		filesArr = append(filesArr, fileMap)
	}

	return map[string]interface{}{
		"file": filesArr,
	}
}

func messageToMap(msg *descriptorpb.DescriptorProto) map[string]interface{} {
	m := map[string]interface{}{
		"name": msg.GetName(),
	}

	if len(msg.Field) > 0 {
		fields := make([]map[string]interface{}, 0, len(msg.Field))
		for _, f := range msg.Field {
			fm := map[string]interface{}{
				"name":   f.GetName(),
				"number": f.GetNumber(),
				"type":   f.GetType().String(),
			}
			if f.TypeName != nil {
				fm["typeName"] = f.GetTypeName()
			}
			if f.Label != nil {
				fm["label"] = f.GetLabel().String()
			}
			fields = append(fields, fm)
		}
		m["field"] = fields
	}

	if len(msg.NestedType) > 0 {
		nested := make([]map[string]interface{}, 0, len(msg.NestedType))
		for _, n := range msg.NestedType {
			nested = append(nested, messageToMap(n))
		}
		m["nestedType"] = nested
	}

	if len(msg.EnumType) > 0 {
		enums := make([]map[string]interface{}, 0, len(msg.EnumType))
		for _, e := range msg.EnumType {
			enums = append(enums, enumToMap(e))
		}
		m["enumType"] = enums
	}

	return m
}

func enumToMap(enum *descriptorpb.EnumDescriptorProto) map[string]interface{} {
	m := map[string]interface{}{
		"name": enum.GetName(),
	}

	if len(enum.Value) > 0 {
		values := make([]map[string]interface{}, 0, len(enum.Value))
		for _, v := range enum.Value {
			values = append(values, map[string]interface{}{
				"name":   v.GetName(),
				"number": v.GetNumber(),
			})
		}
		m["value"] = values
	}

	return m
}

func serviceToMap(svc *descriptorpb.ServiceDescriptorProto) map[string]interface{} {
	m := map[string]interface{}{
		"name": svc.GetName(),
	}

	if len(svc.Method) > 0 {
		methods := make([]map[string]interface{}, 0, len(svc.Method))
		for _, method := range svc.Method {
			mm := map[string]interface{}{
				"name":       method.GetName(),
				"inputType":  method.GetInputType(),
				"outputType": method.GetOutputType(),
			}
			methods = append(methods, mm)
		}
		m["method"] = methods
	}

	return m
}

// toRelSlash converts absolute paths to relative slash paths.
func toRelSlash(base string, paths []string) []string {
	rel := make([]string, 0, len(paths))
	for _, p := range paths {
		r, err := filepath.Rel(base, p)
		if err != nil {
			r = p
		}
		rel = append(rel, filepath.ToSlash(r))
	}
	return rel
}

// BreakingIssue represents a breaking change.
type BreakingIssue struct {
	File    string
	Line    int
	Column  int
	Rule    string
	Message string
}

// detectBreakingChanges compares two FileDescriptorSets for breaking changes.
func detectBreakingChanges(current, against *descriptorpb.FileDescriptorSet, excludeImports bool) []BreakingIssue {
	var issues []BreakingIssue

	// Index current files by name
	currentFiles := make(map[string]*descriptorpb.FileDescriptorProto)
	for _, f := range current.File {
		currentFiles[f.GetName()] = f
	}

	// Index against files by name
	againstFiles := make(map[string]*descriptorpb.FileDescriptorProto)
	for _, f := range against.File {
		againstFiles[f.GetName()] = f
	}

	// Check for deleted files
	for name, prevFile := range againstFiles {
		if excludeImports && isWellKnownType(name) {
			continue
		}
		if _, ok := currentFiles[name]; !ok {
			issues = append(issues, BreakingIssue{
				File:    name,
				Line:    1,
				Column:  1,
				Rule:    "FILE_NO_DELETE",
				Message: fmt.Sprintf("Previously present file %q was deleted.", name),
			})
			continue
		}
		curFile := currentFiles[name]

		// Check breaking changes within matched files
		issues = append(issues, checkFileBreaking(curFile, prevFile)...)
	}

	return issues
}

func checkFileBreaking(current, previous *descriptorpb.FileDescriptorProto) []BreakingIssue {
	var issues []BreakingIssue
	fileName := current.GetName()

	// PACKAGE_NO_DELETE
	if previous.GetPackage() != "" && current.GetPackage() != previous.GetPackage() {
		issues = append(issues, BreakingIssue{
			File:    fileName,
			Line:    1,
			Column:  1,
			Rule:    "PACKAGE_NO_DELETE",
			Message: fmt.Sprintf("Package %q was changed to %q.", previous.GetPackage(), current.GetPackage()),
		})
	}

	// Index previous messages
	prevMsgs := make(map[string]*descriptorpb.DescriptorProto)
	for _, m := range previous.MessageType {
		prevMsgs[m.GetName()] = m
	}

	// MESSAGE_NO_DELETE
	for name, prevMsg := range prevMsgs {
		found := false
		for _, m := range current.MessageType {
			if m.GetName() == name {
				found = true
				issues = append(issues, checkMessageBreaking(fileName, m, prevMsg)...)
				break
			}
		}
		if !found {
			issues = append(issues, BreakingIssue{
				File:    fileName,
				Line:    1,
				Column:  1,
				Rule:    "MESSAGE_NO_DELETE",
				Message: fmt.Sprintf("Previously present message %q was deleted.", name),
			})
		}
	}

	// ENUM_NO_DELETE
	prevEnums := make(map[string]*descriptorpb.EnumDescriptorProto)
	for _, e := range previous.EnumType {
		prevEnums[e.GetName()] = e
	}
	for name := range prevEnums {
		found := false
		for _, e := range current.EnumType {
			if e.GetName() == name {
				found = true
				break
			}
		}
		if !found {
			issues = append(issues, BreakingIssue{
				File:    fileName,
				Line:    1,
				Column:  1,
				Rule:    "ENUM_NO_DELETE",
				Message: fmt.Sprintf("Previously present enum %q was deleted.", name),
			})
		}
	}

	// SERVICE_NO_DELETE
	prevSvcs := make(map[string]*descriptorpb.ServiceDescriptorProto)
	for _, s := range previous.Service {
		prevSvcs[s.GetName()] = s
	}
	for name, prevSvc := range prevSvcs {
		found := false
		for _, s := range current.Service {
			if s.GetName() == name {
				found = true
				issues = append(issues, checkServiceBreaking(fileName, s, prevSvc)...)
				break
			}
		}
		if !found {
			issues = append(issues, BreakingIssue{
				File:    fileName,
				Line:    1,
				Column:  1,
				Rule:    "SERVICE_NO_DELETE",
				Message: fmt.Sprintf("Previously present service %q was deleted.", name),
			})
		}
	}

	return issues
}

func checkMessageBreaking(file string, current, previous *descriptorpb.DescriptorProto) []BreakingIssue {
	var issues []BreakingIssue

	// Index previous fields by number
	prevFields := make(map[int32]*descriptorpb.FieldDescriptorProto)
	for _, f := range previous.Field {
		prevFields[f.GetNumber()] = f
	}

	// FIELD_NO_DELETE + FIELD_SAME_TYPE
	for num, prevField := range prevFields {
		found := false
		for _, f := range current.Field {
			if f.GetNumber() == num {
				found = true
				if f.GetType() != prevField.GetType() {
					issues = append(issues, BreakingIssue{
						File:    file,
						Line:    1,
						Column:  1,
						Rule:    "FIELD_SAME_TYPE",
						Message: fmt.Sprintf("Field %d on message %q changed type from %q to %q.", num, current.GetName(), prevField.GetType(), f.GetType()),
					})
				}
				break
			}
		}
		if !found {
			issues = append(issues, BreakingIssue{
				File:    file,
				Line:    1,
				Column:  1,
				Rule:    "FIELD_NO_DELETE",
				Message: fmt.Sprintf("Previously present field %q (number %d) on message %q was deleted.", prevField.GetName(), num, current.GetName()),
			})
		}
	}

	return issues
}

func checkServiceBreaking(file string, current, previous *descriptorpb.ServiceDescriptorProto) []BreakingIssue {
	var issues []BreakingIssue

	prevMethods := make(map[string]*descriptorpb.MethodDescriptorProto)
	for _, m := range previous.Method {
		prevMethods[m.GetName()] = m
	}

	// RPC_NO_DELETE
	for name := range prevMethods {
		found := false
		for _, m := range current.Method {
			if m.GetName() == name {
				found = true
				break
			}
		}
		if !found {
			issues = append(issues, BreakingIssue{
				File:    file,
				Line:    1,
				Column:  1,
				Rule:    "RPC_NO_DELETE",
				Message: fmt.Sprintf("Previously present RPC %q on service %q was deleted.", name, current.GetName()),
			})
		}
	}

	return issues
}

func isWellKnownType(path string) bool {
	return strings.HasPrefix(path, "google/protobuf/")
}
