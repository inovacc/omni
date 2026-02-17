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

package modulelabelinfo

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
)

const formatFlagName = "format"

// NewCommand returns a new Command
func NewCommand(
	name string,
	builder appext.SubCommandBuilder,
	deprecated string,
) *appcmd2.Command {
	flags := newFlags()
	return &appcmd2.Command{
		Use:        name + " <remote/owner/module:label>",
		Short:      "Show label information",
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
	moduleRef, err := bufparse.ParseRef(container.Arg(0))
	if err != nil {
		return appcmd2.WrapInvalidArgumentError(err)
	}
	labelName := moduleRef.Ref()
	if labelName == "" {
		return appcmd2.NewInvalidArgumentError("label is required")
	}
	format, err := bufprint.ParseFormat(flags.Format)
	if err != nil {
		return appcmd2.WrapInvalidArgumentError(err)
	}
	clientConfig, err := bufcli.NewConnectClientConfig(container)
	if err != nil {
		return err
	}
	moduleClientProvider := bufregistryapimodule.NewClientProvider(clientConfig)
	moduleFullName := moduleRef.FullName()
	labelServiceClient := moduleClientProvider.V1LabelServiceClient(moduleFullName.Registry())
	resp, err := labelServiceClient.GetLabels(
		ctx,
		connect.NewRequest(
			&modulev1.GetLabelsRequest{
				LabelRefs: []*modulev1.LabelRef{
					{
						Value: &modulev1.LabelRef_Name_{
							Name: &modulev1.LabelRef_Name{
								Owner:  moduleFullName.Owner(),
								Module: moduleFullName.Name(),
								Label:  labelName,
							},
						},
					},
				},
			},
		),
	)
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			return bufcli.NewLabelNotFoundError(moduleRef)
		}
		return err
	}
	labels := resp.Msg.Labels
	if len(labels) != 1 {
		return syserror.Newf("expect 1 label from response, got %d", len(labels))
	}
	return bufprint.PrintEntity(
		container.Stdout(),
		format,
		bufprint.NewLabelEntity(labels[0], moduleFullName),
	)
}
