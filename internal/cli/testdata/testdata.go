// Package testdata provides test fixtures for JSON, YAML, and TOML tests.
package testdata

import (
	"os"
	"path/filepath"
	"runtime"
)

// Dir returns the absolute path to the testdata directory.
func Dir() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}
	return filepath.Dir(file)
}

// Path returns the absolute path to a file in the testdata directory.
func Path(filename string) string {
	return filepath.Join(Dir(), filename)
}

// Read reads a file from the testdata directory.
func Read(filename string) ([]byte, error) {
	return os.ReadFile(Path(filename))
}

// MustRead reads a file from the testdata directory and panics on error.
func MustRead(filename string) []byte {
	data, err := Read(filename)
	if err != nil {
		panic(err)
	}
	return data
}

// JSON file paths
const (
	SimpleJSON   = "simple.json"
	NestedJSON   = "nested.json"
	ArrayJSON    = "array.json"
	MinifiedJSON = "minified.json"
	TypesJSON    = "types.json"
	UnicodeJSON  = "unicode.json"
	EmptyJSON    = "empty.json"
	InvalidJSON  = "invalid.json"
)

// YAML file paths
const (
	SimpleYAML    = "simple.yaml"
	NestedYAML    = "nested.yaml"
	ArrayYAML     = "array.yaml"
	TypesYAML     = "types.yaml"
	MultilineYAML = "multiline.yaml"
	AnchorsYAML   = "anchors.yaml"
	EmptyYAML     = "empty.yaml"
	InvalidYAML   = "invalid.yaml"
)

// TOML file paths
const (
	SimpleTOML  = "simple.toml"
	NestedTOML  = "nested.toml"
	TypesTOML   = "types.toml"
	ConfigTOML  = "config.toml"
	ArrayTOML   = "array.toml"
	EmptyTOML   = "empty.toml"
	InvalidTOML = "invalid.toml"
)
