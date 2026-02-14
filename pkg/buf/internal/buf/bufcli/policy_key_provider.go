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
	"github.com/inovacc/omni/pkg/buf/internal/app/appext"
	"github.com/inovacc/omni/pkg/buf/internal/bufpkg/bufpolicy"
	"github.com/inovacc/omni/pkg/buf/internal/bufpkg/bufpolicy/bufpolicyapi"
	"github.com/inovacc/omni/pkg/buf/internal/bufpkg/bufregistryapi/bufregistryapipolicy"
)

// NewPolicyKeyProvider returns a new PolicyKeyProvider.
func NewPolicyKeyProvider(container appext.Container) (bufpolicy.PolicyKeyProvider, error) {
	clientConfig, err := NewConnectClientConfig(container)
	if err != nil {
		return nil, err
	}

	return bufpolicyapi.NewPolicyKeyProvider(
		container.Logger(),
		bufregistryapipolicy.NewClientProvider(
			clientConfig,
		),
	), nil
}
