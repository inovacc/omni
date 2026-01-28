package cli

import (
	"fmt"
	"io"
	"runtime"
)

// RunArch prints machine architecture
func RunArch(w io.Writer) error {
	arch := mapMachine(runtime.GOARCH)
	_, _ = fmt.Fprintln(w, arch)

	return nil
}
