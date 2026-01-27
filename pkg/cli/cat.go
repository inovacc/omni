package cli

import (
	"io"
	"os"
)

func RunCat(args []string) error {
	return CatFiles(os.Stdout, args)
}

func Cat(w io.Writer, r io.Reader) error {
	_, err := io.Copy(w, r)
	return err
}

func CatFiles(w io.Writer, paths []string) error {
	if len(paths) == 0 {
		return Cat(w, os.Stdin)
	}

	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		err = Cat(w, f)
		f.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
