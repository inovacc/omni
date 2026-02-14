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

package bufpolicyconfig

import (
	"strings"

	bufconfig2 "github.com/inovacc/omni/pkg/buf/internal/bufpkg/bufconfig"
	"github.com/inovacc/omni/pkg/buf/internal/bufpkg/bufpolicy"
	"github.com/inovacc/omni/pkg/buf/internal/standard/xslices"
)

// LintConfigToBufConfig converts the given LintConfig to a bufconfig.LintConfig.
func LintConfigToBufConfig(lintConfig bufpolicy.LintConfig) (bufconfig2.LintConfig, error) {
	checkConfig, err := bufconfig2.NewEnabledCheckConfig(
		bufconfig2.FileVersionV2,
		lintConfig.UseIDsAndCategories(),
		lintConfig.ExceptIDsAndCategories(),
		nil,
		nil,
		lintConfig.DisableBuiltin(),
	)
	if err != nil {
		return nil, err
	}

	return bufconfig2.NewLintConfig(
		checkConfig,
		lintConfig.EnumZeroValueSuffix(),
		lintConfig.RPCAllowSameRequestResponse(),
		lintConfig.RPCAllowGoogleProtobufEmptyRequests(),
		lintConfig.RPCAllowGoogleProtobufEmptyResponses(),
		lintConfig.ServiceSuffix(),
		false, // Comment ignores are not allowed in Policy files.
	), nil
}

// BreakingConfigToBufConfig converts the given BreakingConfig to a bufconfig.BreakingConfig.
func BreakingConfigToBufConfig(breakingConfig bufpolicy.BreakingConfig) (bufconfig2.BreakingConfig, error) {
	checkConfig, err := bufconfig2.NewEnabledCheckConfig(
		bufconfig2.FileVersionV2,
		breakingConfig.UseIDsAndCategories(),
		breakingConfig.ExceptIDsAndCategories(),
		nil,
		nil,
		breakingConfig.DisableBuiltin(),
	)
	if err != nil {
		return nil, err
	}

	return bufconfig2.NewBreakingConfig(
		checkConfig,
		breakingConfig.IgnoreUnstablePackages(),
	), nil
}

// PluginConfigsToBufConfig converts the given plugin configs to bufconfig.PluginConfig.
func PluginConfigsToBufConfig(pluginConfigs []bufpolicy.PluginConfig) ([]bufconfig2.PluginConfig, error) {
	return xslices.MapError(pluginConfigs, func(pluginConfig bufpolicy.PluginConfig) (bufconfig2.PluginConfig, error) {
		options := make(map[string]any)

		pluginConfig.Options().Range(func(key string, value any) {
			options[key] = value
		})

		args := pluginConfig.Args()
		switch {
		case pluginConfig.Ref() != nil:
			return bufconfig2.NewRemoteWasmPluginConfig(
				pluginConfig.Ref(),
				options,
				args,
			)
		case strings.HasSuffix(pluginConfig.Name(), ".wasm"):
			return bufconfig2.NewLocalWasmPluginConfig(
				pluginConfig.Name(),
				options,
				args,
			)
		default:
			return bufconfig2.NewLocalPluginConfig(
				pluginConfig.Name(),
				options,
				args,
			)
		}
	})
}
