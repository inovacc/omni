package ksuid

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/inovacc/omni/pkg/idgen"
)

// Options configures the ksuid command behavior
type Options struct {
	Count int  // -n: generate N KSUIDs
	JSON  bool // --json: output as JSON
}

// Result represents ksuid output for JSON
type Result struct {
	KSUIDs []string `json:"ksuids"`
	Count  int      `json:"count"`
}

// KSUID represents a K-Sortable Unique IDentifier
type KSUID = idgen.KSUID

// RunKSUID generates KSUIDs
func RunKSUID(w io.Writer, opts Options) error {
	if opts.Count <= 0 {
		opts.Count = 1
	}

	var ksuids []string

	for i := 0; i < opts.Count; i++ {
		k, err := idgen.GenerateKSUID()
		if err != nil {
			return fmt.Errorf("ksuid: %w", err)
		}

		encoded := k.String()
		if opts.JSON {
			ksuids = append(ksuids, encoded)
		} else {
			_, _ = fmt.Fprintln(w, encoded)
		}
	}

	if opts.JSON {
		return json.NewEncoder(w).Encode(Result{KSUIDs: ksuids, Count: len(ksuids)})
	}

	return nil
}

// New generates a new KSUID
func New() (idgen.KSUID, error) {
	return idgen.GenerateKSUID()
}

// Timestamp returns the time the KSUID was created
func Timestamp(k idgen.KSUID) time.Time {
	return k.Timestamp()
}

// NewString returns a new KSUID as a string
func NewString() string {
	return idgen.KSUIDString()
}
