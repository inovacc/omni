package uuid

import (
	"fmt"
	"io"

	"github.com/inovacc/omni/internal/cli/output"
	"github.com/inovacc/omni/pkg/idgen"
)

// UUIDOptions configures the uuid command behavior
type UUIDOptions struct {
	Count        int           // -n: generate N UUIDs
	Upper        bool          // -u: output in uppercase
	NoDashes     bool          // -x: output without dashes
	Version      int           // -v: UUID version (4 = random, default)
	OutputFormat output.Format // output format (text, json, table)
}

// UUIDResult represents uuid output for JSON
type UUIDResult struct {
	UUIDs []string `json:"uuids"`
	Count int      `json:"count"`
}

// RunUUID generates random UUIDs
func RunUUID(w io.Writer, opts UUIDOptions) error {
	if opts.Count <= 0 {
		opts.Count = 1
	}

	if opts.Version == 0 {
		opts.Version = 4
	}

	var uuidOpts []idgen.UUIDOption

	switch opts.Version {
	case 4:
		uuidOpts = append(uuidOpts, idgen.WithUUIDVersion(idgen.V4))
	case 7:
		uuidOpts = append(uuidOpts, idgen.WithUUIDVersion(idgen.V7))
	default:
		return fmt.Errorf("uuid: unsupported version %d (use 4 or 7)", opts.Version)
	}

	if opts.Upper {
		uuidOpts = append(uuidOpts, idgen.WithUppercase())
	}

	if opts.NoDashes {
		uuidOpts = append(uuidOpts, idgen.WithNoDashes())
	}

	uuids, err := idgen.GenerateUUIDs(opts.Count, uuidOpts...)
	if err != nil {
		return fmt.Errorf("uuid: %w", err)
	}

	f := output.New(w, opts.OutputFormat)

	if f.IsJSON() {
		return f.Print(UUIDResult{UUIDs: uuids, Count: len(uuids)})
	}

	for _, u := range uuids {
		_, _ = fmt.Fprintln(w, u)
	}

	return nil
}

// NewUUIDv7 returns a new time-ordered UUID v7 string
func NewUUIDv7() string {
	u, err := idgen.GenerateUUID(idgen.WithUUIDVersion(idgen.V7))
	if err != nil {
		return ""
	}

	return u
}

// NewUUID returns a new random UUID string
func NewUUID() string {
	u, err := idgen.GenerateUUID()
	if err != nil {
		return ""
	}

	return u
}

// MustNewUUID returns a new random UUID string, panics on error
func MustNewUUID() string {
	u, err := idgen.GenerateUUID()
	if err != nil {
		panic(fmt.Sprintf("uuid: %v", err))
	}

	return u
}

// IsValidUUID checks if a string is a valid UUID format
func IsValidUUID(s string) bool {
	return idgen.IsValidUUID(s)
}
