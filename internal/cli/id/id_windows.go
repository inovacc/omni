//go:build windows

package id

import "github.com/inovacc/omni/internal/cli/cmderr"

// On Windows, os/user returns SID strings (e.g. "S-1-5-21-..."), which have no
// numeric uid/gid equivalent. The `id` command itself prints the string IDs and
// works; these numeric library helpers are unsupported here.

// GetUID is unsupported on Windows (SIDs are not numeric).
func GetUID() (int, error) {
	return 0, cmderr.Wrap(cmderr.ErrUnsupported, "id: numeric UID unavailable on Windows (SID-based)")
}

// GetGID is unsupported on Windows (SIDs are not numeric).
func GetGID() (int, error) {
	return 0, cmderr.Wrap(cmderr.ErrUnsupported, "id: numeric GID unavailable on Windows (SID-based)")
}

// GetGroups is unsupported on Windows (SIDs are not numeric).
func GetGroups() ([]int, error) {
	return nil, cmderr.Wrap(cmderr.ErrUnsupported, "id: numeric group IDs unavailable on Windows (SID-based)")
}
