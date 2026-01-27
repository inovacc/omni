//go:build tools

//go:generate go run scripts/cmdtree/cmdtree.go -out-tree docs/COMMANDS.md -out-go cmd/cmdtree.go
//go:generate go run ./scripts/genversion/genversion.go

package main

import (
	_ "github.com/inovacc/genversioninfo"
)
