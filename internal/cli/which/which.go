package which

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// WhichOptions configures the which command behavior
type WhichOptions struct {
	All  bool // -a: print all matches
	JSON bool // --json: output as JSON
}

// WhichResult represents which output for JSON
type WhichResult struct {
	Command string   `json:"command"`
	Paths   []string `json:"paths"`
	Found   bool     `json:"found"`
}

// RunWhich locates a command
func RunWhich(w io.Writer, args []string, opts WhichOptions) error {
	if len(args) == 0 {
		return fmt.Errorf("which: missing operand")
	}

	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		return fmt.Errorf("which: PATH not set")
	}

	pathSep := string(os.PathListSeparator)
	paths := strings.Split(pathEnv, pathSep)

	exitCode := 0
	var jsonResults []WhichResult

	for _, cmd := range args {
		found := false
		var foundPaths []string

		for _, dir := range paths {
			fullPath := filepath.Join(dir, cmd)
			matches := findExecutable(fullPath)

			for _, match := range matches {
				if opts.JSON {
					foundPaths = append(foundPaths, match)
				} else {
					_, _ = fmt.Fprintln(w, match)
				}
				found = true

				if !opts.All {
					break
				}
			}

			if found && !opts.All {
				break
			}
		}

		if opts.JSON {
			jsonResults = append(jsonResults, WhichResult{Command: cmd, Paths: foundPaths, Found: found})
		}

		if !found {
			exitCode = 1
		}
	}

	if opts.JSON {
		return json.NewEncoder(w).Encode(jsonResults)
	}

	if exitCode != 0 {
		return fmt.Errorf("which: no %s in PATH", args[len(args)-1])
	}

	return nil
}

func findExecutable(path string) []string {
	var results []string

	// On Windows, also check with common extensions
	if runtime.GOOS == "windows" {
		pathExt := os.Getenv("PATHEXT")
		if pathExt == "" {
			pathExt = ".COM;.EXE;.BAT;.CMD"
		}

		exts := strings.Split(strings.ToLower(pathExt), ";")

		// Check exact path first
		if isExec(path) {
			results = append(results, path)
		}

		// Check with extensions
		for _, ext := range exts {
			p := path + ext
			if isExec(p) {
				results = append(results, p)
			}
		}
	} else if isExec(path) {
		results = append(results, path)
	}

	return results
}

func isExec(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	if info.IsDir() {
		return false
	}

	// On Unix, check execute permission
	if runtime.GOOS != "windows" {
		return info.Mode()&0111 != 0
	}

	// On Windows, just check if file exists
	return true
}
