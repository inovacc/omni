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

package bufctl

import (
	"github.com/inovacc/omni/pkg/buf/internal/buf/buffetch"
)

// ControllerOption is a controller option.
type ControllerOption func(*controller)

// WithDisableSymlinks returns a new ControllerOption that
func WithDisableSymlinks(disableSymlinks bool) ControllerOption {
	return func(controller *controller) {
		controller.disableSymlinks = disableSymlinks
	}
}

// WithFileAnnotationErrorFormat returns a new ControllerOption that sets the FileAnnotation format.
//
// The flagName is the CLI flag name used to set this format (e.g. "error-format"),
// used in error messages when the format is invalid.
func WithFileAnnotationErrorFormat(fileAnnotationErrorFormat string, flagName string) ControllerOption {
	return func(controller *controller) {
		controller.fileAnnotationErrorFormat = fileAnnotationErrorFormat
		if flagName != "" {
			controller.fileAnnotationErrorFormatFlagName = flagName
		}
	}
}

// WithFileAnnotationsToStdout returns a new ControllerOption that sends FileAnnotations to stdout
func WithFileAnnotationsToStdout() ControllerOption {
	return func(controller *controller) {
		controller.fileAnnotationsToStdout = true
	}
}

// WithCopyToInMemory returns a new ControllerOption that copies to memory.
func WithCopyToInMemory() ControllerOption {
	return func(controller *controller) {
		controller.copyToInMemory = true
	}
}

// WorkspaceOption is an option for workspace operations (GetWorkspace, GetWorkspaceDepManager).
type WorkspaceOption func(*workspaceOptions)

// WithTargetPaths returns a new WorkspaceOption that sets the target paths and their
// corresponding CLI flag names for error messages.
func WithTargetPaths(targetPaths []string, targetExcludePaths []string, pathFlagName string, excludePathFlagName string) WorkspaceOption {
	return func(workspaceOptions *workspaceOptions) {
		workspaceOptions.targetPaths = targetPaths
		workspaceOptions.targetExcludePaths = targetExcludePaths
		if pathFlagName != "" {
			workspaceOptions.pathFlagName = pathFlagName
		}
		if excludePathFlagName != "" {
			workspaceOptions.excludePathFlagName = excludePathFlagName
		}
	}
}

// WithConfigOverride applies the config override for workspace operations.
//
// This flag will only work if no buf.work.yaml is detected, and the buf.yaml is a
// v1beta1 buf.yaml, v1 buf.yaml, or no buf.yaml. This flag will not work if a buf.work.yaml
// is detected, or a v2 buf.yaml is detected.
//
// If used with an image or module ref, this has no effect on the build, i.e. excludes are
// not respected, and the module name is ignored. This matches old behavior.
//
// This implements the soon-to-be-deprecated --config flag.
//
// See bufconfig.GetBufYAMLFileForPrefixOrOverride for more details.
//
// *** DO NOT USE THIS OUTSIDE OF THE CLI AND/OR IF YOU DON'T UNDERSTAND IT. ***
// *** DO NOT ADD THIS TO ANY NEW COMMANDS. ***
//
// Current commands that use this: build, breaking, lint, generate, format,
// export, ls-breaking-rules, ls-lint-rules.
func WithConfigOverride(configOverride string) WorkspaceOption {
	return func(workspaceOptions *workspaceOptions) {
		workspaceOptions.configOverride = configOverride
	}
}

// WithIgnoreAndDisallowV1BufWorkYAMLs returns a new WorkspaceOption that says
// to ignore dependencies from buf.work.yamls at the root of the bucket, and to also
// disallow directories with buf.work.yamls to be directly targeted.
//
// See bufworkspace.WithIgnoreAndDisallowV1BufWorkYAMLs for more details.
func WithIgnoreAndDisallowV1BufWorkYAMLs() WorkspaceOption {
	return func(workspaceOptions *workspaceOptions) {
		workspaceOptions.ignoreAndDisallowV1BufWorkYAMLs = true
	}
}

// ImageOption is an option for image operations (GetImage, GetImageForInputConfig,
// GetImageForWorkspace, GetTargetImageWithConfigsAndCheckClient, GetImportableImageFileInfos,
// PutImage).
type ImageOption func(*imageOptions)

// WithImageTargetPaths returns a new ImageOption that sets the target paths and their
// corresponding CLI flag names for error messages.
func WithImageTargetPaths(targetPaths []string, targetExcludePaths []string, pathFlagName string, excludePathFlagName string) ImageOption {
	return func(imageOptions *imageOptions) {
		imageOptions.targetPaths = targetPaths
		imageOptions.targetExcludePaths = targetExcludePaths
		if pathFlagName != "" {
			imageOptions.pathFlagName = pathFlagName
		}
		if excludePathFlagName != "" {
			imageOptions.excludePathFlagName = excludePathFlagName
		}
	}
}

// WithImageConfigOverride applies the config override for image operations.
//
// See WithConfigOverride for details on behavior.
//
// *** DO NOT USE THIS OUTSIDE OF THE CLI AND/OR IF YOU DON'T UNDERSTAND IT. ***
// *** DO NOT ADD THIS TO ANY NEW COMMANDS. ***
func WithImageConfigOverride(configOverride string) ImageOption {
	return func(imageOptions *imageOptions) {
		imageOptions.configOverride = configOverride
	}
}

