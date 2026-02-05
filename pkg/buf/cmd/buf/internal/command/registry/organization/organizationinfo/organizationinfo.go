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

package organizationinfo

import (
	"context"
	"fmt"

	ownerv1 "github.com/inovacc/omni/pkg/buf/internal/gen/bufbuild/registry/protocolbuffers/go/buf/registry/owner/v1"
	"github.com/inovacc/omni/pkg/buf/internal/app/appcmd"
	"github.com/inovacc/omni/pkg/buf/internal/app/appext"
	"github.com/inovacc/omni/pkg/buf/internal/connect"
	bufcli2 "github.com/inovacc/omni/pkg/buf/internal/buf/bufcli"
	"github.com/inovacc/omni/pkg/buf/internal/buf/bufprint"
	"github.com/inovacc/omni/pkg/buf/internal/bufpkg/bufregistryapi/bufregistryapiowner"
	"github.com/inovacc/omni/pkg/buf/internal/pkg/syserror"
	"github.com/spf13/pflag"
)

const formatFlagName = "format"

// NewCommand returns a new Command
func NewCommand(
	name string,
	builder appext.SubCommandBuilder,
) *appcmd.Command {
	flags := newFlags()
	return &appcmd.Command{
		Use:   name + " <remote/organization>",
		Short: "Show information about a BSR organization",
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
	Format string
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
}

func run(
	ctx context.Context,
	container appext.Container,
	flags *flags,
) error {
	moduleOwner, err := bufcli2.ParseModuleOwner(container.Arg(0))
	if err != nil {
		return appcmd.WrapInvalidArgumentError(err)
	}
	format, err := bufprint.ParseFormat(flags.Format)
	if err != nil {
		return appcmd.WrapInvalidArgumentError(err)
	}

	clientConfig, err := bufcli2.NewConnectClientConfig(container)
	if err != nil {
		return err
	}
	ownerClientProvider := bufregistryapiowner.NewClientProvider(clientConfig)
	organizationServiceClient := ownerClientProvider.V1OrganizationServiceClient(moduleOwner.Registry())
	resp, err := organizationServiceClient.GetOrganizations(
		ctx,
		connect.NewRequest(
			&ownerv1.GetOrganizationsRequest{
				OrganizationRefs: []*ownerv1.OrganizationRef{
					{
						Value: &ownerv1.OrganizationRef_Name{
							Name: moduleOwner.Owner(),
						},
					},
				},
			},
		),
	)
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			return bufcli2.NewOrganizationNotFoundError(container.Arg(0))
		}
		return err
	}
	organizations := resp.Msg.GetOrganizations()
	if len(organizations) != 1 {
		return syserror.Newf("unexpected number of organizations returned from server: %d", len(organizations))
	}
	return bufprint.PrintEntity(
		container.Stdout(),
		format,
		bufprint.NewOrganizationEntity(organizations[0], moduleOwner.Registry()),
	)
}
