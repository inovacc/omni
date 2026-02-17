// Copyright 2020-2025 Buf Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package bufapi provides a public API for buf operations usable from outside
// the github.com/bufbuild/buf module. It wraps internal packages behind a
// simple interface using only stdlib types.
package bufapi

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"

	"github.com/inovacc/omni/pkg/buf/internal/buf/bufanalysis"
	"github.com/inovacc/omni/pkg/buf/internal/buf/bufcheck"
	"github.com/inovacc/omni/pkg/buf/internal/buf/bufconfig"
	"github.com/inovacc/omni/pkg/buf/internal/buf/bufformat"
	"github.com/inovacc/omni/pkg/buf/internal/buf/bufimage"
	"github.com/inovacc/omni/pkg/buf/internal/buf/bufmodule"
	"github.com/inovacc/omni/pkg/buf/internal/protocompile/parser"
	"github.com/inovacc/omni/pkg/buf/internal/protocompile/reporter"
	"github.com/inovacc/omni/pkg/buf/pkg/standard/xlog/xslog"
	"github.com/inovacc/omni/pkg/buf/pkg/protoencoding"
	"github.com/inovacc/omni/pkg/buf/pkg/storage"
	"github.com/inovacc/omni/pkg/buf/pkg/storage/storageos"
)

// FormatProto formats a single .proto file's source content.
// filename is used only for error messages.
// Returns the formatted source as a string.
func FormatProto(filename string, source string) (string, error) {
	fileNode, err := parser.Parse(filename, strings.NewReader(source), reporter.NewHandler(nil))
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := bufformat.FormatFileNode(&buf, fileNode); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// FormatProtoReader formats a single .proto file from a reader.
// filename is used only for error messages.
// Writes the formatted output to w.
func FormatProtoReader(w io.Writer, filename string, r io.Reader) error {
	fileNode, err := parser.Parse(filename, r, reporter.NewHandler(nil))
	if err != nil {
		return err
	}
	return bufformat.FormatFileNode(w, fileNode)
}

// LintDir runs buf lint on a directory of .proto files.
// It reads buf.yaml for configuration if present, falling back to default v1 lint rules.
// Results are written to w in the given format ("text", "json", "github-actions", "msvs", "junit").
// Returns nil if no lint violations, error otherwise.
func LintDir(ctx context.Context, w io.Writer, dir string, errorFormat string) error {
	logger := xslog.NopLogger

	bucket, image, err := buildImageFromDir(ctx, logger, dir)
	if err != nil {
		return handleAnnotationError(w, err, errorFormat)
	}

	// Load lint config from buf.yaml, fall back to default v1.
	lintConfig := bufconfig.DefaultLintConfigV1
	if bufYAML, yamlErr := bufconfig.GetBufYAMLFileForPrefix(ctx, bucket, "."); yamlErr == nil {
		if cfgs := bufYAML.ModuleConfigs(); len(cfgs) > 0 {
			lintConfig = cfgs[0].LintConfig()
		}
	}

	checkClient, err := bufcheck.NewClient(logger)
	if err != nil {
		return err
	}

	if err := checkClient.Lint(ctx, lintConfig, image); err != nil {
		return handleAnnotationError(w, err, errorFormat)
	}
	return nil
}

// BreakingDir checks for breaking changes between dir (current) and againstDir (previous).
// It reads buf.yaml for configuration if present, falling back to default v1 breaking rules.
// Results are written to w in the given format.
// Returns nil if no breaking changes, error otherwise.
func BreakingDir(ctx context.Context, w io.Writer, dir string, againstDir string, errorFormat string) error {
	logger := xslog.NopLogger

	bucket, image, err := buildImageFromDir(ctx, logger, dir)
	if err != nil {
		return handleAnnotationError(w, err, errorFormat)
	}

	_, againstImage, err := buildImageFromDir(ctx, logger, againstDir)
	if err != nil {
		return handleAnnotationError(w, err, errorFormat)
	}

	// Load breaking config from buf.yaml, fall back to default v1.
	breakingConfig := bufconfig.DefaultBreakingConfigV1
	if bufYAML, yamlErr := bufconfig.GetBufYAMLFileForPrefix(ctx, bucket, "."); yamlErr == nil {
		if cfgs := bufYAML.ModuleConfigs(); len(cfgs) > 0 {
			breakingConfig = cfgs[0].BreakingConfig()
		}
	}

	checkClient, err := bufcheck.NewClient(logger)
	if err != nil {
		return err
	}

	if err := checkClient.Breaking(ctx, breakingConfig, image, againstImage); err != nil {
		return handleAnnotationError(w, err, errorFormat)
	}
	return nil
}

// BuildDir compiles proto files in dir and writes the image to w.
// outputFormat is "json" or "bin" (wire/binary). If empty, defaults to "json".
// Returns the number of files compiled, or an error.
func BuildDir(ctx context.Context, w io.Writer, dir string, outputFormat string) (int, error) {
	logger := xslog.NopLogger

	_, image, err := buildImageFromDir(ctx, logger, dir)
	if err != nil {
		return 0, err
	}

	fileCount := len(image.Files())
	if fileCount == 0 {
		return 0, nil
	}

	fds := bufimage.ImageToFileDescriptorSet(image)

	switch outputFormat {
	case "bin", "binary":
		marshaler := protoencoding.NewWireMarshaler()
		data, err := marshaler.Marshal(fds)
		if err != nil {
			return 0, err
		}
		_, err = w.Write(data)
		return fileCount, err
	default: // json
		marshaler := protoencoding.NewJSONMarshaler(nil, protoencoding.JSONMarshalerWithIndent())
		data, err := marshaler.Marshal(fds)
		if err != nil {
			return 0, err
		}
		_, err = w.Write(data)
		return fileCount, err
	}
}

// buildImageFromDir opens a directory, builds a ModuleSet, and compiles an Image.
func buildImageFromDir(ctx context.Context, logger *slog.Logger, dir string) (storage.ReadBucket, bufimage.Image, error) {
	provider := storageos.NewProvider(storageos.ProviderWithSymlinks())
	bucket, err := provider.NewReadWriteBucket(dir, storageos.ReadWriteBucketWithSymlinksIfSupported())
	if err != nil {
		return nil, nil, err
	}

	moduleSetBuilder := bufmodule.NewModuleSetBuilder(ctx, logger, bufmodule.NopModuleDataProvider, bufmodule.NopCommitProvider)
	moduleSetBuilder.AddLocalModule(bucket, dir, true)
	moduleSet, err := moduleSetBuilder.Build()
	if err != nil {
		return nil, nil, err
	}

	moduleReadBucket := bufmodule.ModuleSetToModuleReadBucketWithOnlyProtoFiles(moduleSet)
	image, err := bufimage.BuildImage(ctx, logger, moduleReadBucket)
	if err != nil {
		return nil, nil, err
	}
	return bucket, image, nil
}

// ErrViolationsFound is returned when lint or breaking violations are detected and printed.
var ErrViolationsFound = errors.New("violations found")

// handleAnnotationError checks if err is a FileAnnotationSet, prints it, and
// returns ErrViolationsFound. Otherwise returns the error as-is.
func handleAnnotationError(w io.Writer, err error, errorFormat string) error {
	var fas bufanalysis.FileAnnotationSet
	if errors.As(err, &fas) {
		if printErr := bufanalysis.PrintFileAnnotationSet(w, fas, errorFormat); printErr != nil {
			return printErr
		}
		return ErrViolationsFound
	}
	return err
}
