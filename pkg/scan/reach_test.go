package scan

import (
	"testing"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

func TestScanSourceUnsupported(t *testing.T) {
	_, err := ScanSource("./...", &DB{byPkg: map[string][]osvEntry{}}, Options{})
	if err == nil || !cmderr.IsUnsupported(err) {
		t.Fatalf("ScanSource = %v, want ErrUnsupported", err)
	}
}
