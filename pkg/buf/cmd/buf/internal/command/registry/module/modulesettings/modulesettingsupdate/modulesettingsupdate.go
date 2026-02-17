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

package modulesettingsupdate

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
	"github.com/inovacc/omni/pkg/buf/internal/buf/bufregistryapi/bufregistryapimodule"
	modulev1 "github.com/inovacc/omni/pkg/buf/internal/gen/proto/go/buf/registry/module/v1"
)

const (
	visibilityFlagName   = "visibility"
	descriptionFlagName  = "description"
	urlFlagName          = "url"
	defaultLabelFlagName = "default-label-name"
)

// NewCommand returns a new Command
func NewCommand(
	name string,
	builder appext.SubCommandBuilder,
	deprecated string,
) *appcmd2.Command {
	flags := newFlags()
	return &appcmd2.Command{
		Use:        name + " <remote/owner/module>",
		Short:      "Update BSR module settings",
		Args:       appcmd2.ExactArgs(1),
		Deprecated: deprecated,
		Run: builder.NewRunFunc(
			func(ctx context.Context, container appext.Container) error {
				return run(ctx, container, flags)
			},
		),
		BindFlags: flags.Bind,
	}
}

type flags struct {
	Visibility   string
	Description  *string
	URL          *string
	DefaultLabel string
}

func newFlags() *flags {
	return &flags{}
}

func (f *flags) Bind(flagSet *pflag.FlagSet) {
	bufcli.BindVisibility(flagSet, &f.Visibility, visibilityFlagName, true)
	bufcli.BindStringPointer(
		flagSet,
		descriptionFlagName,
		&f.Description,
		"The new description for the module",
	)
	bufcli.BindStringPointer(
		flagSet,
		urlFlagName,
		&f.URL,
		"The new URL for the module",
	)
	flagSet.StringVar(
		&f.DefaultLabel,
		defaultLabelFlagName,
		"",
		"The label that commits are pushed to by default",
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
	clientConfig, err := bufcli.NewConnectClientConfig(container)
	if err != nil {
		return err
	}
	moduleServiceClient := bufregistryapimodule.NewClientProvider(clientConfig).V1ModuleServiceClient(moduleFullName.Registry())
	visibilityUpdate := &visibility
	if visibility == modulev1.ModuleVisibility_MODULE_VISIBILITY_UNSPECIFIED {
		visibilityUpdate = nil
	}
	defaultLabelUpdate := &flags.DefaultLabel
	if flags.DefaultLabel == "" {
		defaultLabelUpdate = nil
	}
	if _, err := moduleServiceClient.UpdateModules(
		ctx,
		&connect.Request[modulev1.UpdateModulesRequest]{
			Msg: &modulev1.UpdateModulesRequest{
				Values: []*modulev1.UpdateModulesRequest_Value{
					{
						ModuleRef: &modulev1.ModuleRef{
							Value: &modulev1.ModuleRef_Name_{
								Name: &modulev1.ModuleRef_Name{
									Owner:  moduleFullName.Owner(),
									Module: moduleFullName.Name(),
								},
							},
						},
						Description:      flags.Description,
						Url:              flags.URL,
						Visibility:       visibilityUpdate,
						DefaultLabelName: defaultLabelUpdate,
					},
				},
			},
		},
	); err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			return bufcli.NewModuleNotFoundError(container.Arg(0))
		}
		return err
	}
	if _, err := fmt.Fprintf(container.Stdout(), "Updated %s.\n", moduleFullName); err != nil {
		return syserror.Wrap(err)
	}
	return nil
}
