// Package obfuscate runs light heuristics on a binary to detect garble-style
// Go obfuscation. It returns a structured Verdict with the underlying
// signals (buildinfo present, symbol sections present, mangled main path)
// so callers can render or aggregate.
//
// Adapted from github.com/inovacc/gops (MIT) — see THIRD_PARTY_LICENSES/gops-MIT.txt.
package obfuscate

import (
	"debug/buildinfo"
	"debug/elf"
	"debug/macho"
	"debug/pe"
	"fmt"
	"os"
	"regexp"
	"runtime"
)

// VerdictKind is the high-level classification a Detect call returns.
type VerdictKind string

const (
	VerdictNotGo           VerdictKind = "not_go"
	VerdictClean           VerdictKind = "clean"
	VerdictSuspectStripped VerdictKind = "suspect-stripped"
	VerdictLikelyGarble    VerdictKind = "likely-garble"
)

// Confidence qualifies a non-clean verdict.
type Confidence string

const (
	ConfidenceNone   Confidence = ""
	ConfidenceLow    Confidence = "low"
	ConfidenceMedium Confidence = "medium"
)

// Verdict captures the obfuscation classification + raw signals.
type Verdict struct {
	Path           string      `json:"path"`
	Verdict        VerdictKind `json:"verdict"`
	Confidence     Confidence  `json:"confidence,omitempty"`
	BuildInfoFound bool        `json:"buildinfo"`
	SymbolsFound   bool        `json:"symbols"`
	GoVersion      string      `json:"go_version,omitempty"`
	MainMangled    bool        `json:"main_mangled,omitempty"`
}

// garbleNameRE matches the underscore-prefixed lowercase-alnum 6+ chars
// fragment that garble produces in package and main paths
// (e.g. "/_a1b2c3/" or "_xyzpdq.").
var garbleNameRE = regexp.MustCompile(`(^|[/\.])_[a-z0-9]{6,}([/\.]|$)`)

// Detect runs the heuristics on a binary at path. Returns a Verdict whose
// .Verdict field is one of: not_go, clean, suspect-stripped, likely-garble.
// Returns an error only when the file cannot be read or symbol probing fails;
// non-Go binaries return VerdictNotGo without an error.
func Detect(path string) (Verdict, error) {
	v := Verdict{Path: path}
	if _, err := os.Stat(path); err != nil {
		return v, err
	}
	bi, biErr := buildinfo.ReadFile(path)
	v.BuildInfoFound = biErr == nil
	if biErr == nil {
		v.GoVersion = bi.GoVersion
		if garbleNameRE.MatchString(bi.Main.Path) || garbleNameRE.MatchString(bi.Path) {
			v.MainMangled = true
		}
	}
	symbols, err := hasGoSymbols(path)
	if err != nil {
		return v, fmt.Errorf("symbol check: %w", err)
	}
	v.SymbolsFound = symbols

	// On Windows, Go's PE writer places pclntab inside .rdata rather than a
	// dedicated section, so SymbolsFound may be false even for clean binaries.
	// Treat buildinfo-present + non-mangled as clean on Windows.
	windowsClean := runtime.GOOS == "windows" && v.BuildInfoFound && !v.MainMangled

	switch {
	case !v.BuildInfoFound && !v.SymbolsFound:
		v.Verdict = VerdictNotGo
	case v.MainMangled:
		v.Verdict = VerdictLikelyGarble
		v.Confidence = ConfidenceMedium
	case v.BuildInfoFound && (v.SymbolsFound || windowsClean):
		v.Verdict = VerdictClean
	default:
		v.Verdict = VerdictSuspectStripped
		v.Confidence = ConfidenceLow
	}
	return v, nil
}

// hasGoSymbols returns true if the binary still contains Go symbol/pclntab sections.
//
// Each format-specific parser (elf.NewFile, pe.NewFile, macho.NewFile) wraps f and
// takes ownership of closing the underlying file via its own Close(). We must NOT
// also defer f.Close() — doing so would close the same fd twice, which can recycle
// the descriptor to another goroutine on Windows.
func hasGoSymbols(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	if ef, err := elf.NewFile(f); err == nil {
		defer func() { _ = ef.Close() }()
		for _, s := range ef.Sections {
			if s.Name == ".gopclntab" || s.Name == ".gosymtab" {
				return true, nil
			}
		}
		return false, nil
	}
	if _, err := f.Seek(0, 0); err != nil {
		_ = f.Close()
		return false, err
	}
	if pf, err := pe.NewFile(f); err == nil {
		defer func() { _ = pf.Close() }()
		for _, s := range pf.Sections {
			if s.Name == ".gopclntab" || s.Name == ".symtab" {
				return true, nil
			}
		}
		return false, nil
	}
	if _, err := f.Seek(0, 0); err != nil {
		_ = f.Close()
		return false, err
	}
	if mf, err := macho.NewFile(f); err == nil {
		defer func() { _ = mf.Close() }()
		for _, s := range mf.Sections {
			if s.Name == "__gopclntab" || s.Name == "__gosymtab" {
				return true, nil
			}
		}
		return false, nil
	}
	_ = f.Close()
	return false, nil
}
