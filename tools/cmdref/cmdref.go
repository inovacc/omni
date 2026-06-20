//go:build ignore

// Command cmdref regenerates docs/COMMANDS.md from the in-process omni command
// tree by calling cmd.GenerateCommandReference. Run: go run tools/cmdref/cmdref.go
// Pure Go, no exec of the omni binary. Output is deterministic so CI can diff it.
package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/inovacc/omni/cmd"
)

func main() {
	var buf bytes.Buffer
	if err := cmd.GenerateCommandReference(&buf); err != nil {
		fmt.Fprintln(os.Stderr, "cmdref:", err)
		os.Exit(1)
	}
	if err := os.WriteFile("docs/COMMANDS.md", buf.Bytes(), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "cmdref:", err)
		os.Exit(1)
	}
}
