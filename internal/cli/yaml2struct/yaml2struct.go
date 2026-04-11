package yaml2struct

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/internal/cli/input"
	"github.com/inovacc/omni/internal/cli/json2struct"
	"gopkg.in/yaml.v3"
)

// wrapInputErr classifies input-reading errors into cmderr sentinels.
func wrapInputErr(cmd string, err error) error {
	if errors.Is(err, os.ErrNotExist) {
		return cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("%s: %s", cmd, err))
	}
	if errors.Is(err, os.ErrPermission) {
		return cmderr.Wrap(cmderr.ErrPermission, fmt.Sprintf("%s: %s", cmd, err))
	}
	return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("%s: %s", cmd, err))
}

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
		return wrapInputErr("yaml2struct", err)
	}
	defer input.CloseAll(sources)

	for _, src := range sources {
		data, err := io.ReadAll(src.Reader)
		if err != nil {
			return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("yaml2struct: read: %s", err))
		}

		// Parse YAML
		var v any
		if err := yaml.Unmarshal(data, &v); err != nil {
			return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("yaml2struct: invalid YAML: %s", err))
		}

		// Convert to JSON (to normalize the data)
		jsonData, err := json.Marshal(v)
		if err != nil {
			return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("yaml2struct: normalize: %s", err))
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
