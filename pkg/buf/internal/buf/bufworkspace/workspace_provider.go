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

package bufworkspace

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"

	"buf.build/go/standard/xlog/xslog"
	"buf.build/go/standard/xslices"
	"buf.build/go/standard/xstrings"
	"github.com/google/uuid"
	"github.com/inovacc/omni/pkg/buf/internal/buf/buftarget"
	bufconfig2 "github.com/inovacc/omni/pkg/buf/internal/bufpkg/bufconfig"
	bufmodule2 "github.com/inovacc/omni/pkg/buf/internal/bufpkg/bufmodule"
	bufparse2 "github.com/inovacc/omni/pkg/buf/internal/bufpkg/bufparse"
	bufplugin2 "github.com/inovacc/omni/pkg/buf/internal/bufpkg/bufplugin"
	"github.com/inovacc/omni/pkg/buf/internal/bufpkg/bufpolicy"
	"github.com/inovacc/omni/pkg/buf/internal/pkg/normalpath"
	"github.com/inovacc/omni/pkg/buf/internal/pkg/storage"
	"github.com/inovacc/omni/pkg/buf/internal/pkg/syserror"
)

// WorkspaceProvider provides Workspaces and UpdateableWorkspaces.
type WorkspaceProvider interface {
	// GetWorkspaceForBucket returns a new Workspace for the given Bucket.
	//
	// If the underlying bucket has a v2 buf.yaml at the root, this builds a Workspace for this buf.yaml,
	// using TargetSubDirPath for targeting.
	//
	// If the underlying bucket has a buf.work.yaml at the root, this builds a Workspace with all the modules
	// specified in the buf.work.yaml, using TargetSubDirPath for targeting.
	//
	// Otherwise, this builds a Workspace with a single module at the TargetSubDirPath (which may be "."),
	// assuming v1 defaults.
	//
	// If a config override is specified, all buf.work.yamls are ignored. If the config override is v1,
	// this builds a single module at the TargetSubDirPath, if the config override is v2, the builds
	// at the root, using TargetSubDirPath for targeting.
	//
	// All parsing of configuration files is done behind the scenes here.
	GetWorkspaceForBucket(
		ctx context.Context,
		bucket storage.ReadBucket,
		bucketTargeting buftarget.BucketTargeting,
		options ...WorkspaceBucketOption,
	) (Workspace, error)

	// GetWorkspaceForModuleKey wraps the ModuleKey into a workspace, returning defaults
	// for config values, and empty ConfiguredDepModuleRefs.
	//
	// This is useful for getting Workspaces for remote modules, but you still need
	// associated configuration.
	GetWorkspaceForModuleKey(
		ctx context.Context,
		moduleKey bufmodule2.ModuleKey,
		options ...WorkspaceModuleKeyOption,
	) (Workspace, error)
}

// NewWorkspaceProvider returns a new WorkspaceProvider.
func NewWorkspaceProvider(
	logger *slog.Logger,
	graphProvider bufmodule2.GraphProvider,
	moduleDataProvider bufmodule2.ModuleDataProvider,
	commitProvider bufmodule2.CommitProvider,
	pluginKeyProvider bufplugin2.PluginKeyProvider,
) WorkspaceProvider {
	return newWorkspaceProvider(
		logger,
		graphProvider,
		moduleDataProvider,
		commitProvider,
		pluginKeyProvider,
	)
}

// *** PRIVATE ***

type workspaceProvider struct {
	logger             *slog.Logger
	graphProvider      bufmodule2.GraphProvider
	moduleDataProvider bufmodule2.ModuleDataProvider
	commitProvider     bufmodule2.CommitProvider

	// pluginKeyProvider is only used for getting remote plugin keys for a single module
	// when an override is specified.
	pluginKeyProvider bufplugin2.PluginKeyProvider
}

func newWorkspaceProvider(
	logger *slog.Logger,
	graphProvider bufmodule2.GraphProvider,
	moduleDataProvider bufmodule2.ModuleDataProvider,
	commitProvider bufmodule2.CommitProvider,
	pluginKeyProvider bufplugin2.PluginKeyProvider,
) *workspaceProvider {
	return &workspaceProvider{
		logger:             logger,
		graphProvider:      graphProvider,
		moduleDataProvider: moduleDataProvider,
		commitProvider:     commitProvider,
		pluginKeyProvider:  pluginKeyProvider,
	}
}

