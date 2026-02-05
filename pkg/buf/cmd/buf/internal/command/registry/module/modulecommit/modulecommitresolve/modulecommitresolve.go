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

package modulecommitresolve

import (
	"context"
	"fmt"

	"github.com/inovacc/omni/pkg/buf/internal/app/appcmd"
	"github.com/inovacc/omni/pkg/buf/internal/app/appext"
	bufcli2 "github.com/inovacc/omni/pkg/buf/internal/buf/bufcli"
	"github.com/inovacc/omni/pkg/buf/internal/buf/bufprint"
	"github.com/inovacc/omni/pkg/buf/internal/bufpkg/bufparse"
	"github.com/inovacc/omni/pkg/buf/internal/bufpkg/bufregistryapi/bufregistryapimodule"
	"github.com/inovacc/omni/pkg/buf/internal/connect"
	modulev1 "github.com/inovacc/omni/pkg/buf/internal/gen/bufbuild/registry/protocolbuffers/go/buf/registry/module/v1"
	"github.com/inovacc/omni/pkg/buf/internal/pkg/syserror"
	"github.com/spf13/pflag"
)

const formatFlagName = "format"

// NewCommand returns a new Command
func NewCommand(
	name string,
	builder appext.SubCommandBuilder,
	deprecated string,
) *appcmd.Command {
	flags := newFlags()
	return &appcmd.Command{
		Use:        name + " <remote/owner/repository[:ref]>",
		Short:      "Resolve commit from a reference",
		Args:       appcmd.ExactArgs(1),
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
	commitServiceClient := bufregistryapimodule.NewClientProvider(clientConfig).V1CommitServiceClient(moduleRef.FullName().Registry())
	resp, err := commitServiceClient.GetCommits(
		ctx,
		connect.NewRequest(
			&modulev1.GetCommitsRequest{
				ResourceRefs: []*modulev1.ResourceRef{
					{
						Value: &modulev1.ResourceRef_Name_{
							Name: &modulev1.ResourceRef_Name{
								Owner:  moduleRef.FullName().Owner(),
								Module: moduleRef.FullName().Name(),
								Child: &modulev1.ResourceRef_Name_Ref{
									Ref: moduleRef.Ref(),
								},
							},
						},
					},
				},
			},
		),
	)
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			return bufcli2.NewRefNotFoundError(moduleRef)
		}
		return err
	}
	commits := resp.Msg.Commits
	if len(commits) != 1 {
		return syserror.Newf("expect 1 commit from response, got %d", len(commits))
	}
	commit := commits[0]
	return bufprint.PrintNames(
		container.Stdout(),
		format,
		bufprint.NewCommitEntity(commit, moduleRef.FullName(), commit.GetSourceControlUrl()),
	)
}
