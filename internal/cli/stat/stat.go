package stat

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/inovacc/omni/internal/cli/output"
)

// StatOptions configures the stat command behavior
type StatOptions struct {
	OutputFormat output.Format // output format (text, json, table)
}

// TouchOptions configures the touch command behavior
type TouchOptions struct{}

func RunTouch(args []string, _ TouchOptions) error {
	if len(args) == 0 {
		return fmt.Errorf("touch: missing operand")
	}

	for _, path := range args {
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			f, createErr := os.Create(path)
			if createErr != nil {
				return fmt.Errorf("touch: %w", createErr)
			}

			_ = f.Close()

			continue
		}

		now := time.Now()
		if err := os.Chtimes(path, now, now); err != nil {
			return fmt.Errorf("touch: %w", err)
		}
	}

	return nil
}

type StatInfo struct {
	Name    string      `json:"name"`
	Size    int64       `json:"size"`
	Mode    os.FileMode `json:"mode"`
	ModTime string      `json:"mod_time"`
	IsDir   bool        `json:"is_dir"`
}

func RunStat(w io.Writer, args []string, opts StatOptions) error {
	if len(args) == 0 {
		return fmt.Errorf("stat: missing operand")
	}

	f := output.New(w, opts.OutputFormat)
	jsonMode := f.IsJSON()

	var results []StatInfo

	for _, path := range args {
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("stat: %w", err)
		}

		statInfo := StatInfo{
			Name:    info.Name(),
			Size:    info.Size(),
			Mode:    info.Mode(),
			ModTime: info.ModTime().Format("2006-01-02 15:04:05.000000000 -0700"),
			IsDir:   info.IsDir(),
		}
		results = append(results, statInfo)

		if !jsonMode {
			_, _ = fmt.Fprintf(w, "  File: %s\n", statInfo.Name)
			_, _ = fmt.Fprintf(w, "  Size: %d\tBlocks: %d\tIO Block: %d\t", statInfo.Size, 0, 0) // Simplified

			if statInfo.IsDir {
				_, _ = fmt.Fprint(w, "directory\n")
			} else {
				_, _ = fmt.Fprint(w, "regular file\n")
			}

			_, _ = fmt.Fprintf(w, "Device: %s\tInode: %d\tLinks: %d\n", "unknown", 0, 0)
			_, _ = fmt.Fprintf(w, "Access: (%04o/%s)  Uid: (%d/ %s)   Gid: (%d/ %s)\n", statInfo.Mode.Perm(), statInfo.Mode.String(), 0, "unknown", 0, "unknown")
			_, _ = fmt.Fprintf(w, "Access: %s\n", statInfo.ModTime)
			_, _ = fmt.Fprintf(w, "Modify: %s\n", statInfo.ModTime)
			_, _ = fmt.Fprintf(w, "Change: %s\n", statInfo.ModTime)
			_, _ = fmt.Fprintf(w, " Birth: -\n")
		}
	}

	if jsonMode {
		return f.Print(results)
	}

	return nil
}
