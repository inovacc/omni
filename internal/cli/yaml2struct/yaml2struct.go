package yaml2struct

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/inovacc/omni/internal/cli/input"
	"github.com/inovacc/omni/internal/cli/json2struct"
	"gopkg.in/yaml.v3"
)

// Options configures the yaml2struct command behavior
type Options struct {
	Name      string // struct name (default: "Root")
	Package   string // package name (default: "main")
	Inline    bool   // inline nested structs
	OmitEmpty bool   // add omitempty to all fields
}

// RunYAML2Struct converts YAML to Go struct definition
func RunYAML2Struct(w io.Writer, r io.Reader, args []string, opts Options) error {
	sources, err := input.Open(args, r)
	if err != nil {
		return fmt.Errorf("yaml2struct: %w", err)
	}
	defer input.CloseAll(sources)

	for _, src := range sources {
		data, err := io.ReadAll(src.Reader)
		if err != nil {
			return fmt.Errorf("yaml2struct: %w", err)
		}

		// Parse YAML
		var v any
		if err := yaml.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("yaml2struct: invalid YAML: %w", err)
		}

		// Convert to JSON (to normalize the data)
		jsonData, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("yaml2struct: %w", err)
		}

		// Use json2struct to generate the Go struct
		jsonOpts := json2struct.Options{
			Name:      opts.Name,
			Package:   opts.Package,
			Inline:    opts.Inline,
			OmitEmpty: opts.OmitEmpty,
		}

		if err := json2struct.RunJSON2Struct(w, strings.NewReader(string(jsonData)), nil, jsonOpts); err != nil {
			return err
		}
	}

	return nil
}
