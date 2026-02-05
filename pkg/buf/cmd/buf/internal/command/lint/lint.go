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

package lint

import (
	"context"
	"errors"
	"fmt"

	"github.com/inovacc/omni/pkg/buf/internal/app/appcmd"
	"github.com/inovacc/omni/pkg/buf/internal/app/appext"
	bufcli2 "github.com/inovacc/omni/pkg/buf/internal/buf/bufcli"
	bufctl2 "github.com/inovacc/omni/pkg/buf/internal/buf/bufctl"
	"github.com/inovacc/omni/pkg/buf/internal/bufpkg/bufanalysis"
	"github.com/inovacc/omni/pkg/buf/internal/bufpkg/bufcheck"
	"github.com/inovacc/omni/pkg/buf/internal/bufpkg/bufconfig"
	"github.com/inovacc/omni/pkg/buf/internal/standard/xstrings"
	"github.com/spf13/pflag"
)

const (
	errorFormatFlagName     = "error-format"
	configFlagName          = "config"
	pathsFlagName           = "path"
	excludePathsFlagName    = "exclude-path"
	disableSymlinksFlagName = "disable-symlinks"
)

// NewCommand returns a new Command.
func NewCommand(
	name string,
	builder appext.SubCommandBuilder,
) *appcmd.Command {
	flags := newFlags()
	return &appcmd.Command{
		Use:   name + " <input>",
		Short: "Run linting on Protobuf files",
		Long:  bufcli2.GetInputLong(`the source, module, or Image to lint`),
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
	ErrorFormat     string
	Config          string
	Paths           []string
	ExcludePaths    []string
	DisableSymlinks bool
	// special
	InputHashtag string
}

func newFlags() *flags {
	return &flags{}
}

func (f *flags) Bind(flagSet *pflag.FlagSet) {
	bufcli2.BindInputHashtag(flagSet, &f.InputHashtag)
	bufcli2.BindPaths(flagSet, &f.Paths, pathsFlagName)
	bufcli2.BindExcludePaths(flagSet, &f.ExcludePaths, excludePathsFlagName)
	bufcli2.BindDisableSymlinks(flagSet, &f.DisableSymlinks, disableSymlinksFlagName)
	flagSet.StringVar(
		&f.ErrorFormat,
		errorFormatFlagName,
		"text",
		fmt.Sprintf(
			"The format for build errors or check violations printed to stdout. Must be one of %s",
			xstrings.SliceToString(bufcli2.AllLintFormatStrings),
		),
	)
	flagSet.StringVar(
		&f.Config,
		configFlagName,
		"",
		`The buf.yaml file or data to use for configuration`,
	)
}

func run(
	ctx context.Context,
	container appext.Container,
	flags *flags,
) (retErr error) {
	if err := bufcli2.ValidateErrorFormatFlagLint(flags.ErrorFormat, errorFormatFlagName); err != nil {
		return err
	}
	// Parse out if this is config-ignore-yaml.
	// This is messed.
	controllerErrorFormat := flags.ErrorFormat
	if controllerErrorFormat == "config-ignore-yaml" {
		controllerErrorFormat = "text"
	}
	input, err := bufcli2.GetInputValue(container, flags.InputHashtag, ".")
	if err != nil {
		return err
	}
	controller, err := bufcli2.NewController(
		container,
		bufctl2.WithDisableSymlinks(flags.DisableSymlinks),
		bufctl2.WithFileAnnotationErrorFormat(controllerErrorFormat),
		bufctl2.WithFileAnnotationsToStdout(),
	)
	if err != nil {
		return err
	}
	wasmRuntime, err := bufcli2.NewWasmRuntime(ctx, container)
	if err != nil {
		return err
	}
	defer func() {
		retErr = errors.Join(retErr, wasmRuntime.Close(ctx))
	}()
	imageWithConfigs, checkClient, err := controller.GetTargetImageWithConfigsAndCheckClient(
		ctx,
		input,
		wasmRuntime,
		bufctl2.WithTargetPaths(flags.Paths, flags.ExcludePaths),
		bufctl2.WithConfigOverride(flags.Config),
	)
	if err != nil {
		return err
	}
	var allFileAnnotations []bufanalysis.FileAnnotation
	// We add all check configs (both lint and breaking) as related configs to check if plugins
	// have rules configured.
	// We allocated twice the size of imageWithConfigs for both lint and breaking configs.
	allCheckConfigs := make([]bufconfig.CheckConfig, 0, len(imageWithConfigs)*2)
	for _, imageWithConfig := range imageWithConfigs {
		allCheckConfigs = append(allCheckConfigs, imageWithConfig.LintConfig())
		allCheckConfigs = append(allCheckConfigs, imageWithConfig.BreakingConfig())
	}
	for _, imageWithConfig := range imageWithConfigs {
		lintOptions := []bufcheck.LintOption{
			bufcheck.WithPluginConfigs(imageWithConfig.PluginConfigs()...),
			bufcheck.WithPolicyConfigs(imageWithConfig.PolicyConfigs()...),
			bufcheck.WithRelatedCheckConfigs(allCheckConfigs...),
		}
		if err := checkClient.Lint(
			ctx,
			imageWithConfig.LintConfig(),
			imageWithConfig,
			lintOptions...,
		); err != nil {
			var fileAnnotationSet bufanalysis.FileAnnotationSet
			if errors.As(err, &fileAnnotationSet) {
				allFileAnnotations = append(allFileAnnotations, fileAnnotationSet.FileAnnotations()...)
			} else {
				return err
			}
		}
	}
	if len(allFileAnnotations) > 0 {
		allFileAnnotationSet := bufanalysis.NewFileAnnotationSet(allFileAnnotations...)
		if flags.ErrorFormat == "config-ignore-yaml" {
			if err := bufcli2.PrintFileAnnotationSetLintConfigIgnoreYAMLV1(
				container.Stdout(),
				allFileAnnotationSet,
			); err != nil {
				return err
			}
		} else {
			if err := bufanalysis.PrintFileAnnotationSet(
				container.Stdout(),
				allFileAnnotationSet,
				flags.ErrorFormat,
			); err != nil {
				return err
			}
		}
		return bufctl2.ErrFileAnnotation
	}
	return nil
}