func (w *workspaceProvider) GetWorkspaceForModuleKey(
	ctx context.Context,
	moduleKey bufmodule2.ModuleKey,
	options ...WorkspaceModuleKeyOption,
) (Workspace, error) {
	defer xslog.DebugProfile(w.logger)()

	config, err := newWorkspaceModuleKeyConfig(options)
	if err != nil {
		return nil, err
	}
	// By default, the associated configuration for a Module gotten by ModuleKey is just
	// the default config. However, if we have a config override, we may have different
	// lint or breaking config. We will only apply this different config for the specific
	// module we are targeting, while the rest will retain the default config - generally,
	// you shouldn't be linting or doing breaking change detection for any module other
	// than the one your are targeting (which matches v1 behavior as well). In v1, we didn't
	// have a "workspace" for modules gotten by module reference, we just had the single
	// module we were building against, and whatever config override we had only applied
	// to that module. In v2, we have a ModuleSet, and we need lint and breaking config
	// for every modules in the ModuleSet, so we attach default lint and breaking config,
	// but given the presence of ignore_only, we don't want to apply configOverride to
	// non-target modules as the config override might have file paths, and we won't
	// lint or breaking change detect against non-target modules anyways.
	targetModuleConfig := bufconfig2.DefaultModuleConfigV1
	// By default, there will be no plugin configs, however, similar to the lint and breaking
	// configs, there may be an override, in which case, we need to populate the plugin configs
	// from the override. Any remote plugin refs will be resolved by the pluginKeyProvider.
	var (
		pluginConfigs    []bufconfig2.PluginConfig
		remotePluginKeys []bufplugin2.PluginKey
	)
	if config.configOverride != "" {
		bufYAMLFile, err := bufconfig2.GetBufYAMLFileForOverride(config.configOverride)
		if err != nil {
			return nil, err
		}
		moduleConfigs := bufYAMLFile.ModuleConfigs()
		switch len(moduleConfigs) {
		case 0:
			return nil, syserror.New("had BufYAMLFile with 0 ModuleConfigs")
		case 1:
			// If we have a single ModuleConfig, we assume that regardless of whether or not
			// This ModuleConfig has a name, that this is what the user intends to associate
			// with the target module. This also handles the v1 case - v1 buf.yamls will always
			// only have a single ModuleConfig, and it was expected pre-refactor that regardless
			// of if the ModuleConfig had a name associated with it or not, the lint and breaking
			// config that came from it would be associated.
			targetModuleConfig = moduleConfigs[0]
		default:
			// If we have more than one ModuleConfig, find the ModuleConfig that matches the
			// name from the ModuleKey. If none is found, just fall back to the default (ie do nothing here).
			for _, moduleConfig := range moduleConfigs {
				moduleFullName := moduleConfig.FullName()
				if moduleFullName == nil {
					continue
				}
				if bufparse2.FullNameEqual(moduleFullName, moduleKey.FullName()) {
					targetModuleConfig = moduleConfig
					// We know that the ModuleConfigs are unique by FullName.
					break
				}
			}
		}
		if bufYAMLFile.FileVersion() == bufconfig2.FileVersionV2 {
			pluginConfigs = bufYAMLFile.PluginConfigs()
			// To support remote plugins when using a config override, we need to resolve the remote
			// Refs to PluginKeys. We use the pluginKeyProvider to resolve any remote plugin Refs.
			remotePluginRefs := xslices.Filter(
				xslices.Map(pluginConfigs, func(pluginConfig bufconfig2.PluginConfig) bufparse2.Ref {
					return pluginConfig.Ref()
				}),
				func(ref bufparse2.Ref) bool {
					return ref != nil
				},
			)
			if len(remotePluginRefs) > 0 {
				var err error
				remotePluginKeys, err = w.pluginKeyProvider.GetPluginKeysForPluginRefs(
					ctx,
					remotePluginRefs,
					bufplugin2.DigestTypeP1,
				)
				if err != nil {
					return nil, err
				}
			}
			// TODO: Supporting Policies for ModuleKeys requires resolving each policies remote
			// plugins.  This is not currently supported. Instead of requiring the resolved Keys for
			// the workspace, passing resolvers for the remote refs would allow for lazy resolution.
			policyConfigs := bufYAMLFile.PolicyConfigs()
			if len(policyConfigs) > 0 {
				return nil, fmt.Errorf("policies are not supported for ModuleKeys")
			}
		}
	}

	moduleSet, err := bufmodule2.NewModuleSetForRemoteModule(
		ctx,
		w.logger,
		w.graphProvider,
		w.moduleDataProvider,
		w.commitProvider,
		moduleKey,
		bufmodule2.RemoteModuleWithTargetPaths(
			config.targetPaths,
			config.targetExcludePaths,
		),
	)
	if err != nil {
		return nil, err
	}

	opaqueIDToLintConfig := make(map[string]bufconfig2.LintConfig)
	opaqueIDToBreakingConfig := make(map[string]bufconfig2.BreakingConfig)
	for _, module := range moduleSet.Modules() {
		if bufparse2.FullNameEqual(module.FullName(), moduleKey.FullName()) {
			// Set the lint and breaking config for the single targeted Module.
			opaqueIDToLintConfig[module.OpaqueID()] = targetModuleConfig.LintConfig()
			opaqueIDToBreakingConfig[module.OpaqueID()] = targetModuleConfig.BreakingConfig()
		} else {
			// For all non-targets, set the default lint and breaking config.
			opaqueIDToLintConfig[module.OpaqueID()] = bufconfig2.DefaultLintConfigV1
			opaqueIDToBreakingConfig[module.OpaqueID()] = bufconfig2.DefaultBreakingConfigV1
		}
	}
	return newWorkspace(
		moduleSet,
		opaqueIDToLintConfig,
		opaqueIDToBreakingConfig,
		pluginConfigs,
		remotePluginKeys,
		nil, // PolicyConfigs are not supported for ModuleKeys
		nil,
		nil,
		nil,
		false,
	), nil
}

