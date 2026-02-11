package userdirs

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// DownloadsDir returns the user's Downloads directory path.
func DownloadsDir() (string, error) {
	home, err := homeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, "Downloads"), nil
}

// DocumentsDir returns the user's Documents directory path.
func DocumentsDir() (string, error) {
	home, err := homeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, "Documents"), nil
}

func homeDir() (string, error) {
	if runtime.GOOS == "windows" {
		if profile := os.Getenv("USERPROFILE"); profile != "" {
			return profile, nil
		}
	}

	if home, err := os.UserHomeDir(); err == nil && home != "" {
		return home, nil
	}

	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}

	if runtime.GOOS == "windows" {
		drive := os.Getenv("HOMEDRIVE")
		path := os.Getenv("HOMEPATH")
		if drive != "" && path != "" {
			return drive + path, nil
		}
	}

	return "", fmt.Errorf("cannot determine user home directory")
}
