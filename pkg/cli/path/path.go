package path

import "path/filepath"

func Realpath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	return filepath.EvalSymlinks(abs)
}

func Dirname(path string) string {
	return filepath.Dir(path)
}

func Basename(path string) string {
	return filepath.Base(path)
}

func Join(paths ...string) string {
	return filepath.Join(paths...)
}
