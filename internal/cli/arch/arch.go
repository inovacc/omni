package arch

import (
	"fmt"
	"io"
	"runtime"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/internal/cli/uname"
	"github.com/inovacc/omni/pkg/cobra/helper/output"
)

// ArchOptions configures the arch command behavior
type ArchOptions struct {
	OutputFormat output.Format // output format
}

// ArchResult represents arch output for JSON
type ArchResult struct {
	Architecture string `json:"architecture"`
	GOARCH       string `json:"goarch"`
}

// RunArch prints machine architecture
func RunArch(w io.Writer, opts ArchOptions) error {
	arch := uname.MapMachine(runtime.GOARCH)

	f := output.New(w, opts.OutputFormat)
	if f.IsJSON() {
		return f.Print(ArchResult{Architecture: arch, GOARCH: runtime.GOARCH})
	}

	if _, err := fmt.Fprintln(w, arch); err != nil {
		return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("arch: write: %s", err))
	}

	return nil
}
