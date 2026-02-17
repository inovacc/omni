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
	"context"
	"fmt"

	"github.com/inovacc/omni/pkg/buf/internal/buf/bufmodule"
	"github.com/inovacc/omni/pkg/buf/internal/buf/bufplugin"
	"github.com/inovacc/omni/pkg/buf/internal/buf/bufpolicy"
)

// offlineModuleDataProvider is a ModuleDataProvider that always returns an error,
// used as the delegate for cache providers in offline mode.
type offlineModuleDataProvider struct{}

func (offlineModuleDataProvider) GetModuleDatasForModuleKeys(
	_ context.Context,
	moduleKeys []bufmodule.ModuleKey,
) ([]bufmodule.ModuleData, error) {
	if len(moduleKeys) == 0 {
		return nil, nil
	}
	return nil, fmt.Errorf(
		"module %q is not available in the local cache; run 'buf dep update' with network access first",
		moduleKeys[0].FullName(),
	)
}

// offlineCommitProvider is a CommitProvider that always returns an error,
// used as the delegate for cache providers in offline mode.
type offlineCommitProvider struct{}

func (offlineCommitProvider) GetCommitsForModuleKeys(
	_ context.Context,
	moduleKeys []bufmodule.ModuleKey,
) ([]bufmodule.Commit, error) {
	if len(moduleKeys) == 0 {
		return nil, nil
	}
	return nil, fmt.Errorf(
		"commit for module %q is not available in the local cache; run 'buf dep update' with network access first",
		moduleKeys[0].FullName(),
	)
}

func (offlineCommitProvider) GetCommitsForCommitKeys(
	_ context.Context,
	commitKeys []bufmodule.CommitKey,
) ([]bufmodule.Commit, error) {
	if len(commitKeys) == 0 {
		return nil, nil
	}
	return nil, fmt.Errorf(
		"commit %q is not available in the local cache; run 'buf dep update' with network access first",
		commitKeys[0].CommitID(),
	)
}

// offlinePluginDataProvider is a PluginDataProvider that always returns an error,
// used as the delegate for cache providers in offline mode.
type offlinePluginDataProvider struct{}

func (offlinePluginDataProvider) GetPluginDatasForPluginKeys(
	_ context.Context,
	pluginKeys []bufplugin.PluginKey,
) ([]bufplugin.PluginData, error) {
	if len(pluginKeys) == 0 {
		return nil, nil
	}
	return nil, fmt.Errorf(
		"plugin %q is not available in the local cache; run 'buf dep update' with network access first",
		pluginKeys[0].FullName(),
	)
}

// offlinePolicyDataProvider is a PolicyDataProvider that always returns an error,
// used as the delegate for cache providers in offline mode.
type offlinePolicyDataProvider struct{}

func (offlinePolicyDataProvider) GetPolicyDatasForPolicyKeys(
	_ context.Context,
	policyKeys []bufpolicy.PolicyKey,
) ([]bufpolicy.PolicyData, error) {
	if len(policyKeys) == 0 {
		return nil, nil
	}
	return nil, fmt.Errorf(
		"policy %q is not available in the local cache; run 'buf dep update' with network access first",
		policyKeys[0].FullName(),
	)
}
