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

package bufplugincache

import (
	"context"
	"log/slog"
	"sync/atomic"

	"github.com/google/uuid"
	bufplugin2 "github.com/inovacc/omni/pkg/buf/internal/bufpkg/bufplugin"
	"github.com/inovacc/omni/pkg/buf/internal/bufpkg/bufplugin/bufpluginstore"
	"github.com/inovacc/omni/pkg/buf/internal/pkg/syserror"
	"github.com/inovacc/omni/pkg/buf/internal/pkg/uuidutil"
	"github.com/inovacc/omni/pkg/buf/internal/standard/xslices"
)

// NewPluginDataProvider returns a new PluginDataProvider that caches the results of the delegate.
//
// The PluginDataStore is used as a cache.
func NewPluginDataProvider(
	logger *slog.Logger,
	delegate bufplugin2.PluginDataProvider,
	store bufpluginstore.PluginDataStore,
) bufplugin2.PluginDataProvider {
	return newPluginDataProvider(logger, delegate, store)
}

/// *** PRIVATE ***

type pluginDataProvider struct {
	logger   *slog.Logger
	delegate bufplugin2.PluginDataProvider
	store    bufpluginstore.PluginDataStore

	keysRetrieved atomic.Int64
	keysHit       atomic.Int64
}

func newPluginDataProvider(
	logger *slog.Logger,
	delegate bufplugin2.PluginDataProvider,
	store bufpluginstore.PluginDataStore,
) *pluginDataProvider {
	return &pluginDataProvider{
		logger:   logger,
		delegate: delegate,
		store:    store,
	}
}

func (p *pluginDataProvider) GetPluginDatasForPluginKeys(
	ctx context.Context,
	pluginKeys []bufplugin2.PluginKey,
) ([]bufplugin2.PluginData, error) {
	foundValues, notFoundKeys, err := p.store.GetPluginDatasForPluginKeys(ctx, pluginKeys)
	if err != nil {
		return nil, err
	}

	delegateValues, err := p.delegate.GetPluginDatasForPluginKeys(ctx, notFoundKeys)
	if err != nil {
		return nil, err
	}

	if err := p.store.PutPluginDatas(ctx, delegateValues); err != nil {
		return nil, err
	}

	p.keysRetrieved.Add(int64(len(pluginKeys)))
	p.keysHit.Add(int64(len(foundValues)))

	commitIDToIndexedKey, err := xslices.ToUniqueIndexedValuesMap(
		pluginKeys,
		func(pluginKey bufplugin2.PluginKey) uuid.UUID {
			return pluginKey.CommitID()
		},
	)
	if err != nil {
		return nil, err
	}

	indexedValues, err := xslices.MapError(
		append(foundValues, delegateValues...),
		func(value bufplugin2.PluginData) (xslices.Indexed[bufplugin2.PluginData], error) {
			commitID := value.PluginKey().CommitID()

			indexedKey, ok := commitIDToIndexedKey[commitID]
			if !ok {
				return xslices.Indexed[bufplugin2.PluginData]{}, syserror.Newf("did not get value from store with commitID %q", uuidutil.ToDashless(commitID))
			}

			return xslices.Indexed[bufplugin2.PluginData]{
				Value: value,
				Index: indexedKey.Index,
			}, nil
		},
	)
	if err != nil {
		return nil, err
	}

	return xslices.IndexedToSortedValues(indexedValues), nil
}
