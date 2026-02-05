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

package moduledelete

import (
	"context"
	"fmt"

	"github.com/inovacc/omni/pkg/buf/internal/app/appcmd"
	"github.com/inovacc/omni/pkg/buf/internal/app/appext"
	bufcli2 "github.com/inovacc/omni/pkg/buf/internal/buf/bufcli"
	"github.com/inovacc/omni/pkg/buf/internal/bufpkg/bufparse"
	"github.com/inovacc/omni/pkg/buf/internal/bufpkg/bufregistryapi/bufregistryapimodule"
	"github.com/inovacc/omni/pkg/buf/internal/connect"
	modulev1 "github.com/inovacc/omni/pkg/buf/internal/gen/bufbuild/registry/protocolbuffers/go/buf/registry/module/v1"
	"github.com/inovacc/omni/pkg/buf/internal/pkg/syserror"
	"github.com/spf13/pflag"
)

const forceFlagName = "force"

// NewCommand returns a new Command
func NewCommand(
	name string,
	builder appext.SubCommandBuilder,
) *appcmd.Command {
	flags := newFlags()
	return &appcmd.Command{
		Use:   name + " <remote/owner/module>",
		Short: "Delete a BSR module",
		Args:  appcmd.ExactArgs(1),
		Run: builder.NewRunFunc(
			func(ctx context.Context, container appext.Container) error {
				return run(ctx, container, flags)
			},
		),
		BindFlags: flags.Bind,
	}
}

type flags struct {
	Force bool
}

func newFlags() *flags {
	return &flags{}
}

func (f *flags) Bind(flagSet *pflag.FlagSet) {
	flagSet.BoolVar(
		&f.Force,
		forceFlagName,
		false,
		"Force deletion without confirming. Use with caution",
	)
}

func run(
	ctx context.Context,
	container appext.Container,
	flags *flags,
) error {
	moduleFullName, err := bufparse.ParseFullName(container.Arg(0))
	if err != nil {
		return appcmd.WrapInvalidArgumentError(err)
	}
	clientConfig, err := bufcli2.NewConnectClientConfig(container)
	if err != nil {
		return err
	}
	if !flags.Force {
		if err := bufcli2.PromptUserForDelete(container, "entity", moduleFullName.Name()); err != nil {
			return err
		}
	}
	moduleServiceClient := bufregistryapimodule.NewClientProvider(clientConfig).V1ModuleServiceClient(moduleFullName.Registry())
	if _, err := moduleServiceClient.DeleteModules(
		ctx,
		connect.NewRequest(
			&modulev1.DeleteModulesRequest{
				ModuleRefs: []*modulev1.ModuleRef{
					{
						Value: &modulev1.ModuleRef_Name_{
							Name: &modulev1.ModuleRef_Name{
								Owner:  moduleFullName.Owner(),
								Module: moduleFullName.Name(),
							},
						},
					},
				},
			},
		),
	); err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			return bufcli2.NewModuleNotFoundError(container.Arg(0))
		}
		return err
	}
	if _, err := fmt.Fprintf(container.Stdout(), "Deleted %s.\n", moduleFullName); err != nil {
		return syserror.Wrap(err)
	}
	return nil
}
