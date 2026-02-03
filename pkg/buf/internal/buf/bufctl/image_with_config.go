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

package bufctl

import (
	bufconfig2 "github.com/inovacc/omni/pkg/buf/internal/bufpkg/bufconfig"
	"github.com/inovacc/omni/pkg/buf/internal/bufpkg/bufimage"
	"github.com/inovacc/omni/pkg/buf/internal/bufpkg/bufparse"
)

type imageWithConfig struct {
	bufimage.Image

	moduleFullName bufparse.FullName
	moduleOpaqueID string
	lintConfig     bufconfig2.LintConfig
	breakingConfig bufconfig2.BreakingConfig
	pluginConfigs  []bufconfig2.PluginConfig
	policyConfigs  []bufconfig2.PolicyConfig
}

func newImageWithConfig(
	image bufimage.Image,
	moduleFullName bufparse.FullName,
	moduleOpaqueID string,
	lintConfig bufconfig2.LintConfig,
	breakingConfig bufconfig2.BreakingConfig,
	pluginConfigs []bufconfig2.PluginConfig,
	policyConfigs []bufconfig2.PolicyConfig,
) *imageWithConfig {
	return &imageWithConfig{
		Image:          image,
		moduleFullName: moduleFullName,
		moduleOpaqueID: moduleOpaqueID,
		lintConfig:     lintConfig,
		breakingConfig: breakingConfig,
		pluginConfigs:  pluginConfigs,
		policyConfigs:  policyConfigs,
	}
}

func (i *imageWithConfig) ModuleFullName() bufparse.FullName {
	return i.moduleFullName
}

func (i *imageWithConfig) ModuleOpaqueID() string {
	return i.moduleOpaqueID
}

func (i *imageWithConfig) LintConfig() bufconfig2.LintConfig {
	return i.lintConfig
}

func (i *imageWithConfig) BreakingConfig() bufconfig2.BreakingConfig {
	return i.breakingConfig
}

func (i *imageWithConfig) PluginConfigs() []bufconfig2.PluginConfig {
	return i.pluginConfigs
}

func (i *imageWithConfig) PolicyConfigs() []bufconfig2.PolicyConfig {
	return i.policyConfigs
}

func (*imageWithConfig) isImageWithConfig() {}
