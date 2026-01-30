package buf

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ProtoImage represents a compiled protobuf image
type ProtoImage struct {
	Files []ProtoImageFile `json:"files"`
}

// ProtoImageFile represents a file in the image
type ProtoImageFile struct {
	Name     string            `json:"name"`
	Package  string            `json:"package"`
	Syntax   string            `json:"syntax"`
	Messages []string          `json:"messages,omitempty"`
	Enums    []string          `json:"enums,omitempty"`
	Services []string          `json:"services,omitempty"`
	Imports  []string          `json:"imports,omitempty"`
	Options  map[string]string `json:"options,omitempty"`
}

// RunBuild builds proto files
func RunBuild(w io.Writer, dir string, opts BuildOptions) error {
	// Find proto files
	files, err := FindProtoFiles(dir, opts.ExcludePath)
	if err != nil {
		return fmt.Errorf("buf: %w", err)
	}

	if len(files) == 0 {
		_, _ = fmt.Fprintln(w, "No proto files found")

		return nil
	}

	image := &ProtoImage{}

	var errors []string

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", file, err))

			continue
		}

		protoFile, err := ParseProtoFile(string(content))
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: parse error: %v", file, err))

			continue
		}

		relPath, _ := filepath.Rel(dir, file)
		relPath = strings.ReplaceAll(relPath, "\\", "/")

		imageFile := ProtoImageFile{
			Name:    relPath,
			Package: protoFile.Package,
			Syntax:  protoFile.Syntax,
			Options: make(map[string]string),
		}

		// Add messages
		for _, msg := range protoFile.Messages {
			imageFile.Messages = append(imageFile.Messages, msg.Name)
		}

		// Add enums
		for _, enum := range protoFile.Enums {
			imageFile.Enums = append(imageFile.Enums, enum.Name)
		}

		// Add services
		for _, svc := range protoFile.Services {
			imageFile.Services = append(imageFile.Services, svc.Name)
		}

		// Add imports
		for _, imp := range protoFile.Imports {
			imageFile.Imports = append(imageFile.Imports, imp.Path)
		}

		// Add options
		for _, opt := range protoFile.Options {
			imageFile.Options[opt.Name] = opt.Value
		}

		image.Files = append(image.Files, imageFile)
	}

	// Sort files by name
	sort.Slice(image.Files, func(i, j int) bool {
		return image.Files[i].Name < image.Files[j].Name
	})

	if len(errors) > 0 {
		for _, e := range errors {
			_, _ = fmt.Fprintf(w, "error: %s\n", e)
		}

		return fmt.Errorf("build failed with %d error(s)", len(errors))
	}

	// Output result
	if opts.Output != "" {
		var (
			output     []byte
			marshalErr error
		)

		if strings.HasSuffix(opts.Output, ".json") {
			output, marshalErr = json.MarshalIndent(image, "", "  ")
		} else {
			// Binary format (simplified as JSON for now)
			output, marshalErr = json.Marshal(image)
		}

		if marshalErr != nil {
			return fmt.Errorf("buf: failed to marshal output: %w", marshalErr)
		}

		if err := os.WriteFile(opts.Output, output, 0644); err != nil {
			return fmt.Errorf("buf: failed to write output: %w", err)
		}

		_, _ = fmt.Fprintf(w, "Built %d file(s) to %s\n", len(image.Files), opts.Output)
	} else {
		_, _ = fmt.Fprintf(w, "Built %d file(s)\n", len(image.Files))

		for _, f := range image.Files {
			_, _ = fmt.Fprintf(w, "  %s\n", f.Name)

			if len(f.Messages) > 0 {
				_, _ = fmt.Fprintf(w, "    messages: %s\n", strings.Join(f.Messages, ", "))
			}

			if len(f.Enums) > 0 {
				_, _ = fmt.Fprintf(w, "    enums: %s\n", strings.Join(f.Enums, ", "))
			}

			if len(f.Services) > 0 {
				_, _ = fmt.Fprintf(w, "    services: %s\n", strings.Join(f.Services, ", "))
			}
		}
	}

	return nil
}

