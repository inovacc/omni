package cli

import (
	"github.com/inovacc/goshell/pkg/fs"
	"os"
)

func RunCat(args []string) error {
	return fs.CatFiles(os.Stdout, args)
}
