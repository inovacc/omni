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

package buftesting

import (
	"context"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/inovacc/omni/pkg/buf/internal/buf/bufprotoc"
	bufmodule2 "github.com/inovacc/omni/pkg/buf/internal/bufpkg/bufmodule"
	"github.com/inovacc/omni/pkg/buf/internal/pkg/github/githubtesting"
	"github.com/inovacc/omni/pkg/buf/internal/pkg/normalpath"
	"github.com/inovacc/omni/pkg/buf/internal/pkg/storage/storageos"
	"github.com/inovacc/omni/pkg/buf/internal/standard/xlog/xslog"
	"github.com/stretchr/testify/require"
)

const (
	// NumGoogleapisFiles is the number of googleapis files on the current test commit.
	NumGoogleapisFiles = 1574
	// NumGoogleapisFilesWithImports is the number of googleapis files on the current test commit with imports.
	NumGoogleapisFilesWithImports = 1585

	testGoogleapisCommit = "37c923effe8b002884466074f84bc4e78e6ade62"
)

var (
	testHTTPClient = &http.Client{
		Timeout: 10 * time.Second,
	}
	testStorageosProvider = storageos.NewProvider(storageos.ProviderWithSymlinks())
	testArchiveReader     = githubtesting.NewArchiveReader(
		xslog.NopLogger,
		testStorageosProvider,
		testHTTPClient,
	)
	testGoogleapisDirPath = filepath.Join("cache", "googleapis")
)

// GetGoogleapisDirPath gets the path to a clone of googleapis.
func GetGoogleapisDirPath(t *testing.T, buftestingDirPath string) string {
	googleapisDirPath := filepath.Join(buftestingDirPath, testGoogleapisDirPath)
	require.NoError(
		t,
		testArchiveReader.GetArchive(
			context.Background(),
			googleapisDirPath,
			"googleapis",
			"googleapis",
			testGoogleapisCommit,
		),
	)

	return googleapisDirPath
}

// GetProtocFilePaths gets the file paths for protoc.
//
// Limit limits the number of files returned if > 0.
// protoc has a fixed size for number of characters to argument list.
func GetProtocFilePaths(t *testing.T, dirPath string, limit int) []string {
	realFilePaths, err := GetProtocFilePathsErr(context.Background(), dirPath, limit)
	require.NoError(t, err)

	return realFilePaths
}

// GetProtocFilePathsErr is like GetProtocFilePaths except it returns an error and accepts a ctx.
func GetProtocFilePathsErr(ctx context.Context, dirPath string, limit int) ([]string, error) {
	// TODO FUTURE: This is a really convoluted way to get protoc files. It also may have an
	// impact on our dependency tree.
	moduleSet, err := bufprotoc.NewModuleSetForProtoc(
		ctx,
		xslog.NopLogger,
		testStorageosProvider,
		[]string{dirPath},
		nil,
	)
	if err != nil {
		return nil, err
	}

	targetFileInfos, err := bufmodule2.GetTargetFileInfos(
		ctx,
		bufmodule2.ModuleSetToModuleReadBucketWithOnlyProtoFiles(
			moduleSet,
		),
	)
	if err != nil {
		return nil, err
	}

	realFilePaths := make([]string, len(targetFileInfos))
	for i, fileInfo := range targetFileInfos {
		realFilePaths[i] = normalpath.Unnormalize(normalpath.Join(dirPath, fileInfo.Path()))
	}

	if limit > 0 && len(realFilePaths) > limit {
		realFilePaths = realFilePaths[:limit]
	}

	return realFilePaths, nil
}
