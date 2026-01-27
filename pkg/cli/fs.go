package cli

import (
	"fmt"
	"io"
	"os"
	"time"
)

func Cd(path string) error {
	return os.Chdir(path)
}

func Chmod(path string, mode os.FileMode) error {
	return os.Chmod(path, mode)
}

func Mkdir(path string, perm os.FileMode, parents bool) error {
	if parents {
		return os.MkdirAll(path, perm)
	}
	return os.Mkdir(path, perm)
}

func Rmdir(path string) error {
	return os.Remove(path)
}

func Touch(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		return f.Close()
	}

	now := time.Now()
	return os.Chtimes(path, now, now)
}

func Rm(path string, recursive bool) error {
	if recursive {
		return os.RemoveAll(path)
	}
	return os.Remove(path)
}

func IsNotExist(err error) bool {
	return os.IsNotExist(err)
}

func Copy(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}

func Move(src, dst string) error {
	return os.Rename(src, dst)
}

// Stat returns file information
func Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

// Lstat returns file information without following symlinks
func Lstat(path string) (os.FileInfo, error) {
	return os.Lstat(path)
}