func (w *workspaceProvider) getWorkspaceTargetingForBucket(
	ctx context.Context,
	bucket storage.ReadBucket,
	bucketTargeting buftarget.BucketTargeting,
	options ...WorkspaceBucketOption,
) (*workspaceTargeting, error) {
	config, err := newWorkspaceBucketConfig(options)
	if err != nil {
		return nil, err
	}
	var overrideBufYAMLFile bufconfig2.BufYAMLFile
	if config.configOverride != "" {
		overrideBufYAMLFile, err = bufconfig2.GetBufYAMLFileForOverride(config.configOverride)
		if err != nil {
			return nil, err
		}
	}
	return newWorkspaceTargeting(
		ctx,
		w.logger,
		config,
		bucket,
		bucketTargeting,
		overrideBufYAMLFile,
		config.ignoreAndDisallowV1BufWorkYAMLs,
	)
}

func (w *workspaceProvider) GetWorkspaceForBucket(
	ctx context.Context,
	bucket storage.ReadBucket,
	bucketTargeting buftarget.BucketTargeting,
	options ...WorkspaceBucketOption,
) (Workspace, error) {
	defer xslog.DebugProfile(w.logger)()
	workspaceTargeting, err := w.getWorkspaceTargetingForBucket(
		ctx,
		bucket,
		bucketTargeting,
		options...,
	)
	if err != nil {
		return nil, err
	}
	if workspaceTargeting.v2 != nil {
		return w.getWorkspaceForBucketBufYAMLV2(
			ctx,
			bucket,
			workspaceTargeting.v2,
		)
	}
	return w.getWorkspaceForBucketAndModuleDirPathsV1Beta1OrV1(
		ctx,
		bucket,
		workspaceTargeting.v1,
	)
}

