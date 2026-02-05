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
	"log/slog"
	"time"

	"github.com/google/uuid"
	bufmodule2 "github.com/inovacc/omni/pkg/buf/internal/bufpkg/bufmodule"
	"github.com/inovacc/omni/pkg/buf/internal/bufpkg/bufregistryapi/bufregistryapimodule"
	"github.com/inovacc/omni/pkg/buf/internal/bufpkg/bufregistryapi/bufregistryapiowner"
	"github.com/inovacc/omni/pkg/buf/internal/pkg/syserror"
	"github.com/inovacc/omni/pkg/buf/internal/pkg/uuidutil"
	"github.com/inovacc/omni/pkg/buf/internal/standard/xslices"
)

// NewCommitProvider returns a new CommitProvider for the given API client.
func NewCommitProvider(
	logger *slog.Logger,
	moduleClientProvider interface {
		bufregistryapimodule.V1CommitServiceClientProvider
		bufregistryapimodule.V1ModuleServiceClientProvider
		bufregistryapimodule.V1Beta1CommitServiceClientProvider
	},
	ownerClientProvider bufregistryapiowner.V1OwnerServiceClientProvider,
) bufmodule2.CommitProvider {
	return newCommitProvider(logger, moduleClientProvider, ownerClientProvider)
}

// *** PRIVATE ***

type commitProvider struct {
	logger               *slog.Logger
	moduleClientProvider interface {
		bufregistryapimodule.V1CommitServiceClientProvider
		bufregistryapimodule.V1ModuleServiceClientProvider
		bufregistryapimodule.V1Beta1CommitServiceClientProvider
	}
	ownerClientProvider bufregistryapiowner.V1OwnerServiceClientProvider
}

func newCommitProvider(
	logger *slog.Logger,
	moduleClientProvider interface {
		bufregistryapimodule.V1CommitServiceClientProvider
		bufregistryapimodule.V1ModuleServiceClientProvider
		bufregistryapimodule.V1Beta1CommitServiceClientProvider
	},
	ownerClientProvider bufregistryapiowner.V1OwnerServiceClientProvider,
) *commitProvider {
	return &commitProvider{
		logger:               logger,
		moduleClientProvider: moduleClientProvider,
		ownerClientProvider:  ownerClientProvider,
	}
}

func (a *commitProvider) GetCommitsForModuleKeys(
	ctx context.Context,
	moduleKeys []bufmodule2.ModuleKey,
) ([]bufmodule2.Commit, error) {
	if len(moduleKeys) == 0 {
		return nil, nil
	}
	digestType, err := bufmodule2.UniqueDigestTypeForModuleKeys(moduleKeys)
	if err != nil {
		return nil, err
	}

	registryToIndexedModuleKeys := xslices.ToIndexedValuesMap(
		moduleKeys,
		func(moduleKey bufmodule2.ModuleKey) string {
			return moduleKey.FullName().Registry()
		},
	)
	indexedCommits := make([]xslices.Indexed[bufmodule2.Commit], 0, len(moduleKeys))
	for registry, indexedModuleKeys := range registryToIndexedModuleKeys {
		registryIndexedCommits, err := a.getIndexedCommitsForRegistryAndIndexedModuleKeys(
			ctx,
			registry,
			indexedModuleKeys,
			digestType,
		)
		if err != nil {
			return nil, err
		}
		indexedCommits = append(indexedCommits, registryIndexedCommits...)
	}
	return xslices.IndexedToSortedValues(indexedCommits), nil
}

func (a *commitProvider) GetCommitsForCommitKeys(
	ctx context.Context,
	commitKeys []bufmodule2.CommitKey,
) ([]bufmodule2.Commit, error) {
	if len(commitKeys) == 0 {
		return nil, nil
	}
	digestType, err := bufmodule2.UniqueDigestTypeForCommitKeys(commitKeys)
	if err != nil {
		return nil, err
	}

	// We don't want to persist these across calls - this could grow over time and this cache
	// isn't an LRU cache, and the information also may change over time.
	v1ProtoModuleProvider := newV1ProtoModuleProvider(a.logger, a.moduleClientProvider)
	v1ProtoOwnerProvider := newV1ProtoOwnerProvider(a.logger, a.ownerClientProvider)

	registryToIndexedCommitKeys := xslices.ToIndexedValuesMap(
		commitKeys,
		func(commitKey bufmodule2.CommitKey) string {
			return commitKey.Registry()
		},
	)
	indexedCommits := make([]xslices.Indexed[bufmodule2.Commit], 0, len(commitKeys))
	for registry, indexedCommitKeys := range registryToIndexedCommitKeys {
		registryIndexedCommits, err := a.getIndexedCommitsForRegistryAndIndexedCommitKeys(
			ctx,
			v1ProtoModuleProvider,
			v1ProtoOwnerProvider,
			registry,
			indexedCommitKeys,
			digestType,
		)
		if err != nil {
			return nil, err
		}
		indexedCommits = append(indexedCommits, registryIndexedCommits...)
	}
	return xslices.IndexedToSortedValues(indexedCommits), nil
}

