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

package modulecreate

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	appcmd2 "github.com/inovacc/omni/pkg/buf/pkg/app/appcmd"
	"github.com/inovacc/omni/pkg/buf/pkg/app/appext"
	"github.com/inovacc/omni/pkg/buf/pkg/syserror"
	"github.com/spf13/pflag"

	"github.com/inovacc/omni/pkg/buf/internal/buf/bufcli"
	"github.com/inovacc/omni/pkg/buf/internal/buf/bufparse"
	"github.com/inovacc/omni/pkg/buf/internal/buf/bufprint"
	"github.com/inovacc/omni/pkg/buf/internal/buf/bufregistryapi/bufregistryapimodule"
	modulev1 "github.com/inovacc/omni/pkg/buf/internal/gen/proto/go/buf/registry/module/v1"
	ownerv1 "github.com/inovacc/omni/pkg/buf/internal/gen/proto/go/buf/registry/owner/v1"
)

const (
	formatFlagName      = "format"
	visibilityFlagName  = "visibility"
	defaultLabeFlagName = "default-label-name"

	defaultDefaultLabel = "main"
)

// NewCommand returns a new Command
func NewCommand(
	name string,
	builder appext.SubCommandBuilder,
) *appcmd2.Command {
	flags := newFlags()
	return &appcmd2.Command{
		Use:   name + " <remote/owner/module>",
		Short: "Create a BSR module",
		Args:  appcmd2.ExactArgs(1),
		Run: builder.NewRunFunc(
			func(ctx context.Context, container appext.Container) error {
				return run(ctx, container, flags)
			},
		),
		BindFlags: flags.Bind,
	}
}

type flags struct {
	Format       string
	Visibility   string
	DefaultLabel string
}

func newFlags() *flags {
	return &flags{}
}

func (f *flags) Bind(flagSet *pflag.FlagSet) {
	bufcli.BindVisibility(flagSet, &f.Visibility, visibilityFlagName, false)
	flagSet.StringVar(
		&f.Format,
		formatFlagName,
		bufprint.FormatText.String(),
		fmt.Sprintf(`The output format to use. Must be one of %s`, bufprint.AllFormatsString),
	)
	flagSet.StringVar(
		&f.DefaultLabel,
		defaultLabeFlagName,
		defaultDefaultLabel,
		"The default label name of the module",
	)
}

func run(
	ctx context.Context,
	container appext.Container,
	flags *flags,
) error {
	moduleFullName, err := bufparse.ParseFullName(container.Arg(0))
	if err != nil {
		return appcmd2.WrapInvalidArgumentError(err)
	}
	visibility, err := bufcli.VisibilityFlagToVisibilityAllowUnspecified(flags.Visibility)
	if err != nil {
		return appcmd2.WrapInvalidArgumentError(err)
	}
	format, err := bufprint.ParseFormat(flags.Format)
	if err != nil {
		return appcmd2.WrapInvalidArgumentError(err)
	}

	clientConfig, err := bufcli.NewConnectClientConfig(container)
	if err != nil {
		return err
	}
	moduleServiceClient := bufregistryapimodule.NewClientProvider(clientConfig).V1ModuleServiceClient(moduleFullName.Registry())
	resp, err := moduleServiceClient.CreateModules(
		ctx,
		connect.NewRequest(
			&modulev1.CreateModulesRequest{
				Values: []*modulev1.CreateModulesRequest_Value{
					{
						OwnerRef: &ownerv1.OwnerRef{
							Value: &ownerv1.OwnerRef_Name{
								Name: moduleFullName.Owner(),
							},
						},
						Name:             moduleFullName.Name(),
						Visibility:       visibility,
						DefaultLabelName: flags.DefaultLabel,
					},
				},
			},
		),
	)
	if err != nil {
		if connect.CodeOf(err) == connect.CodeAlreadyExists {
			return bufcli.NewModuleNameAlreadyExistsError(moduleFullName.String())
		}
		return err
	}
	modules := resp.Msg.Modules
	if len(modules) != 1 {
		return syserror.Newf("unexpected number of modules returned from server: %d", len(modules))
	}
	if format == bufprint.FormatText {
		_, err = fmt.Fprintf(container.Stdout(), "Created %s.\n", moduleFullName)
		if err != nil {
			return syserror.Wrap(err)
		}
		return nil
	}
	return bufprint.PrintNames(
		container.Stdout(),
		format,
		bufprint.NewModuleEntity(modules[0], moduleFullName),
	)
}