func (w *workspaceProvider) getWorkspaceForBucketAndModuleDirPathsV1Beta1OrV1(
	ctx context.Context,
	bucket storage.ReadBucket,
	v1WorkspaceTargeting *v1Targeting,
) (*workspace, error) {
	moduleSetBuilder := bufmodule2.NewModuleSetBuilder(ctx, w.logger, w.moduleDataProvider, w.commitProvider)
	for _, moduleBucketAndTargeting := range v1WorkspaceTargeting.moduleBucketsAndTargeting {
		mappedModuleBucket := moduleBucketAndTargeting.bucket
		moduleTargeting := moduleBucketAndTargeting.moduleTargeting
		bufLockFile, err := bufconfig2.GetBufLockFileForPrefix(
			ctx,
			bucket, // Need to use the non-mapped bucket since the mapped bucket excludes the buf.lock
			moduleTargeting.moduleDirPath,
			bufconfig2.BufLockFileWithDigestResolver(
				func(ctx context.Context, remote string, commitID uuid.UUID) (bufmodule2.Digest, error) {
					commitKey, err := bufmodule2.NewCommitKey(remote, commitID, bufmodule2.DigestTypeB4)
					if err != nil {
						return nil, err
					}
					commits, err := w.commitProvider.GetCommitsForCommitKeys(ctx, []bufmodule2.CommitKey{commitKey})
					if err != nil {
						return nil, err
					}
					return commits[0].ModuleKey().Digest()
				},
			),
		)
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return nil, err
			}
		} else {
			switch fileVersion := bufLockFile.FileVersion(); fileVersion {
			case bufconfig2.FileVersionV1Beta1, bufconfig2.FileVersionV1:
			case bufconfig2.FileVersionV2:
				return nil, errors.New("got a v2 buf.lock file for a v1 buf.yaml - this is not allowed, run buf mod update to update your buf.lock file")
			default:
				return nil, syserror.Newf("unknown FileVersion: %v", fileVersion)
			}
			for _, depModuleKey := range bufLockFile.DepModuleKeys() {
				// DepModuleKeys from a BufLockFile is expected to have all transitive dependencies,
				// and we can rely on this property.
				moduleSetBuilder.AddRemoteModule(
					depModuleKey,
					false,
				)
			}
		}
		v1BufYAMLObjectData, err := bufconfig2.GetBufYAMLV1Beta1OrV1ObjectDataForPrefix(ctx, bucket, moduleTargeting.moduleDirPath)
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return nil, err
			}
		}
		v1BufLockObjectData, err := bufconfig2.GetBufLockV1Beta1OrV1ObjectDataForPrefix(ctx, bucket, moduleTargeting.moduleDirPath)
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return nil, err
			}
		}
		// Each moduleBucketAndTargeting represents a local module that we want to add to the moduleSet,
		// and we look up its moduleConfig by its bucketID, because that is guaranteed to be unique (technically,
		// moduleDirPath is also unique in v1/v1beta1, but just to be extra safe).
		moduleConfig, ok := v1WorkspaceTargeting.bucketIDToModuleConfig[moduleBucketAndTargeting.bucketID]
		if !ok {
			// This should not happen since moduleBucketAndTargeting is derived from the module
			// configs, however, we return this error as a safety check
			return nil, fmt.Errorf("no module config found for module at: %q", moduleTargeting.moduleDirPath)
		}
		moduleSetBuilder.AddLocalModule(
			mappedModuleBucket,
			moduleBucketAndTargeting.bucketID,
			moduleTargeting.isTargetModule,
			bufmodule2.LocalModuleWithFullName(moduleConfig.FullName()),
			bufmodule2.LocalModuleWithTargetPaths(
				moduleTargeting.moduleTargetPaths,
				moduleTargeting.moduleTargetExcludePaths,
			),
			bufmodule2.LocalModuleWithProtoFileTargetPath(
				moduleTargeting.moduleProtoFileTargetPath,
				moduleTargeting.includePackageFiles,
			),
			bufmodule2.LocalModuleWithV1Beta1OrV1BufYAMLObjectData(v1BufYAMLObjectData),
			bufmodule2.LocalModuleWithV1Beta1OrV1BufLockObjectData(v1BufLockObjectData),
			bufmodule2.LocalModuleWithDescription(
				getLocalModuleDescription(
					// See comments on getLocalModuleDescription.
					moduleBucketAndTargeting.bucketID,
					moduleConfig,
				),
			),
		)
	}
	moduleSet, err := moduleSetBuilder.Build()
	if err != nil {
		return nil, err
	}
	return w.getWorkspaceForBucketModuleSet(
		moduleSet,
		v1WorkspaceTargeting.bucketIDToModuleConfig,
		nil, // No PluginConfigs for v1
		nil, // No remote PluginKeys for v1
		nil, // No PolicyConfigs for v1
		nil, // No remote PolicyKeys for v1
		nil, // No Policy's PluginKeys for v1.
		v1WorkspaceTargeting.allConfiguredDepModuleRefs,
		false,
	)
}

