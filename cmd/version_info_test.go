package cmd

import "testing"

func TestRootVersionNonEmpty(t *testing.T) {
	if rootVersion() == "" {
		t.Fatal("rootVersion() must never be empty (fallback to (devel) or build info)")
	}
}
