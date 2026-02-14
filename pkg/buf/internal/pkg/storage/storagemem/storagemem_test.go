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

package storagemem_test

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/inovacc/omni/pkg/buf/internal/pkg/normalpath"
	storage2 "github.com/inovacc/omni/pkg/buf/internal/pkg/storage"
	"github.com/inovacc/omni/pkg/buf/internal/pkg/storage/storagemem"
	"github.com/inovacc/omni/pkg/buf/internal/pkg/storage/storageos"
	"github.com/inovacc/omni/pkg/buf/internal/pkg/storage/storagetesting"
	"github.com/stretchr/testify/require"
)

var storagetestingDirPath string

func init() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("cannot get current file path")
	}

	storagetestingDirPath = filepath.Join(filepath.Dir(filename), "..", "storagetesting")
}

func TestMem(t *testing.T) {
	t.Parallel()
	storagetesting.RunTestSuite(
		t,
		storagetestingDirPath,
		testNewReadBucket,
		testNewWriteBucket,
		testWriteBucketToReadBucket,
		false,
	)
}

func testNewReadBucket(t *testing.T, dirPath string, storageosProvider storageos.Provider) (storage2.ReadBucket, storagetesting.GetExternalPathFunc) {
	osBucket, err := storageosProvider.NewReadWriteBucket(
		dirPath,
		storageos.ReadWriteBucketWithSymlinksIfSupported(),
	)
	require.NoError(t, err)

	readWriteBucket := storagemem.NewReadWriteBucket()
	_, err = storage2.Copy(
		context.Background(),
		osBucket,
		readWriteBucket,
		storage2.CopyWithExternalAndLocalPaths(),
	)
	require.NoError(t, err)

	return readWriteBucket, func(t *testing.T, rootPath string, path string) string {
		// Join calls Clean
		return normalpath.Unnormalize(normalpath.Join(rootPath, path))
	}
}

func testNewWriteBucket(*testing.T, storageos.Provider) storage2.WriteBucket {
	return storagemem.NewReadWriteBucket()
}

func testWriteBucketToReadBucket(t *testing.T, writeBucket storage2.WriteBucket) storage2.ReadBucket {
	// hacky
	readWriteBucket, ok := writeBucket.(storage2.ReadWriteBucket)
	require.True(t, ok)

	return readWriteBucket
}