// RunBreaking checks for breaking changes
func RunBreaking(w io.Writer, dir string, opts BreakingOptions) error {
	if opts.Against == "" {
		return fmt.Errorf("buf: --against is required")
	}

	// Find current proto files
	currentFiles, err := FindProtoFiles(dir, opts.ExcludePath)
	if err != nil {
		return fmt.Errorf("buf: %w", err)
	}

	// Parse current files
	currentSchemas := make(map[string]*ProtoFile)

	for _, file := range currentFiles {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		protoFile, err := ParseProtoFile(string(content))
		if err != nil {
			continue
		}

		relPath, _ := filepath.Rel(dir, file)
		relPath = strings.ReplaceAll(relPath, "\\", "/")
		currentSchemas[relPath] = protoFile
	}

	// Find against proto files
	againstFiles, err := FindProtoFiles(opts.Against, opts.ExcludePath)
	if err != nil {
		return fmt.Errorf("buf: failed to read against directory: %w", err)
	}

	// Parse against files
	againstSchemas := make(map[string]*ProtoFile)

	for _, file := range againstFiles {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		protoFile, err := ParseProtoFile(string(content))
		if err != nil {
			continue
		}

		relPath, _ := filepath.Rel(opts.Against, file)
		relPath = strings.ReplaceAll(relPath, "\\", "/")
		againstSchemas[relPath] = protoFile
	}

	// Check for breaking changes
	var results []LintResult

	// Check for removed files
	for path := range againstSchemas {
		if _, ok := currentSchemas[path]; !ok {
			results = append(results, LintResult{
				File:    path,
				Line:    1,
				Column:  1,
				Rule:    "FILE_NO_DELETE",
				Message: "File was deleted",
			})
		}
	}

	// Check each file for breaking changes
	for path, current := range currentSchemas {
		against, ok := againstSchemas[path]
		if !ok {
			continue
		}

		// Check package changes
		if current.Package != against.Package {
			results = append(results, LintResult{
				File:    path,
				Line:    1,
				Column:  1,
				Rule:    "PACKAGE_NO_DELETE",
				Message: fmt.Sprintf("Package changed from %q to %q", against.Package, current.Package),
			})
		}

		// Check for deleted messages
		currentMsgs := make(map[string]ProtoMessage)
		for _, msg := range current.Messages {
			currentMsgs[msg.Name] = msg
		}

		for _, msg := range against.Messages {
			if _, ok := currentMsgs[msg.Name]; !ok {
				results = append(results, LintResult{
					File:    path,
					Line:    msg.Line,
					Column:  1,
					Rule:    "MESSAGE_NO_DELETE",
					Message: fmt.Sprintf("Message %q was deleted", msg.Name),
				})
			}
		}

		// Check for deleted enums
		currentEnums := make(map[string]ProtoEnum)
		for _, enum := range current.Enums {
			currentEnums[enum.Name] = enum
		}

		for _, enum := range against.Enums {
			if _, ok := currentEnums[enum.Name]; !ok {
				results = append(results, LintResult{
					File:    path,
					Line:    enum.Line,
					Column:  1,
					Rule:    "ENUM_NO_DELETE",
					Message: fmt.Sprintf("Enum %q was deleted", enum.Name),
				})
			}
		}

		// Check for deleted services
		currentSvcs := make(map[string]ProtoService)
		for _, svc := range current.Services {
			currentSvcs[svc.Name] = svc
		}

		for _, svc := range against.Services {
			currentSvc, ok := currentSvcs[svc.Name]
			if !ok {
				results = append(results, LintResult{
					File:    path,
					Line:    svc.Line,
					Column:  1,
					Rule:    "SERVICE_NO_DELETE",
					Message: fmt.Sprintf("Service %q was deleted", svc.Name),
				})

				continue
			}

			// Check for deleted RPCs
			currentMethods := make(map[string]ProtoMethod)
			for _, method := range currentSvc.Methods {
				currentMethods[method.Name] = method
			}

			for _, method := range svc.Methods {
				if _, ok := currentMethods[method.Name]; !ok {
					results = append(results, LintResult{
						File:    path,
						Line:    method.Line,
						Column:  1,
						Rule:    "RPC_NO_DELETE",
						Message: fmt.Sprintf("RPC %q was deleted from service %q", method.Name, svc.Name),
					})
				}
			}
		}

		// Check for field changes in messages
		results = append(results, checkMessageFieldChanges(path, current.Messages, against.Messages)...)
	}

	if len(results) == 0 {
		_, _ = fmt.Fprintln(w, "No breaking changes detected")

		return nil
	}

	// Output results
	if err := OutputResults(w, results, opts.ErrorFormat); err != nil {
		return err
	}

	return fmt.Errorf("found %d breaking change(s)", len(results))
}

func checkMessageFieldChanges(path string, current, against []ProtoMessage) []LintResult {
	var results []LintResult

	currentMsgs := make(map[string]ProtoMessage)
	for _, msg := range current {
		currentMsgs[msg.Name] = msg
	}

	for _, againstMsg := range against {
		currentMsg, ok := currentMsgs[againstMsg.Name]
		if !ok {
			continue
		}

		// Build field maps
		currentFields := make(map[int]ProtoField)
		for _, field := range currentMsg.Fields {
			currentFields[field.Number] = field
		}

		for _, field := range againstMsg.Fields {
			currentField, ok := currentFields[field.Number]
			if !ok {
				results = append(results, LintResult{
					File:    path,
					Line:    field.Line,
					Column:  1,
					Rule:    "FIELD_NO_DELETE",
					Message: fmt.Sprintf("Field %q (number %d) was deleted from message %q", field.Name, field.Number, againstMsg.Name),
				})

				continue
			}

			// Check type changes
			if currentField.Type != field.Type {
				results = append(results, LintResult{
					File:    path,
					Line:    currentField.Line,
					Column:  1,
					Rule:    "FIELD_SAME_TYPE",
					Message: fmt.Sprintf("Field %q type changed from %q to %q", field.Name, field.Type, currentField.Type),
				})
			}

			// Check name changes (technically not breaking, but worth noting)
			if currentField.Name != field.Name {
				results = append(results, LintResult{
					File:    path,
					Line:    currentField.Line,
					Column:  1,
					Rule:    "FIELD_SAME_NAME",
					Message: fmt.Sprintf("Field name changed from %q to %q", field.Name, currentField.Name),
				})
			}
		}

		// Check nested messages
		results = append(results, checkMessageFieldChanges(path, currentMsg.Nested, againstMsg.Nested)...)
	}

	return results
}
