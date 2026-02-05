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

package bufmoduleapi

import (
	"context"
	"fmt"
	"log/slog"
	"sort"

	"github.com/google/uuid"
	bufmodule2 "github.com/inovacc/omni/pkg/buf/internal/bufpkg/bufmodule"
	"github.com/inovacc/omni/pkg/buf/internal/bufpkg/bufparse"
	"github.com/inovacc/omni/pkg/buf/internal/bufpkg/bufregistryapi/bufregistryapimodule"
	modulev1 "github.com/inovacc/omni/pkg/buf/internal/gen/bufbuild/registry/protocolbuffers/go/buf/registry/module/v1"
	"github.com/inovacc/omni/pkg/buf/internal/pkg/storage"
	"github.com/inovacc/omni/pkg/buf/internal/pkg/syserror"
	"github.com/inovacc/omni/pkg/buf/internal/pkg/uuidutil"
	"github.com/inovacc/omni/pkg/buf/internal/standard/xslices"
)

// NewModuleDataProvider returns a new ModuleDataProvider for the given API client.
//
// A warning is printed to the logger if a given Module is deprecated.
func NewModuleDataProvider(
	logger *slog.Logger,
	moduleClientProvider interface {
		bufregistryapimodule.V1DownloadServiceClientProvider
		bufregistryapimodule.V1ModuleServiceClientProvider
		bufregistryapimodule.V1Beta1DownloadServiceClientProvider
	},
	graphProvider bufmodule2.GraphProvider,
) bufmodule2.ModuleDataProvider {
	return newModuleDataProvider(logger, moduleClientProvider, graphProvider)
}

// *** PRIVATE ***

type moduleDataProvider struct {
	logger               *slog.Logger
	moduleClientProvider interface {
		bufregistryapimodule.V1DownloadServiceClientProvider
		bufregistryapimodule.V1ModuleServiceClientProvider
		bufregistryapimodule.V1Beta1DownloadServiceClientProvider
	}
	graphProvider bufmodule2.GraphProvider
}

func newModuleDataProvider(
	logger *slog.Logger,
	moduleClientProvider interface {
		bufregistryapimodule.V1DownloadServiceClientProvider
		bufregistryapimodule.V1ModuleServiceClientProvider
		bufregistryapimodule.V1Beta1DownloadServiceClientProvider
	},
	graphProvider bufmodule2.GraphProvider,
) *moduleDataProvider {
	return &moduleDataProvider{
		logger:               logger,
		moduleClientProvider: moduleClientProvider,
		graphProvider:        graphProvider,
	}
}

func (a *moduleDataProvider) GetModuleDatasForModuleKeys(
	ctx context.Context,
	moduleKeys []bufmodule2.ModuleKey,
) ([]bufmodule2.ModuleData, error) {
	if len(moduleKeys) == 0 {
		return nil, nil
	}
	digestType, err := bufmodule2.UniqueDigestTypeForModuleKeys(moduleKeys)
	if err != nil {
		return nil, err
	}
	if _, err := bufparse.FullNameStringToUniqueValue(moduleKeys); err != nil {
		return nil, err
	}

	// We don't want to persist this across calls - this could grow over time and this cache
	// isn't an LRU cache, and the information also may change over time.
	v1ProtoModuleProvider := newV1ProtoModuleProvider(a.logger, a.moduleClientProvider)

	registryToIndexedModuleKeys := xslices.ToIndexedValuesMap(
		moduleKeys,
		func(moduleKey bufmodule2.ModuleKey) string {
			return moduleKey.FullName().Registry()
		},
	)
	indexedModuleDatas := make([]xslices.Indexed[bufmodule2.ModuleData], 0, len(moduleKeys))
	for registry, indexedModuleKeys := range registryToIndexedModuleKeys {
		// registryModuleDatas are in the same order as indexedModuleKeys.
		indexedRegistryModuleDatas, err := a.getIndexedModuleDatasForRegistryAndIndexedModuleKeys(
			ctx,
			v1ProtoModuleProvider,
			registry,
			indexedModuleKeys,
			digestType,
		)
		if err != nil {
			return nil, err
		}
		indexedModuleDatas = append(indexedModuleDatas, indexedRegistryModuleDatas...)
	}
	return xslices.IndexedToSortedValues(indexedModuleDatas), nil
}

