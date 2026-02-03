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

// NewCommitProvider returns a new CommitProvider that caches the results of the delegate.
//
// The CommitStore is used as a cache.
func NewCommitProvider(
	logger *slog.Logger,
	delegate bufmodule2.CommitProvider,
	store bufmodulestore.CommitStore,
) bufmodule2.CommitProvider {
	return newCommitProvider(logger, delegate, store)
}

/// *** PRIVATE ***

type commitProvider struct {
	byModuleKey *baseProvider[bufmodule2.ModuleKey, bufmodule2.Commit]
	byCommitKey *baseProvider[bufmodule2.CommitKey, bufmodule2.Commit]
}

func newCommitProvider(
	logger *slog.Logger,
	delegate bufmodule2.CommitProvider,
	store bufmodulestore.CommitStore,
) *commitProvider {
	return &commitProvider{
		byModuleKey: newBaseProvider(
			logger,
			delegate.GetCommitsForModuleKeys,
			store.GetCommitsForModuleKeys,
			store.PutCommits,
			bufmodule2.ModuleKey.CommitID,
			func(commit bufmodule2.Commit) uuid.UUID {
				return commit.ModuleKey().CommitID()
			},
		),
		byCommitKey: newBaseProvider(
			logger,
			delegate.GetCommitsForCommitKeys,
			store.GetCommitsForCommitKeys,
			store.PutCommits,
			bufmodule2.CommitKey.CommitID,
			func(commit bufmodule2.Commit) uuid.UUID {
				return commit.ModuleKey().CommitID()
			},
		),
	}
}

func (p *commitProvider) GetCommitsForModuleKeys(
	ctx context.Context,
	moduleKeys []bufmodule2.ModuleKey,
) ([]bufmodule2.Commit, error) {
	return p.byModuleKey.getValuesForKeys(ctx, moduleKeys)
}

func (p *commitProvider) GetCommitsForCommitKeys(
	ctx context.Context,
	commitKeys []bufmodule2.CommitKey,
) ([]bufmodule2.Commit, error) {
	return p.byCommitKey.getValuesForKeys(ctx, commitKeys)
}
