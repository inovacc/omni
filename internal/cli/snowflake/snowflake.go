package snowflake

import (
	"fmt"
	"io"
	"time"

	"github.com/inovacc/omni/internal/cli/output"
	"github.com/inovacc/omni/pkg/idgen"
)

// Options configures the snowflake command behavior
type Options struct {
	Count        int           // -n: generate N Snowflake IDs
	WorkerID     int64         // -w: worker ID (0-1023)
	OutputFormat output.Format // output format (text, json, table)
}

// Result represents snowflake output for JSON
type Result struct {
	Snowflakes []int64 `json:"snowflakes"`
	Count      int     `json:"count"`
}

// Generator generates Snowflake IDs
type Generator = idgen.SnowflakeGenerator

// RunSnowflake generates Snowflake IDs
func RunSnowflake(w io.Writer, opts Options) error {
	if opts.Count <= 0 {
		opts.Count = 1
	}

	if opts.WorkerID < 0 || opts.WorkerID > 1023 {
		return fmt.Errorf("snowflake: worker ID must be between 0 and 1023")
	}

	gen := idgen.NewSnowflakeGenerator(opts.WorkerID)
	f := output.New(w, opts.OutputFormat)

	var snowflakes []int64

	for i := 0; i < opts.Count; i++ {
		id, err := gen.Generate()
		if err != nil {
			return fmt.Errorf("snowflake: %w", err)
		}

		if f.IsJSON() {
			snowflakes = append(snowflakes, id)
		} else {
			_, _ = fmt.Fprintln(w, id)
		}
	}

	if f.IsJSON() {
		return f.Print(Result{Snowflakes: snowflakes, Count: len(snowflakes)})
	}

	return nil
}

// NewGenerator creates a new Snowflake generator
func NewGenerator(workerID int64) *idgen.SnowflakeGenerator {
	return idgen.NewSnowflakeGenerator(workerID)
}

// New generates a new Snowflake ID using the default generator
func New() (int64, error) {
	return idgen.GenerateSnowflake()
}

// NewString returns a new Snowflake ID as a string
func NewString() string {
	return idgen.SnowflakeString()
}

// Parse extracts components from a Snowflake ID
func Parse(id int64) (timestamp time.Time, workerID int64, sequence int64) {
	return idgen.ParseSnowflake(id)
}