// Returns ModuleDatas in the same order as the input ModuleKeys
func (a *moduleDataProvider) getIndexedModuleDatasForRegistryAndIndexedModuleKeys(
	ctx context.Context,
	v1ProtoModuleProvider *v1ProtoModuleProvider,
	registry string,
	indexedModuleKeys []xslices.Indexed[bufmodule2.ModuleKey],
	digestType bufmodule2.DigestType,
) ([]xslices.Indexed[bufmodule2.ModuleData], error) {
	graph, err := a.graphProvider.GetGraphForModuleKeys(ctx, xslices.IndexedToValues(indexedModuleKeys))
	if err != nil {
		return nil, err
	}
	commitIDToIndexedModuleKey, err := xslices.ToUniqueValuesMapError(
		indexedModuleKeys,
		func(indexedModuleKey xslices.Indexed[bufmodule2.ModuleKey]) (uuid.UUID, error) {
			return indexedModuleKey.Value.CommitID(), nil
		},
	)
	if err != nil {
		return nil, err
	}
	commitIDToUniversalProtoContent, err := a.getCommitIDToUniversalProtoContentForRegistryAndIndexedModuleKeys(
		ctx,
		v1ProtoModuleProvider,
		registry,
		commitIDToIndexedModuleKey,
		digestType,
	)
	if err != nil {
		return nil, err
	}
	indexedModuleDatas := make([]xslices.Indexed[bufmodule2.ModuleData], 0, len(indexedModuleKeys))
	for _, indexedModuleKey := range indexedModuleKeys {
		moduleKey := indexedModuleKey.Value
		// TopoSort will get us both the direct and transitive dependencies for the key.
		depModuleKeys, err := graph.TopoSort(bufmodule2.ModuleKeyToRegistryCommitID(moduleKey))
		if err != nil {
			return nil, err
		}
		// Remove this moduleKey from the depModuleKeys.
		depModuleKeys = depModuleKeys[:len(depModuleKeys)-1]
		sort.Slice(
			depModuleKeys,
			func(i int, j int) bool {
				return depModuleKeys[i].FullName().String() < depModuleKeys[j].FullName().String()
			},
		)

		universalProtoContent, ok := commitIDToUniversalProtoContent[moduleKey.CommitID()]
		if !ok {
			return nil, syserror.Newf("could not find universalProtoContent for commit ID %q", moduleKey.CommitID())
		}
		indexedModuleData := xslices.Indexed[bufmodule2.ModuleData]{
			Value: bufmodule2.NewModuleData(
				ctx,
				moduleKey,
				func() (storage.ReadBucket, error) {
					return universalProtoFilesToBucket(universalProtoContent.Files)
				},
				func() ([]bufmodule2.ModuleKey, error) { return depModuleKeys, nil },
				func() (bufmodule2.ObjectData, error) {
					return universalProtoFileToObjectData(universalProtoContent.V1BufYAMLFile)
				},
				func() (bufmodule2.ObjectData, error) {
					return universalProtoFileToObjectData(universalProtoContent.V1BufLockFile)
				},
			),
			Index: indexedModuleKey.Index,
		}
		indexedModuleDatas = append(indexedModuleDatas, indexedModuleData)
	}
	return indexedModuleDatas, nil
}

func (a *moduleDataProvider) getCommitIDToUniversalProtoContentForRegistryAndIndexedModuleKeys(
	ctx context.Context,
	v1ProtoModuleProvider *v1ProtoModuleProvider,
	registry string,
	commitIDToIndexedModuleKey map[uuid.UUID]xslices.Indexed[bufmodule2.ModuleKey],
	digestType bufmodule2.DigestType,
) (map[uuid.UUID]*universalProtoContent, error) {
	commitIDs := xslices.MapKeysToSlice(commitIDToIndexedModuleKey)
	universalProtoContents, err := getUniversalProtoContentsForRegistryAndCommitIDs(
		ctx,
		a.moduleClientProvider,
		registry,
		commitIDs,
		digestType,
	)
	if err != nil {
		return nil, err
	}
	commitIDToUniversalProtoContent, err := xslices.ToUniqueValuesMapError(
		universalProtoContents,
		func(universalProtoContent *universalProtoContent) (uuid.UUID, error) {
			return uuidutil.FromDashless(universalProtoContent.CommitID)
		},
	)
	if err != nil {
		return nil, err
	}
	for commitID, indexedModuleKey := range commitIDToIndexedModuleKey {
		universalProtoContent, ok := commitIDToUniversalProtoContent[commitID]
		if !ok {
			return nil, fmt.Errorf("no content returned for commit ID %s", commitID)
		}
		if err := a.warnIfDeprecated(
			ctx,
			v1ProtoModuleProvider,
			registry,
			universalProtoContent.ModuleID,
			indexedModuleKey.Value,
		); err != nil {
			return nil, err
		}
	}
	return commitIDToUniversalProtoContent, nil
}

// In the future, we might want to add State, Visibility, etc as parameters to bufmodule.Module, to
// match what we are doing with Commit and Graph to some degree, and then bring this warning
// out of the ModuleDataProvider. However, if we did this, this has unintended consequences - right now,
// by this being here, we only warn when we don't have the module in the cache, which we sort of want?
// State is a property only on the BSR, it's not a property on a per-commit basis, so this gets into
// weird territory.
func (a *moduleDataProvider) warnIfDeprecated(
	ctx context.Context,
	v1ProtoModuleProvider *v1ProtoModuleProvider,
	registry string,
	// Dashless
	protoModuleID string,
	moduleKey bufmodule2.ModuleKey,
) error {
	v1ProtoModule, err := v1ProtoModuleProvider.getV1ProtoModuleForProtoModuleID(
		ctx,
		registry,
		protoModuleID,
	)
	if err != nil {
		return err
	}
	if v1ProtoModule.State == modulev1.ModuleState_MODULE_STATE_DEPRECATED {
		a.logger.Warn(fmt.Sprintf("%s is deprecated", moduleKey.FullName().String()))
	}
	return nil
}
