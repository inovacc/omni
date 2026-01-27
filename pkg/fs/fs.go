package fs

import (
	"os"
)

func Cd(path string) error {
	return os.Chdir(path)
}

func Chmod(path string, mode os.FileMode) error {
	return os.Chmod(path, mode)
}
