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

package webhookcreate

import (
	"context"
	"encoding/json"
	"fmt"

	"connectrpc.com/connect"
	appcmd2 "github.com/inovacc/omni/pkg/buf/pkg/app/appcmd"
	"github.com/inovacc/omni/pkg/buf/pkg/app/appext"
	"github.com/inovacc/omni/pkg/buf/pkg/connectclient"
	"github.com/spf13/pflag"

	"github.com/inovacc/omni/pkg/buf/internal/buf/bufcli"
	"github.com/inovacc/omni/pkg/buf/internal/gen/proto/connect/buf/alpha/registry/v1alpha1/registryv1alpha1connect"
	registryv1alpha1 "github.com/inovacc/omni/pkg/buf/internal/gen/proto/go/buf/alpha/registry/v1alpha1"
)

const (
	ownerFlagName        = "owner"
	repositoryFlagName   = "repository"
	callbackURLFlagName  = "callback-url"
	webhookEventFlagName = "event"
	remoteFlagName       = "remote"
)

// NewCommand returns a new Command
func NewCommand(
	name string,
	builder appext.SubCommandBuilder,
) *appcmd2.Command {
	flags := newFlags()
	return &appcmd2.Command{
		Use:   name,
		Short: "Create a repository webhook",
		Args:  appcmd2.ExactArgs(0),
		Run: builder.NewRunFunc(
			func(ctx context.Context, container appext.Container) error {
				return run(ctx, container, flags)
			},
		),
		BindFlags: flags.Bind,
	}
}

type flags struct {
	WebhookEvent   string
	OwnerName      string
	RepositoryName string
	CallbackURL    string
	Remote         string
}

func newFlags() *flags {
	return &flags{}
}

func (f *flags) Bind(flagSet *pflag.FlagSet) {
	flagSet.StringVar(
		&f.WebhookEvent,
		webhookEventFlagName,
		"",
		"The event type to create a webhook for. The proto enum string value is used for this input (e.g. 'WEBHOOK_EVENT_REPOSITORY_PUSH')",
	)
	_ = appcmd2.MarkFlagRequired(flagSet, webhookEventFlagName)
	flagSet.StringVar(
		&f.OwnerName,
		ownerFlagName,
		"",
		`The owner name of the repository to create a webhook for`,
	)
	_ = appcmd2.MarkFlagRequired(flagSet, ownerFlagName)
	flagSet.StringVar(
		&f.RepositoryName,
		repositoryFlagName,
		"",
		"The repository name to create a webhook for",
	)
	_ = appcmd2.MarkFlagRequired(flagSet, repositoryFlagName)
	flagSet.StringVar(
		&f.CallbackURL,
		callbackURLFlagName,
		"",
		"The url for the webhook to callback to on a given event",
	)
	_ = appcmd2.MarkFlagRequired(flagSet, callbackURLFlagName)
	flagSet.StringVar(
		&f.Remote,
		remoteFlagName,
		"",
		"The remote of the repository the created webhook will belong to",
	)
	_ = appcmd2.MarkFlagRequired(flagSet, remoteFlagName)
}

func run(
	ctx context.Context,
	container appext.Container,
	flags *flags,
) error {
	bufcli.WarnBetaCommand(ctx, container)
	clientConfig, err := bufcli.NewConnectClientConfig(container)
	if err != nil {
		return err
	}
	service := connectclient.Make(clientConfig, flags.Remote, registryv1alpha1connect.NewWebhookServiceClient)
	event, ok := registryv1alpha1.WebhookEvent_value[flags.WebhookEvent]
	if !ok || event == int32(registryv1alpha1.WebhookEvent_WEBHOOK_EVENT_UNSPECIFIED) {
		return fmt.Errorf("webhook event must be specified")
	}
	resp, err := service.CreateWebhook(
		ctx,
		connect.NewRequest(
			registryv1alpha1.CreateWebhookRequest_builder{
				WebhookEvent:   registryv1alpha1.WebhookEvent(event),
				OwnerName:      flags.OwnerName,
				RepositoryName: flags.RepositoryName,
				CallbackUrl:    flags.CallbackURL,
			}.Build(),
		),
	)
	if err != nil {
		return err
	}
	webhookJSON, err := json.MarshalIndent(resp.Msg.GetWebhook(), "", "\t")
	if err != nil {
		return err
	}
	// Ignore errors for writing to stdout.
	_, _ = container.Stdout().Write(webhookJSON)
	return nil
}
