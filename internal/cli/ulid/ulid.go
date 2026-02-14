package ulid

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/inovacc/omni/internal/cli/output"
	"github.com/inovacc/omni/pkg/idgen"
)

// Options configures the ulid command behavior
type Options struct {
	Count        int           // -n: generate N ULIDs
	Lower        bool          // -l: output in lowercase
	OutputFormat output.Format // output format (text, json, table)
}

// Result represents ulid output for JSON
type Result struct {
	ULIDs []string `json:"ulids"`
	Count int      `json:"count"`
}

// ULID represents a Universally Unique Lexicographically Sortable Identifier
type ULID = idgen.ULID

// RunULID generates ULIDs
func RunULID(w io.Writer, opts Options) error {
	if opts.Count <= 0 {
		opts.Count = 1
	}

	f := output.New(w, opts.OutputFormat)

	var ulids []string

	for i := 0; i < opts.Count; i++ {
		u, err := idgen.GenerateULID()
		if err != nil {
			return fmt.Errorf("ulid: %w", err)
		}

		encoded := u.String()
		if opts.Lower {
			encoded = strings.ToLower(encoded)
		}

		if f.IsJSON() {
			ulids = append(ulids, encoded)
		} else {
			_, _ = fmt.Fprintln(w, encoded)
		}
	}

	if f.IsJSON() {
		return f.Print(Result{ULIDs: ulids, Count: len(ulids)})
	}

	return nil
}

// New generates a new ULID
func New() (idgen.ULID, error) {
	return idgen.GenerateULID()
}

// NewWithTime generates a new ULID with the given timestamp
func NewWithTime(t time.Time) (idgen.ULID, error) {
	return idgen.GenerateULIDWithTime(t)
}

// NewString returns a new ULID as a string
func NewString() string {
	return idgen.ULIDString()
}
