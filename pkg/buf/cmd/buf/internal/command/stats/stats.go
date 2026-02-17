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

package stats

import (
	"context"
	"fmt"

	appcmd2 "github.com/inovacc/omni/pkg/buf/pkg/app/appcmd"
	"github.com/inovacc/omni/pkg/buf/pkg/app/appext"
	"github.com/inovacc/omni/pkg/buf/pkg/protostat"
	"github.com/inovacc/omni/pkg/buf/pkg/protostat/protostatstorage"
	"github.com/spf13/pflag"

	"github.com/inovacc/omni/pkg/buf/internal/buf/bufcli"
	"github.com/inovacc/omni/pkg/buf/internal/buf/bufctl"
	"github.com/inovacc/omni/pkg/buf/internal/buf/bufmodule"
	"github.com/inovacc/omni/pkg/buf/internal/buf/bufprint"
)

const (
	formatFlagName          = "format"
	disableSymlinksFlagName = "disable-symlinks"
)

// NewCommand returns a new Command.
func NewCommand(
	name string,
	builder appext.SubCommandBuilder,
) *appcmd2.Command {
	flags := newFlags()
	return &appcmd2.Command{
		Use:   name + " <source>",
		Short: "Get statistics for a given source or module",
		Long:  bufcli.GetSourceOrModuleLong(`the source or module to get statistics for`),
		Args:  appcmd2.MaximumNArgs(1),
		Run: builder.NewRunFunc(
			func(ctx context.Context, container appext.Container) error {
				return run(ctx, container, flags)
			},
		),
		BindFlags: flags.Bind,
	}
}

type flags struct {
	Format          string
	DisableSymlinks bool

	// special
	InputHashtag string
}

func newFlags() *flags {
	return &flags{}
}

func (f *flags) Bind(flagSet *pflag.FlagSet) {
	flagSet.StringVar(
		&f.Format,
		formatFlagName,
		bufprint.FormatText.String(),
		fmt.Sprintf(`The output format to use. Must be one of %s`, bufprint.AllFormatsString),
	)
	bufcli.BindDisableSymlinks(flagSet, &f.DisableSymlinks, disableSymlinksFlagName)
	bufcli.BindInputHashtag(flagSet, &f.InputHashtag)
}

func run(
	ctx context.Context,
	container appext.Container,
	flags *flags,
) error {
	format, err := bufprint.ParseFormat(flags.Format)
	if err != nil {
		return appcmd2.WrapInvalidArgumentError(err)
	}
	input, err := bufcli.GetInputValue(container, flags.InputHashtag, ".")
	if err != nil {
		return err
	}
	controller, err := bufcli.NewController(
		container,
		bufctl.WithDisableSymlinks(flags.DisableSymlinks),
	)
	if err != nil {
		return err
	}
	workspace, err := controller.GetWorkspace(
		ctx,
		input,
	)
	if err != nil {
		return err
	}
	stats, err := protostat.GetStats(
		ctx,
		protostatstorage.NewFileWalker(
			bufmodule.ModuleReadBucketToStorageReadBucket(
				bufmodule.ModuleSetToModuleReadBucketWithOnlyProtoFilesForTargetModules(
					workspace,
				),
			),
		),
	)
	if err != nil {
		return err
	}
	return bufprint.NewStatsPrinter(container.Stdout()).PrintStats(
		ctx,
		format,
		stats,
	)
}
