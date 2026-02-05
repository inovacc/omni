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

package build

import (
	"context"
	"fmt"

	"github.com/inovacc/omni/pkg/buf/internal/app"
	"github.com/inovacc/omni/pkg/buf/internal/app/appcmd"
	"github.com/inovacc/omni/pkg/buf/internal/app/appext"
	bufcli2 "github.com/inovacc/omni/pkg/buf/internal/buf/bufcli"
	"github.com/inovacc/omni/pkg/buf/internal/buf/bufctl"
	"github.com/inovacc/omni/pkg/buf/internal/buf/buffetch"
	"github.com/inovacc/omni/pkg/buf/internal/bufpkg/bufanalysis"
	"github.com/inovacc/omni/pkg/buf/internal/bufpkg/bufimage/bufimageutil"
	"github.com/inovacc/omni/pkg/buf/internal/standard/xstrings"
	"github.com/spf13/pflag"
)

const (
	asFileDescriptorSetFlagName           = "as-file-descriptor-set"
	errorFormatFlagName                   = "error-format"
	excludeImportsFlagName                = "exclude-imports"
	excludeSourceInfoFlagName             = "exclude-source-info"
	excludeSourceRetentionOptionsFlagName = "exclude-source-retention-options"
	pathsFlagName                         = "path"
	outputFlagName                        = "output"
	outputFlagShortName                   = "o"
	configFlagName                        = "config"
	excludePathsFlagName                  = "exclude-path"
	disableSymlinksFlagName               = "disable-symlinks"
	typeFlagName                          = "type"
)

// NewCommand returns a new Command.
func NewCommand(
	name string,
	builder appext.SubCommandBuilder,
) *appcmd.Command {
	flags := newFlags()
	return &appcmd.Command{
		Use:   name + " <input>",
		Short: "Build Protobuf files into a Buf image",
		Long:  bufcli2.GetInputLong(`the source or module to build or image to convert`),
		Args:  appcmd.MaximumNArgs(1),
		Run: builder.NewRunFunc(
			func(ctx context.Context, container appext.Container) error {
				return run(ctx, container, flags)
			},
		),
		BindFlags: flags.Bind,
	}
}

type flags struct {
	AsFileDescriptorSet           bool
	ErrorFormat                   string
	ExcludeImports                bool
	ExcludeSourceInfo             bool
	ExcludeSourceRetentionOptions bool
	Paths                         []string
	Output                        string
	Config                        string
	ExcludePaths                  []string
	DisableSymlinks               bool
	Types                         []string
	// special
	InputHashtag string
}

func newFlags() *flags {
	return &flags{}
}

func (f *flags) Bind(flagSet *pflag.FlagSet) {
	bufcli2.BindInputHashtag(flagSet, &f.InputHashtag)
	bufcli2.BindAsFileDescriptorSet(flagSet, &f.AsFileDescriptorSet, asFileDescriptorSetFlagName)
	bufcli2.BindExcludeImports(flagSet, &f.ExcludeImports, excludeImportsFlagName)
	bufcli2.BindExcludeSourceInfo(flagSet, &f.ExcludeSourceInfo, excludeSourceInfoFlagName)
	bufcli2.BindPaths(flagSet, &f.Paths, pathsFlagName)
	bufcli2.BindExcludePaths(flagSet, &f.ExcludePaths, excludePathsFlagName)
	bufcli2.BindDisableSymlinks(flagSet, &f.DisableSymlinks, disableSymlinksFlagName)
	flagSet.BoolVar(
		&f.ExcludeSourceRetentionOptions,
		excludeSourceRetentionOptionsFlagName,
		false,
		"Exclude options whose retention is source",
	)
	flagSet.StringVar(
		&f.ErrorFormat,
		errorFormatFlagName,
		"text",
		fmt.Sprintf(
			"The format for build errors printed to stderr. Must be one of %s",
			xstrings.SliceToString(bufanalysis.AllFormatStrings),
		),
	)
	flagSet.StringVarP(
		&f.Output,
		outputFlagName,
		outputFlagShortName,
		app.DevNullFilePath,
		fmt.Sprintf(
			`The output location for the built image. Must be one of format %s`,
			buffetch.MessageFormatsString,
		),
	)
	flagSet.StringVar(
		&f.Config,
		configFlagName,
		"",
		`The buf.yaml file or data to use for configuration`,
	)
	flagSet.StringSliceVar(
		&f.Types,
		typeFlagName,
		nil,
		"The types (package, message, enum, extension, service, method) that should be included in this image. When specified, the resulting image will only include descriptors to describe the requested types",
	)
}

func run(
	ctx context.Context,
	container appext.Container,
	flags *flags,
) error {
	if err := bufcli2.ValidateRequiredFlag(outputFlagName, flags.Output); err != nil {
		return err
	}
	input, err := bufcli2.GetInputValue(container, flags.InputHashtag, ".")
	if err != nil {
		return err
	}
	controller, err := bufcli2.NewController(
		container,
		bufctl.WithDisableSymlinks(flags.DisableSymlinks),
		bufctl.WithFileAnnotationErrorFormat(flags.ErrorFormat),
	)
	if err != nil {
		return err
	}
	image, err := controller.GetImage(
		ctx,
		input,
		bufctl.WithTargetPaths(flags.Paths, flags.ExcludePaths),
		bufctl.WithImageExcludeSourceInfo(flags.ExcludeSourceInfo),
		bufctl.WithImageExcludeImports(flags.ExcludeImports),
		bufctl.WithImageIncludeTypes(flags.Types),
		bufctl.WithConfigOverride(flags.Config),
	)
	if err != nil {
		return err
	}
	if flags.ExcludeSourceRetentionOptions {
		image, err = bufimageutil.StripSourceRetentionOptions(image)
		if err != nil {
			return err
		}
	}
	return controller.PutImage(
		ctx,
		flags.Output,
		image,
		bufctl.WithImageAsFileDescriptorSet(flags.AsFileDescriptorSet),
	)
}