// WithImageExcludeSourceInfo returns a new ImageOption that excludes source code info.
func WithImageExcludeSourceInfo(imageExcludeSourceInfo bool) ImageOption {
	return func(imageOptions *imageOptions) {
		imageOptions.imageExcludeSourceInfo = imageExcludeSourceInfo
	}
}

// WithImageExcludeImports returns a new ImageOption that excludes imports.
func WithImageExcludeImports(imageExcludeImports bool) ImageOption {
	return func(imageOptions *imageOptions) {
		imageOptions.imageExcludeImports = imageExcludeImports
	}
}

// WithImageIncludeTypes returns a new ImageOption that includes the given types.
func WithImageIncludeTypes(imageTypes []string) ImageOption {
	return func(imageOptions *imageOptions) {
		imageOptions.imageIncludeTypes = imageTypes
	}
}

// WithImageExcludeTypes returns a new ImageOption that excludes the given types.
func WithImageExcludeTypes(imageExcludeTypes []string) ImageOption {
	return func(imageOptions *imageOptions) {
		imageOptions.imageExcludeTypes = imageExcludeTypes
	}
}

// WithImageAsFileDescriptorSet returns a new ImageOption that returns the image as a FileDescriptorSet.
func WithImageAsFileDescriptorSet(imageAsFileDescriptorSet bool) ImageOption {
	return func(imageOptions *imageOptions) {
		imageOptions.imageAsFileDescriptorSet = imageAsFileDescriptorSet
	}
}

// MessageOption is an option for message operations (GetMessage, PutMessage).
type MessageOption func(*messageOptions)

// WithMessageValidation returns a new MessageOption that says to validate the
// message as it is being read.
//
// We want to do this as part of the read/unmarshal, as protoyaml has specific logic
// on unmarshal that will pretty print validations.
func WithMessageValidation() MessageOption {
	return func(messageOptions *messageOptions) {
		messageOptions.messageValidation = true
	}
}

// *** PRIVATE ***

type workspaceOptions struct {
	copyToInMemory bool

	targetPaths                     []string
	targetExcludePaths              []string
	pathFlagName                    string
	excludePathFlagName             string
	configOverride                  string
	ignoreAndDisallowV1BufWorkYAMLs bool
}

func newWorkspaceOptions(controller *controller) *workspaceOptions {
	return &workspaceOptions{
		copyToInMemory:  controller.copyToInMemory,
		pathFlagName:    "path",
		excludePathFlagName: "exclude-path",
	}
}

func (f *workspaceOptions) getGetReadBucketCloserOptions() []buffetch.GetReadBucketCloserOption {
	var getReadBucketCloserOptions []buffetch.GetReadBucketCloserOption
	if f.copyToInMemory {
		getReadBucketCloserOptions = append(
			getReadBucketCloserOptions,
			buffetch.GetReadBucketCloserWithCopyToInMemory(),
		)
	}
	if len(f.targetPaths) > 0 {
		getReadBucketCloserOptions = append(
			getReadBucketCloserOptions,
			buffetch.GetReadBucketCloserWithTargetPaths(f.targetPaths),
		)
	}
	if len(f.targetExcludePaths) > 0 {
		getReadBucketCloserOptions = append(
			getReadBucketCloserOptions,
			buffetch.GetReadBucketCloserWithTargetExcludePaths(f.targetExcludePaths),
		)
	}
	if f.configOverride != "" {
		// If we have a config override, we do not search for buf.yamls or buf.work.yamls,
		// instead acting as if the config override was the only configuration file available.
		//
		// Note that this is slightly different behavior than the pre-refactor CLI had, but this
		// was always the intended behavior. The pre-refactor CLI would error if you had a buf.work.yaml,
		// and did the same search behavior for buf.yamls, which didn't really make sense. In the new
		// world where buf.yamls also represent the behavior of buf.work.yamls, you should be able
		// to specify whatever want here.
		getReadBucketCloserOptions = append(
			getReadBucketCloserOptions,
			buffetch.GetReadBucketCloserWithNoSearch(),
		)
	}
	return getReadBucketCloserOptions
}

func (f *workspaceOptions) getGetReadWriteBucketOptions() []buffetch.GetReadWriteBucketOption {
	if f.configOverride != "" {
		// If we have a config override, we do not search for buf.yamls or buf.work.yamls,
		// instead acting as if the config override was the only configuration file available.
		//
		// Note that this is slightly different behavior than the pre-refactor CLI had, but this
		// was always the intended behavior. The pre-refactor CLI would error if you had a buf.work.yaml,
		// and did the same search behavior for buf.yamls, which didn't really make sense. In the new
		// world where buf.yamls also represent the behavior of buf.work.yamls, you should be able
		// to specify whatever want here.
		return []buffetch.GetReadWriteBucketOption{
			buffetch.GetReadWriteBucketWithNoSearch(),
		}
	}
	return nil
}

type imageOptions struct {
	workspaceOptions

	imageExcludeSourceInfo   bool
	imageExcludeImports      bool
	imageIncludeTypes        []string
	imageExcludeTypes        []string
	imageAsFileDescriptorSet bool
}

func newImageOptions(controller *controller) *imageOptions {
	return &imageOptions{
		workspaceOptions: *newWorkspaceOptions(controller),
	}
}

type messageOptions struct {
	messageValidation bool
}

func newMessageOptions() *messageOptions {
	return &messageOptions{}
}
