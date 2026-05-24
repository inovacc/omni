package procutil

import (
	"debug/buildinfo"
	"path/filepath"
	"strings"
)

// GoBinaryInfo captures Go-specific data extracted from a binary on disk.
// Populated only when the binary is a Go binary (Runtime == RuntimeGo).
type GoBinaryInfo struct {
	Path      string
	GoVersion string
	MainPath  string
	Module    string
	Settings  map[string]string
}

// ReadGoBinary attempts to read Go build info from a binary path.
// Returns (info, true, nil) on success, (zero, false, nil) for non-Go binaries
// or unreadable files. Errors are returned only when the file exists but
// inspection failed for a non-classification reason.
func ReadGoBinary(path string) (GoBinaryInfo, bool, error) {
	bi, err := buildinfo.ReadFile(path)
	if err != nil {
		// buildinfo.ReadFile returns an error both for non-Go binaries and
		// real read failures (permission denied, truncated file, etc.).
		// For procutil's needs, ok=false is enough — callers that care about
		// the distinction can stat the file separately.
		return GoBinaryInfo{}, false, nil
	}
	settings := make(map[string]string, len(bi.Settings))
	for _, s := range bi.Settings {
		settings[s.Key] = s.Value
	}
	return GoBinaryInfo{
		Path:      path,
		GoVersion: bi.GoVersion,
		MainPath:  bi.Main.Path,
		Module:    bi.Path,
		Settings:  settings,
	}, true, nil
}

// classifyExe inspects an executable path and returns its runtime plus Go
// build info (non-nil only when Runtime == RuntimeGo). It returns
// RuntimeUnknown for binaries it cannot identify.
func classifyExe(exe string) (Runtime, *GoBinaryInfo) {
	// Try Go first — the most specific test (buildinfo only present in Go binaries).
	if info, ok, _ := ReadGoBinary(exe); ok {
		gbi := info
		return RuntimeGo, &gbi
	}
	// Fall back to executable basename matching for runtimes whose binaries
	// have no portable in-file marker.
	switch runtimeForName(filepath.Base(exe)) {
	case RuntimeNode:
		return RuntimeNode, nil
	case RuntimePython:
		return RuntimePython, nil
	case RuntimeJava:
		return RuntimeJava, nil
	}
	return RuntimeUnknown, nil
}

// runtimeForName classifies a binary by its file basename, case-insensitively.
// Covers Node.js (node, node.exe, nodejs), CPython (python, python3, python3.X,
// pythonw.exe), and Java (java, java.exe, javaw.exe).
//
// Exposed for tests and for callers that want to classify a string without
// hitting the filesystem.
func runtimeForName(base string) Runtime {
	b := strings.ToLower(base)
	// Drop .exe suffix on Windows.
	b = strings.TrimSuffix(b, ".exe")
	switch {
	case b == "node", b == "nodejs":
		return RuntimeNode
	case b == "java", b == "javaw":
		return RuntimeJava
	case b == "python", b == "pythonw", b == "python3", strings.HasPrefix(b, "python3."):
		return RuntimePython
	}
	return RuntimeUnknown
}
