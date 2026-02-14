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

package bufwktstore

import (
	"context"
	"log/slog"

	"github.com/inovacc/omni/pkg/buf/internal/gen/data/datawkt"
	storage2 "github.com/inovacc/omni/pkg/buf/internal/pkg/storage"
)

type store struct {
	logger *slog.Logger
	bucket storage2.ReadWriteBucket
}

func newStore(
	logger *slog.Logger,
	bucket storage2.ReadWriteBucket,
) *store {
	return &store{
		logger: logger,
		bucket: bucket,
	}
}

func (s *store) GetBucket(ctx context.Context) (storage2.ReadBucket, error) {
	wktBucket := storage2.MapReadWriteBucket(s.bucket, storage2.MapOnPrefix(datawkt.Version))

	isEmpty, err := storage2.IsEmpty(ctx, wktBucket, "")
	if err != nil {
		return nil, err
	}

	if isEmpty {
		if err := copyToBucket(ctx, wktBucket); err != nil {
			return nil, err
		}
	} else {
		diff, err := storage2.DiffBytes(ctx, datawkt.ReadBucket, wktBucket)
		if err != nil {
			return nil, err
		}

		if len(diff) > 0 {
			if err := deleteBucket(ctx, wktBucket); err != nil {
				return nil, err
			}

			if err := copyToBucket(ctx, wktBucket); err != nil {
				return nil, err
			}
		}
	}

	return storage2.StripReadBucketExternalPaths(wktBucket), nil
}

func copyToBucket(ctx context.Context, wktBucket storage2.WriteBucket) error {
	_, err := storage2.Copy(ctx, datawkt.ReadBucket, wktBucket)
	return err
}

func deleteBucket(ctx context.Context, wktBucket storage2.WriteBucket) error {
	return wktBucket.DeleteAll(ctx, "")
}
