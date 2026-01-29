package arch

import (
	"fmt"
	"io"
	"runtime"

	"github.com/inovacc/omni/internal/cli/uname"
)

// RunArch prints machine architecture
func RunArch(w io.Writer) error {
	arch := uname.MapMachine(runtime.GOARCH)
	_, _ = fmt.Fprintln(w, arch)

	return nil
}
