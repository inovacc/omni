package nanoid

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/inovacc/omni/pkg/idgen"
)

// Options configures the nanoid command behavior
type Options struct {
	Count    int    // -n: generate N NanoIDs
	Length   int    // -l: length of NanoID (default 21)
	Alphabet string // -a: custom alphabet
	JSON     bool   // --json: output as JSON
}

// Result represents nanoid output for JSON
type Result struct {
	NanoIDs []string `json:"nanoids"`
	Count   int      `json:"count"`
}

// RunNanoID generates NanoIDs
func RunNanoID(w io.Writer, opts Options) error {
	if opts.Count <= 0 {
		opts.Count = 1
	}

	var genOpts []idgen.NanoidOption
	if opts.Length > 0 {
		genOpts = append(genOpts, idgen.WithNanoidLength(opts.Length))
	}
	if opts.Alphabet != "" {
		genOpts = append(genOpts, idgen.WithNanoidAlphabet(opts.Alphabet))
	}

	var nanoids []string

	for i := 0; i < opts.Count; i++ {
		n, err := idgen.GenerateNanoid(genOpts...)
		if err != nil {
			return fmt.Errorf("nanoid: %w", err)
		}

		if opts.JSON {
			nanoids = append(nanoids, n)
		} else {
			_, _ = fmt.Fprintln(w, n)
		}
	}

	if opts.JSON {
		return json.NewEncoder(w).Encode(Result{NanoIDs: nanoids, Count: len(nanoids)})
	}

	return nil
}

// Generate creates a NanoID with custom alphabet and length
func Generate(alphabet string, length int) (string, error) {
	var opts []idgen.NanoidOption
	if alphabet != "" {
		opts = append(opts, idgen.WithNanoidAlphabet(alphabet))
	}
	if length > 0 {
		opts = append(opts, idgen.WithNanoidLength(length))
	}
	return idgen.GenerateNanoid(opts...)
}

// New generates a new NanoID with default settings
func New() (string, error) {
	return idgen.GenerateNanoid()
}

// NewString returns a new NanoID as a string, empty on error
func NewString() string {
	return idgen.NanoidString()
}

// MustNew generates a new NanoID and panics on error
func MustNew() string {
	n, err := idgen.GenerateNanoid()
	if err != nil {
		panic(fmt.Sprintf("nanoid: %v", err))
	}
	return n
}
