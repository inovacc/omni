package cli

import (
	"context"
	"fmt"

	"github.com/inovacc/twig/pkg/tree"
)

//t := tree.NewTree(
//    tree.WithMaxDepth(3),                              // Limit depth
//    tree.WithShowHidden(true),                         // Show hidden files
//    tree.WithIgnorePatterns("node_modules", ".git"),   // Ignore patterns
//    tree.WithDirSlash(false),                          // No trailing slashes
//    tree.WithColors(true),                             // Enable colors
//)

func Tree(path string, opts ...tree.TreeOption) error {
	t := tree.NewTree(opts...)
	output, err := t.Generate(context.Background(), path)
	if err != nil {
		return err
	}
	fmt.Println(output)

	return nil
}
