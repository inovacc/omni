package cli

import (
	"encoding/json"
	"fmt"
	"os"
)

func RunLs(args []string, jsonMode bool) error {
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	entries, err := Ls(dir)
	if err != nil {
		return err
	}

	if jsonMode {
		return json.NewEncoder(os.Stdout).Encode(entries)
	}

	for _, name := range entries {
		fmt.Println(name)
	}
	return nil
}

func Ls(path string) ([]string, error) {
	if path == "" {
		path = "."
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	out := make([]string, 0, len(entries))
	for _, e := range entries {
		out = append(out, e.Name())
	}
	return out, nil
}