func (w *workspaceProvider) getWorkspaceForBucketBufYAMLV2(
	ctx context.Context,
	bucket storage.ReadBucket,
	v2Targeting *v2Targeting,
) (*workspace, error) {
	moduleSetBuilder := bufmodule2.NewModuleSetBuilder(ctx, w.logger, w.moduleDataProvider, w.commitProvider)
	var (
		remotePluginKeys             []bufplugin2.PluginKey
		remotePolicyKeys             []bufpolicy.PolicyKey
		policyNameToRemotePluginKeys map[string][]bufplugin2.PluginKey
	)
	bufLockFile, err := bufconfig2.GetBufLockFileForPrefix(
		ctx,
		bucket,
		// buf.lock files live next to the buf.yaml
		".",
		// We are not passing BufLockFileWithDigestResolver here because a buf.lock
		// v2 is expected to have digests
	)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
	} else {
		switch fileVersion := bufLockFile.FileVersion(); fileVersion {
		case bufconfig2.FileVersionV1Beta1, bufconfig2.FileVersionV1:
			return nil, fmt.Errorf("got a %s buf.lock file for a v2 buf.yaml", bufLockFile.FileVersion().String())
		case bufconfig2.FileVersionV2:
		default:
			return nil, syserror.Newf("unknown FileVersion: %v", fileVersion)
		}
		for _, depModuleKey := range bufLockFile.DepModuleKeys() {
			// DepModuleKeys from a BufLockFile is expected to have all transitive dependencies,
			// and we can rely on this property.
			moduleSetBuilder.AddRemoteModule(
				depModuleKey,
				false,
			)
		}
		remotePluginKeys = bufLockFile.RemotePluginKeys()
		remotePolicyKeys = bufLockFile.RemotePolicyKeys()
		policyNameToRemotePluginKeys = bufLockFile.PolicyNameToRemotePluginKeys()
	}
	// Only check for duplicate module description in v2, which would be an user error, i.e.
	// This is not a system error:
	// modules:
	//   - path: proto
	//     excludes:
	//       - proot/foo
	//   - path: proto
	//     excludes:
	//       - proot/foo
	// but duplicate module description in v1 is a system error, which the ModuleSetBuilder catches.
	seenModuleDescriptions := make(map[string]struct{})
	for _, moduleBucketAndTargeting := range v2Targeting.moduleBucketsAndTargeting {
		mappedModuleBucket := moduleBucketAndTargeting.bucket
		moduleTargeting := moduleBucketAndTargeting.moduleTargeting
		// Each moduleBucketAndTargeting represents a local module that we want to add to the moduleSet,
		// and we look up its moduleConfig by its bucketID, because that is guaranteed to be unique (moduleDirPaths
		// are not in a v2 workspace).
		moduleConfig, ok := v2Targeting.bucketIDToModuleConfig[moduleBucketAndTargeting.bucketID]
		if !ok {
			// This should not happen since moduleBucketAndTargeting is derived from the module
			// configs, however, we return this error as a safety check
			return nil, fmt.Errorf("no module config found for module at: %q", moduleTargeting.moduleDirPath)
		}
		moduleDescription := getLocalModuleDescription(
			// See comments on getLocalModuleDescription.
			moduleConfig.DirPath(),
			moduleConfig,
		)
		if _, ok := seenModuleDescriptions[moduleDescription]; ok {
			return nil, fmt.Errorf("multiple module configs found with the same description: %s", moduleDescription)
		}
		seenModuleDescriptions[moduleDescription] = struct{}{}
		moduleSetBuilder.AddLocalModule(
			mappedModuleBucket,
			moduleBucketAndTargeting.bucketID,
			moduleTargeting.isTargetModule,
			bufmodule2.LocalModuleWithFullName(moduleConfig.FullName()),
			bufmodule2.LocalModuleWithTargetPaths(
				moduleTargeting.moduleTargetPaths,
				moduleTargeting.moduleTargetExcludePaths,
			),
			bufmodule2.LocalModuleWithProtoFileTargetPath(
				moduleTargeting.moduleProtoFileTargetPath,
				moduleTargeting.includePackageFiles,
			),
			bufmodule2.LocalModuleWithDescription(moduleDescription),
		)
	}
	moduleSet, err := moduleSetBuilder.Build()
	if err != nil {
		return nil, err
	}
	return w.getWorkspaceForBucketModuleSet(
		moduleSet,
		v2Targeting.bucketIDToModuleConfig,
		v2Targeting.bufYAMLFile.PluginConfigs(),
		remotePluginKeys,
		v2Targeting.bufYAMLFile.PolicyConfigs(),
		remotePolicyKeys,
		policyNameToRemotePluginKeys,
		v2Targeting.bufYAMLFile.ConfiguredDepModuleRefs(),
		true,
	)
}

