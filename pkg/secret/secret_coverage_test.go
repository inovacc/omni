package secret_test

import (
	"testing"

	"github.com/inovacc/omni/pkg/secret"
)

func TestKeyDestroyZeroesBytes(t *testing.T) {
	tests := []struct {
		name string
		raw  []byte
	}{
		{"non-empty", []byte{1, 2, 3, 4, 5}},
		{"single byte", []byte{0xff}},
		{"empty", []byte{}},
		{"nil", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := secret.New(tt.raw)
			k.Destroy()

			got := k.Bytes()
			for i, b := range got {
				if b != 0 {
					t.Errorf("Destroy() left byte[%d] = %d, want 0", i, b)
				}
			}

			if len(got) != len(tt.raw) {
				t.Errorf("Destroy() changed length: got %d, want %d", len(got), len(tt.raw))
			}
		})
	}
}
