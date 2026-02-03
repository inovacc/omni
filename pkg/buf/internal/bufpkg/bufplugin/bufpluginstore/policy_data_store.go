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

package bufpluginstore

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"

	bufplugin2 "github.com/inovacc/omni/pkg/buf/internal/bufpkg/bufplugin"
	"github.com/inovacc/omni/pkg/buf/internal/pkg/normalpath"
	storage2 "github.com/inovacc/omni/pkg/buf/internal/pkg/storage"
	"github.com/inovacc/omni/pkg/buf/internal/pkg/uuidutil"
)

// PluginDataStore reads and writes PluginsDatas.
type PluginDataStore interface {
	// GetPluginDatasForPluginKeys gets the PluginDatas from the store for the PluginKeys.
	//
	// Returns the found PluginDatas, and the input PluginKeys that were not found, each
	// ordered by the order of the input PluginKeys.
	GetPluginDatasForPluginKeys(context.Context, []bufplugin2.PluginKey) (
		foundPluginDatas []bufplugin2.PluginData,
		notFoundPluginKeys []bufplugin2.PluginKey,
		err error,
	)
	// PutPluginDatas puts the PluginDatas to the store.
	PutPluginDatas(ctx context.Context, moduleDatas []bufplugin2.PluginData) error
}

// NewPluginDataStore returns a new PluginDataStore for the given bucket.
//
// It is assumed that the PluginDataStore has complete control of the bucket.
//
// This is typically used to interact with a cache directory.
func NewPluginDataStore(
	logger *slog.Logger,
	bucket storage2.ReadWriteBucket,
) PluginDataStore {
	return newPluginDataStore(logger, bucket)
}

/// *** PRIVATE ***

type pluginDataStore struct {
	logger *slog.Logger
	bucket storage2.ReadWriteBucket
}

func newPluginDataStore(
	logger *slog.Logger,
	bucket storage2.ReadWriteBucket,
) *pluginDataStore {
	return &pluginDataStore{
		logger: logger,
		bucket: bucket,
	}
}

func (p *pluginDataStore) GetPluginDatasForPluginKeys(
	ctx context.Context,
	pluginKeys []bufplugin2.PluginKey,
) ([]bufplugin2.PluginData, []bufplugin2.PluginKey, error) {
	var foundPluginDatas []bufplugin2.PluginData
	var notFoundPluginKeys []bufplugin2.PluginKey
	for _, pluginKey := range pluginKeys {
		pluginData, err := p.getPluginDataForPluginKey(ctx, pluginKey)
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return nil, nil, err
			}
			notFoundPluginKeys = append(notFoundPluginKeys, pluginKey)
		} else {
			foundPluginDatas = append(foundPluginDatas, pluginData)
		}
	}
	return foundPluginDatas, notFoundPluginKeys, nil
}

func (p *pluginDataStore) PutPluginDatas(
	ctx context.Context,
	pluginDatas []bufplugin2.PluginData,
) error {
	for _, pluginData := range pluginDatas {
		if err := p.putPluginData(ctx, pluginData); err != nil {
			return err
		}
	}
	return nil
}

// getPluginDataForPluginKey reads the plugin data for the plugin key from the cache.
func (p *pluginDataStore) getPluginDataForPluginKey(
	ctx context.Context,
	pluginKey bufplugin2.PluginKey,
) (bufplugin2.PluginData, error) {
	pluginDataStorePath, err := getPluginDataStorePath(pluginKey)
	if err != nil {
		return nil, err
	}
	if exists, err := storage2.Exists(ctx, p.bucket, pluginDataStorePath); err != nil {
		return nil, err
	} else if !exists {
		return nil, fs.ErrNotExist
	}
	return bufplugin2.NewPluginData(
		ctx,
		pluginKey,
		func() ([]byte, error) {
			// Data is stored uncompressed.
			return storage2.ReadPath(ctx, p.bucket, pluginDataStorePath)
		},
	)
}

// putPluginData puts the plugin data into the plugin cache.
func (p *pluginDataStore) putPluginData(
	ctx context.Context,
	pluginData bufplugin2.PluginData,
) error {
	pluginKey := pluginData.PluginKey()
	pluginDataStorePath, err := getPluginDataStorePath(pluginKey)
	if err != nil {
		return err
	}
	data, err := pluginData.Data()
	if err != nil {
		return err
	}
	// Data is stored uncompressed.
	return storage2.PutPath(ctx, p.bucket, pluginDataStorePath, data)
}

// getPluginDataStorePath returns the path for the plugin data store for the plugin key.
//
// This is "digestType/registry/owner/name/dashlessCommitID", e.g. the plugin
// "buf.build/acme/check-plugin" with commit "12345-abcde" and digest type "p1"
// will return "p1/buf.build/acme/check-plugin/12345abcde.wasm".
func getPluginDataStorePath(pluginKey bufplugin2.PluginKey) (string, error) {
	digest, err := pluginKey.Digest()
	if err != nil {
		return "", err
	}
	fullName := pluginKey.FullName()
	return normalpath.Join(
		digest.Type().String(),
		fullName.Registry(),
		fullName.Owner(),
		fullName.Name(),
		uuidutil.ToDashless(pluginKey.CommitID())+".wasm",
	), nil
}
