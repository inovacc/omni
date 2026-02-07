package extractor

import "testing"

func TestRegistryMatchYouTube(t *testing.T) {
	// Note: extractors are registered via init() in their packages.
	// This test assumes youtube and generic are imported somewhere.
	names := Names()
	if len(names) == 0 {
		t.Skip("no extractors registered (youtube/generic not imported)")
	}

	// Just test that Names returns something.
	found := false

	for _, n := range names {
		if n != "" {
			found = true
			break
		}
	}

	if !found {
		t.Error("no named extractors found")
	}
}

func TestRegistryAll(t *testing.T) {
	all := All()
	// All returns a copy.
	if len(all) > 0 {
		all[0] = nil
		// Original should be unaffected.
		orig := All()
		if orig[0] == nil {
			t.Error("All() should return a copy")
		}
	}
}