// only use for workspaces created from buckets
func (w *workspaceProvider) getWorkspaceForBucketModuleSet(
	moduleSet bufmodule2.ModuleSet,
	bucketIDToModuleConfig map[string]bufconfig2.ModuleConfig,
	pluginConfigs []bufconfig2.PluginConfig,
	remotePluginKeys []bufplugin2.PluginKey,
	policyConfigs []bufconfig2.PolicyConfig,
	remotePolicyKeys []bufpolicy.PolicyKey,
	policyNameToRemotePluginKeys map[string][]bufplugin2.PluginKey,
	// Expected to already be unique by FullName.
	configuredDepModuleRefs []bufparse2.Ref,
	isV2 bool,
) (*workspace, error) {
	opaqueIDToLintConfig := make(map[string]bufconfig2.LintConfig)
	opaqueIDToBreakingConfig := make(map[string]bufconfig2.BreakingConfig)
	for _, module := range moduleSet.Modules() {
		if bucketID := module.BucketID(); bucketID != "" {
			moduleConfig, ok := bucketIDToModuleConfig[bucketID]
			if !ok {
				// This is a system error.
				return nil, syserror.Newf("could not get ModuleConfig for BucketID %q", bucketID)
			}
			opaqueIDToLintConfig[module.OpaqueID()] = moduleConfig.LintConfig()
			opaqueIDToBreakingConfig[module.OpaqueID()] = moduleConfig.BreakingConfig()
		} else {
			opaqueIDToLintConfig[module.OpaqueID()] = bufconfig2.DefaultLintConfigV1
			opaqueIDToBreakingConfig[module.OpaqueID()] = bufconfig2.DefaultBreakingConfigV1
		}
	}
	return newWorkspace(
		moduleSet,
		opaqueIDToLintConfig,
		opaqueIDToBreakingConfig,
		pluginConfigs,
		remotePluginKeys,
		policyConfigs,
		remotePolicyKeys,
		policyNameToRemotePluginKeys,
		configuredDepModuleRefs,
		isV2,
	), nil
}

// This formats a module name based on its module config entry in the v2 buf.yaml:
// `path: foo, includes: ["foo/v1, "foo/v2"], excludes: "foo/v1/internal"`.
//
// For v1/v1beta1 modules, pathDescription should be bucketID.
// For v2 modules, pathDescription should be moduleConfig.DirPath().
//
// We edit bucketIDs in v2 to include an index since directories can be overlapping.
// We would want to use moduleConfig.DirPath() everywhere, but it is always "." in
// v1/v1beta1, and it's not a good description.
func getLocalModuleDescription(pathDescription string, moduleConfig bufconfig2.ModuleConfig) string {
	description := fmt.Sprintf("path: %q", pathDescription)
	moduleDirPath := moduleConfig.DirPath()
	relIncludePaths := moduleConfig.RootToIncludes()["."]
	includePaths := xslices.Map(relIncludePaths, func(relInclude string) string {
		return normalpath.Join(moduleDirPath, relInclude)
	})
	switch len(includePaths) {
	case 0:
	case 1:
		description = fmt.Sprintf("%s, includes: %q", description, includePaths[0])
	default:
		description = fmt.Sprintf("%s, includes: [%s]", description, xstrings.JoinSliceQuoted(includePaths, ", "))
	}
	relExcludePaths := moduleConfig.RootToExcludes()["."]
	excludePaths := xslices.Map(relExcludePaths, func(relInclude string) string {
		return normalpath.Join(moduleDirPath, relInclude)
	})
	switch len(excludePaths) {
	case 0:
	case 1:
		description = fmt.Sprintf("%s, excludes: %q", description, excludePaths[0])
	default:
		description = fmt.Sprintf("%s, excludes: [%s]", description, xstrings.JoinSliceQuoted(excludePaths, ", "))
	}
	return description
}
