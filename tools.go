//go:build tools

//go:generate go run scripts/cmdtree/cmdtree.go -out-tree docs/COMMANDS.md -out-go cmd/cmdtree.go

package main
