//go:build !windows

package id

import (
	"os/user"
	"strconv"
)

// GetUID returns the current user's UID
func GetUID() (int, error) {
	u, err := user.Current()
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(u.Uid)
}

// GetGID returns the current user's primary GID
func GetGID() (int, error) {
	u, err := user.Current()
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(u.Gid)
}

// GetGroups returns the current user's group IDs
func GetGroups() ([]int, error) {
	u, err := user.Current()
	if err != nil {
		return nil, err
	}

	gids, err := u.GroupIds()
	if err != nil {
		return nil, err
	}

	result := make([]int, 0, len(gids))
	for _, gid := range gids {
		if id, err := strconv.Atoi(gid); err == nil {
			result = append(result, id)
		}
	}

	return result, nil
}