func (a *commitProvider) getIndexedCommitsForRegistryAndIndexedModuleKeys(
	ctx context.Context,
	registry string,
	indexedModuleKeys []xslices.Indexed[bufmodule2.ModuleKey],
	digestType bufmodule2.DigestType,
) ([]xslices.Indexed[bufmodule2.Commit], error) {
	commitIDToIndexedModuleKey, err := xslices.ToUniqueValuesMapError(
		indexedModuleKeys,
		func(indexedModuleKey xslices.Indexed[bufmodule2.ModuleKey]) (uuid.UUID, error) {
			return indexedModuleKey.Value.CommitID(), nil
		},
	)
	if err != nil {
		return nil, err
	}
	commitIDs := xslices.MapKeysToSlice(commitIDToIndexedModuleKey)
	universalProtoCommits, err := getUniversalProtoCommitsForRegistryAndCommitIDs(ctx, a.moduleClientProvider, registry, commitIDs, digestType)
	if err != nil {
		return nil, err
	}
	return xslices.MapError(
		universalProtoCommits,
		func(universalProtoCommit *universalProtoCommit) (xslices.Indexed[bufmodule2.Commit], error) {
			commitID, err := uuidutil.FromDashless(universalProtoCommit.ID)
			if err != nil {
				return xslices.Indexed[bufmodule2.Commit]{}, err
			}
			indexedModuleKey, ok := commitIDToIndexedModuleKey[commitID]
			if !ok {
				return xslices.Indexed[bufmodule2.Commit]{}, syserror.Newf("no ModuleKey for proto commit ID %q", commitID)
			}
			// This is actually backwards - this is not the expected digest, this is the actual digest.
			// TODO FUTURE: It doesn't matter too much, but we should switch around CommitWithExpectedDigest
			// to be CommitWithActualDigest.
			expectedDigest := universalProtoCommit.Digest
			return xslices.Indexed[bufmodule2.Commit]{
				Value: bufmodule2.NewCommit(
					indexedModuleKey.Value,
					func() (time.Time, error) {
						return universalProtoCommit.CreateTime, nil
					},
					bufmodule2.CommitWithExpectedDigest(expectedDigest),
				),
				Index: indexedModuleKey.Index,
			}, nil
		},
	)
}

func (a *commitProvider) getIndexedCommitsForRegistryAndIndexedCommitKeys(
	ctx context.Context,
	v1ProtoModuleProvider *v1ProtoModuleProvider,
	v1ProtoOwnerProvider *v1ProtoOwnerProvider,
	registry string,
	indexedCommitKeys []xslices.Indexed[bufmodule2.CommitKey],
	digestType bufmodule2.DigestType,
) ([]xslices.Indexed[bufmodule2.Commit], error) {
	commitIDToIndexedCommitKey, err := xslices.ToUniqueValuesMapError(
		indexedCommitKeys,
		func(indexedCommitKey xslices.Indexed[bufmodule2.CommitKey]) (uuid.UUID, error) {
			return indexedCommitKey.Value.CommitID(), nil
		},
	)
	if err != nil {
		return nil, err
	}
	commitIDs := xslices.MapKeysToSlice(commitIDToIndexedCommitKey)
	universalProtoCommits, err := getUniversalProtoCommitsForRegistryAndCommitIDs(ctx, a.moduleClientProvider, registry, commitIDs, digestType)
	if err != nil {
		return nil, err
	}
	return xslices.MapError(
		universalProtoCommits,
		func(universalProtoCommit *universalProtoCommit) (xslices.Indexed[bufmodule2.Commit], error) {
			commitID, err := uuidutil.FromDashless(universalProtoCommit.ID)
			if err != nil {
				return xslices.Indexed[bufmodule2.Commit]{}, err
			}
			indexedCommitKey, ok := commitIDToIndexedCommitKey[commitID]
			if !ok {
				return xslices.Indexed[bufmodule2.Commit]{}, syserror.Newf("no CommitKey for proto commit ID %q", commitID)
			}
			moduleKey, err := getModuleKeyForUniversalProtoCommit(
				ctx,
				v1ProtoModuleProvider,
				v1ProtoOwnerProvider,
				registry,
				universalProtoCommit,
			)
			if err != nil {
				return xslices.Indexed[bufmodule2.Commit]{}, err
			}
			return xslices.Indexed[bufmodule2.Commit]{
				// No digest to compare against to add as CommitOption.
				Value: bufmodule2.NewCommit(
					moduleKey,
					func() (time.Time, error) {
						return universalProtoCommit.CreateTime, nil
					},
				),
				Index: indexedCommitKey.Index,
			}, nil
		},
	)
}
