package cli

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
)

// ChownOptions configures the chown command behavior
type ChownOptions struct {
	Recursive      bool   // -R: operate on files and directories recursively
	Verbose        bool   // -v: output a diagnostic for every file processed
	Changes        bool   // -c: like verbose but report only when a change is made
	Silent         bool   // -f: suppress most error messages
	Dereference    bool   // dereference symbolic links (default)
	NoDereference  bool   // -h: affect symbolic links instead of referenced file
	Reference      string // --reference: use RFILE's owner and group
	PreserveRoot   bool   // --preserve-root: fail to operate recursively on '/'
	NoPreserveRoot bool   // --no-preserve-root: do not treat '/' specially
	From           string // --from: change only if current owner/group match
}

// RunChown changes file owner and group
func RunChown(w io.Writer, args []string, opts ChownOptions) error {
	if len(args) < 2 {
		return fmt.Errorf("chown: missing operand")
	}

	ownerGroup := args[0]
	files := args[1:]

	// Parse owner:group
	uid, gid, err := parseOwnerGroup(ownerGroup, opts.Reference)
	if err != nil {
		return fmt.Errorf("chown: %w", err)
	}

	for _, file := range files {
		if opts.PreserveRoot && opts.Recursive && (file == "/" || filepath.Clean(file) == "/") {
			return fmt.Errorf("chown: it is dangerous to operate recursively on '/'")
		}

		if opts.Recursive {
			err := filepath.WalkDir(file, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					if !opts.Silent {
						_, _ = fmt.Fprintf(os.Stderr, "chown: cannot access '%s': %v\n", path, err)
					}
					return nil
				}
				return chownFile(w, path, uid, gid, opts)
			})
			if err != nil {
				return err
			}
		} else {
			if err := chownFile(w, file, uid, gid, opts); err != nil {
				if !opts.Silent {
					_, _ = fmt.Fprintf(os.Stderr, "chown: %v\n", err)
				}
			}
		}
	}

	return nil
}

func parseOwnerGroup(spec string, reference string) (int, int, error) {
	if reference != "" {
		info, err := os.Stat(reference)
		if err != nil {
			return -1, -1, fmt.Errorf("cannot stat '%s': %w", reference, err)
		}
		return getFileOwner(info)
	}

	uid := -1
	gid := -1

	// Parse owner:group or owner.group
	var owner, group string
	if idx := strings.IndexAny(spec, ":."); idx != -1 {
		owner = spec[:idx]
		group = spec[idx+1:]
	} else {
		owner = spec
	}

	// Parse owner
	if owner != "" {
		if id, err := strconv.Atoi(owner); err == nil {
			uid = id
		} else {
			u, err := user.Lookup(owner)
			if err != nil {
				return -1, -1, fmt.Errorf("invalid user: '%s'", owner)
			}
			uid, _ = strconv.Atoi(u.Uid)
		}
	}

	// Parse group
	if group != "" {
		if id, err := strconv.Atoi(group); err == nil {
			gid = id
		} else {
			g, err := user.LookupGroup(group)
			if err != nil {
				return -1, -1, fmt.Errorf("invalid group: '%s'", group)
			}
			gid, _ = strconv.Atoi(g.Gid)
		}
	}

	return uid, gid, nil
}

func chownFile(w io.Writer, path string, uid, gid int, opts ChownOptions) error {
	var err error

	if opts.NoDereference {
		err = os.Lchown(path, uid, gid)
	} else {
		err = os.Chown(path, uid, gid)
	}

	if err != nil {
		return fmt.Errorf("changing ownership of '%s': %w", path, err)
	}

	if opts.Verbose {
		ownerStr := fmt.Sprintf("%d", uid)
		if uid == -1 {
			ownerStr = ""
		}
		groupStr := fmt.Sprintf("%d", gid)
		if gid == -1 {
			groupStr = ""
		}
		_, _ = fmt.Fprintf(w, "ownership of '%s' changed to %s:%s\n", path, ownerStr, groupStr)
	}

	return nil
}

// Chown changes the numeric uid and gid of the named file
func Chown(name string, uid, gid int) error {
	return os.Chown(name, uid, gid)
}

// Lchown changes the numeric uid and gid of the named file (doesn't follow symlinks)
func Lchown(name string, uid, gid int) error {
	return os.Lchown(name, uid, gid)
}
