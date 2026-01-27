package cli

import (
	"encoding/json"
	"fmt"
	"os"
)

func RunTouch(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("touch: missing operand")
	}

	for _, path := range args {
		err := Touch(path)
		if err != nil {
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

func RunStat(args []string, jsonMode bool) error {
	if len(args) == 0 {
		return fmt.Errorf("stat: missing operand")
	}

	var results []StatInfo
	for _, path := range args {
		info, err := Stat(path)
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
			fmt.Printf("  File: %s\n", statInfo.Name)
			fmt.Printf("  Size: %d\tBlocks: %d\tIO Block: %d\t", statInfo.Size, 0, 0) // Simplified
			if statInfo.IsDir {
				fmt.Print("directory\n")
			} else {
				fmt.Print("regular file\n")
			}
			fmt.Printf("Device: %s\tInode: %d\tLinks: %d\n", "unknown", 0, 0)
			fmt.Printf("Access: (%04o/%s)  Uid: (%d/ %s)   Gid: (%d/ %s)\n", statInfo.Mode.Perm(), statInfo.Mode.String(), 0, "unknown", 0, "unknown")
			fmt.Printf("Access: %s\n", statInfo.ModTime)
			fmt.Printf("Modify: %s\n", statInfo.ModTime)
			fmt.Printf("Change: %s\n", statInfo.ModTime)
			fmt.Printf(" Birth: -\n")
		}
	}

	if jsonMode {
		return json.NewEncoder(os.Stdout).Encode(results)
	}

	return nil
}
