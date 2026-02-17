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

package bufcli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsOfflineModeValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		envValue string
		want     bool
	}{
		{
			name:     "empty",
			envValue: "",
			want:     false,
		},
		{
			name:     "zero",
			envValue: "0",
			want:     false,
		},
		{
			name:     "one",
			envValue: "1",
			want:     true,
		},
		{
			name:     "true_lowercase",
			envValue: "true",
			want:     true,
		},
		{
			name:     "false_lowercase",
			envValue: "false",
			want:     false,
		},
		{
			name:     "TRUE_uppercase",
			envValue: "TRUE",
			want:     false, // case sensitive, only "true" works
		},
		{
			name:     "yes",
			envValue: "yes",
			want:     false, // only "1" and "true" are valid
		},
		{
			name:     "random_string",
			envValue: "offline",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := isOfflineModeValue(tt.envValue)
			require.Equal(t, tt.want, got)
		})
	}
}
