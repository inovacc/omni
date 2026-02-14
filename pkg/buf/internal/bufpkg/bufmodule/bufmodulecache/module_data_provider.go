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

package bufmodulecache

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	bufmodule2 "github.com/inovacc/omni/pkg/buf/internal/bufpkg/bufmodule"
	"github.com/inovacc/omni/pkg/buf/internal/bufpkg/bufmodule/bufmodulestore"
)

// NewModuleDataProvider returns a new ModuleDataProvider that caches the results of the delegate.
//
// The ModuleDataStore is used as a cache.
//
// All files in returned [storage.Bucket]s will have local paths if the cache is on-disk.
func NewModuleDataProvider(
	logger *slog.Logger,
	delegate bufmodule2.ModuleDataProvider,
	store bufmodulestore.ModuleDataStore,
) bufmodule2.ModuleDataProvider {
	return newModuleDataProvider(logger, delegate, store)
}

/// *** PRIVATE ***

type moduleDataProvider struct {
	*baseProvider[bufmodule2.ModuleKey, bufmodule2.ModuleData]
}

func newModuleDataProvider(
	logger *slog.Logger,
	delegate bufmodule2.ModuleDataProvider,
	store bufmodulestore.ModuleDataStore,
) *moduleDataProvider {
	return &moduleDataProvider{
		baseProvider: newBaseProvider(
			logger,
			delegate.GetModuleDatasForModuleKeys,
			store.GetModuleDatasForModuleKeys,
			store.PutModuleDatas,
			bufmodule2.ModuleKey.CommitID,
			func(moduleData bufmodule2.ModuleData) uuid.UUID {
				return moduleData.ModuleKey().CommitID()
			},
		),
	}
}

func (p *moduleDataProvider) GetModuleDatasForModuleKeys(
	ctx context.Context,
	moduleKeys []bufmodule2.ModuleKey,
) ([]bufmodule2.ModuleData, error) {
	return p.getValuesForKeys(ctx, moduleKeys)
}
