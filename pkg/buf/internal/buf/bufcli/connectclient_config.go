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

package bufcli

import (
	"buf.build/go/app/appext"
	"connectrpc.com/connect"
	otelconnect "connectrpc.com/otelconnect"
	"github.com/inovacc/omni/pkg/buf/internal/buf/bufapp"
	bufconnect2 "github.com/inovacc/omni/pkg/buf/internal/bufpkg/bufconnect"
	"github.com/inovacc/omni/pkg/buf/internal/bufpkg/buftransport"
	"github.com/inovacc/omni/pkg/buf/internal/pkg/connectclient"
	"github.com/inovacc/omni/pkg/buf/internal/pkg/netrc"
	"github.com/inovacc/omni/pkg/buf/internal/pkg/transport/http/httpclient"
)

// NewConnectClientConfig creates a new connect.ClientConfig which uses a token reader to look
// up the token in the container or in netrc based on the address of each individual client.
// It is then set in the header of all outgoing requests from clients created using this config.
func NewConnectClientConfig(container appext.Container) (*connectclient.Config, error) {
	envTokenProvider, err := bufconnect2.NewTokenProviderFromContainer(container)
	if err != nil {
		return nil, err
	}
	netrcTokenProvider := bufconnect2.NewNetrcTokenProvider(container, netrc.GetMachineForName)
	return newConnectClientConfigWithOptions(
		container,
		connectclient.WithAuthInterceptorProvider(
			bufconnect2.NewAuthorizationInterceptorProvider(envTokenProvider, netrcTokenProvider),
		),
	)
}

// NewConnectClientConfigWithToken creates a new connect.ClientConfig with a given token. The provided token is
// set in the header of all outgoing requests from this provider
func NewConnectClientConfigWithToken(container appext.Container, token string) (*connectclient.Config, error) {
	tokenProvider, err := bufconnect2.NewTokenProviderFromString(token)
	if err != nil {
		return nil, err
	}
	return newConnectClientConfigWithOptions(
		container,
		connectclient.WithAuthInterceptorProvider(
			bufconnect2.NewAuthorizationInterceptorProvider(tokenProvider),
		),
	)
}

// Returns a registry provider with the given options applied in addition to default ones for all providers
func newConnectClientConfigWithOptions(container appext.Container, opts ...connectclient.ConfigOption) (*connectclient.Config, error) {
	config, err := newConfig(container)
	if err != nil {
		return nil, err
	}
	otelconnectInterceptor, err := otelconnect.NewInterceptor()
	if err != nil {
		return nil, err
	}
	client := httpclient.NewClient(config.TLS)
	options := []connectclient.ConfigOption{
		connectclient.WithAddressMapper(func(address string) string {
			if config.TLS == nil {
				return buftransport.PrependHTTP(address)
			}
			return buftransport.PrependHTTPS(address)
		}),
		connectclient.WithInterceptors(
			[]connect.Interceptor{
				bufconnect2.NewAugmentedConnectErrorInterceptor(),
				bufconnect2.NewSetCLIVersionInterceptor(Version),
				bufconnect2.NewCLIWarningInterceptor(container),
				bufconnect2.NewDebugLoggingInterceptor(container),
				otelconnectInterceptor,
			},
		),
	}
	return connectclient.NewConfig(client, append(options, opts...)...), nil
}

// newConfig creates a new Config.
func newConfig(container appext.Container) (*bufapp.Config, error) {
	externalConfig := bufapp.ExternalConfig{}
	if err := appext.ReadConfig(container, &externalConfig); err != nil {
		return nil, err
	}
	return bufapp.NewConfig(container, externalConfig)
}
