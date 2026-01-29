package id

import (
	"encoding/json"
	"fmt"
	"io"
	"os/user"
	"strconv"
	"strings"
)

// IDOptions configures the id command behavior
type IDOptions struct {
	User     bool   // -u: print only the effective user ID
	Group    bool   // -g: print only the effective group ID
	Groups   bool   // -G: print all group IDs
	Name     bool   // -n: print name instead of number (requires -u, -g, or -G)
	Real     bool   // -r: print real ID instead of effective ID
	Username string // username to look up (optional)
	JSON     bool   // --json: output as JSON
}

// IDInfo contains user identity information
type IDInfo struct {
	UID      string   `json:"uid"`
	GID      string   `json:"gid"`
	Username string   `json:"username"`
	Groups   []string `json:"groups"`
}

// RunID executes the id command
func RunID(w io.Writer, opts IDOptions) error {
	var (
		u   *user.User
		err error
	)

	if opts.Username != "" {
		u, err = user.Lookup(opts.Username)
	} else {
		u, err = user.Current()
	}

	if err != nil {
		return fmt.Errorf("cannot find user: %w", err)
	}

	// Get group info
	groups, _ := u.GroupIds()

	if opts.JSON {
		info := IDInfo{
			UID:      u.Uid,
			GID:      u.Gid,
			Username: u.Username,
			Groups:   groups,
		}

		return json.NewEncoder(w).Encode(info)
	}

	// Single value modes
	if opts.User {
		if opts.Name {
			_, _ = fmt.Fprintln(w, u.Username)
		} else {
			_, _ = fmt.Fprintln(w, u.Uid)
		}

		return nil
	}

	if opts.Group {
		if opts.Name {
			g, err := user.LookupGroupId(u.Gid)
			if err != nil {
				_, _ = fmt.Fprintln(w, u.Gid)
			} else {
				_, _ = fmt.Fprintln(w, g.Name)
			}
		} else {
			_, _ = fmt.Fprintln(w, u.Gid)
		}

		return nil
	}

	if opts.Groups {
		if opts.Name {
			var names []string

			for _, gid := range groups {
				g, err := user.LookupGroupId(gid)
				if err != nil {
					names = append(names, gid)
				} else {
					names = append(names, g.Name)
				}
			}

			for i, name := range names {
				if i > 0 {
					_, _ = fmt.Fprint(w, " ")
				}

				_, _ = fmt.Fprint(w, name)
			}

			_, _ = fmt.Fprintln(w)
		} else {
			for i, gid := range groups {
				if i > 0 {
					_, _ = fmt.Fprint(w, " ")
				}

				_, _ = fmt.Fprint(w, gid)
			}

			_, _ = fmt.Fprintln(w)
		}

		return nil
	}

	// Default: print all info in standard format
	// uid=1000(username) gid=1000(groupname) groups=1000(group1),1001(group2),...
	var output strings.Builder
	output.WriteString(fmt.Sprintf("uid=%s", u.Uid))
	output.WriteString(fmt.Sprintf("(%s)", u.Username))

	g, err := user.LookupGroupId(u.Gid)

	groupName := u.Gid
	if err == nil {
		groupName = g.Name
	}

	output.WriteString(fmt.Sprintf(" gid=%s(%s)", u.Gid, groupName))

	if len(groups) > 0 {
		output.WriteString(" groups=")

		for i, gid := range groups {
			if i > 0 {
				output.WriteString(",")
			}

			gname := gid
			if grp, err := user.LookupGroupId(gid); err == nil {
				gname = grp.Name
			}

			output.WriteString(fmt.Sprintf("%s(%s)", gid, gname))
		}
	}

	_, _ = fmt.Fprintln(w, output.String())

	return nil
}

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
