package snowflake

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"
)

// Snowflake ID structure (Twitter's Snowflake format):
// - 1 bit: unused (sign bit)
// - 41 bits: timestamp (milliseconds since epoch)
// - 10 bits: machine/worker ID
// - 12 bits: sequence number

const (
	// Epoch is the Snowflake epoch (Twitter's default: Nov 04, 2010)
	// Custom epoch: Jan 01, 2020 00:00:00 UTC
	epoch = 1577836800000 // milliseconds

	// Bit lengths
	timestampBits = 41
	workerIDBits  = 10
	sequenceBits  = 12

	// Max values
	maxWorkerID = (1 << workerIDBits) - 1
	maxSequence = (1 << sequenceBits) - 1

	// Bit shifts
	workerIDShift  = sequenceBits
	timestampShift = sequenceBits + workerIDBits
)

// Options configures the snowflake command behavior
type Options struct {
	Count    int   // -n: generate N Snowflake IDs
	WorkerID int64 // -w: worker ID (0-1023)
	JSON     bool  // --json: output as JSON
}

// Result represents snowflake output for JSON
type Result struct {
	Snowflakes []int64 `json:"snowflakes"`
	Count      int     `json:"count"`
}

// Generator generates Snowflake IDs
type Generator struct {
	mu        sync.Mutex
	workerID  int64
	sequence  int64
	lastTime  int64
}

var defaultGenerator *Generator
var once sync.Once

// RunSnowflake generates Snowflake IDs
func RunSnowflake(w io.Writer, opts Options) error {
	if opts.Count <= 0 {
		opts.Count = 1
	}

	if opts.WorkerID < 0 || opts.WorkerID > maxWorkerID {
		return fmt.Errorf("snowflake: worker ID must be between 0 and %d", maxWorkerID)
	}

	gen := NewGenerator(opts.WorkerID)

	var snowflakes []int64

	for i := 0; i < opts.Count; i++ {
		id, err := gen.Generate()
		if err != nil {
			return fmt.Errorf("snowflake: %w", err)
		}

		if opts.JSON {
			snowflakes = append(snowflakes, id)
		} else {
			_, _ = fmt.Fprintln(w, id)
		}
	}

	if opts.JSON {
		return json.NewEncoder(w).Encode(Result{Snowflakes: snowflakes, Count: len(snowflakes)})
	}

	return nil
}

// NewGenerator creates a new Snowflake generator
func NewGenerator(workerID int64) *Generator {
	return &Generator{
		workerID: workerID & maxWorkerID,
	}
}

// Generate creates a new Snowflake ID
func (g *Generator) Generate() (int64, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now().UnixMilli() - epoch

	if now < g.lastTime {
		return 0, fmt.Errorf("clock moved backwards")
	}

	if now == g.lastTime {
		g.sequence = (g.sequence + 1) & maxSequence
		if g.sequence == 0 {
			// Sequence overflow, wait for next millisecond
			for now <= g.lastTime {
				now = time.Now().UnixMilli() - epoch
			}
		}
	} else {
		g.sequence = 0
	}

	g.lastTime = now

	id := (now << timestampShift) |
		(g.workerID << workerIDShift) |
		g.sequence

	return id, nil
}

// New generates a new Snowflake ID using the default generator
func New() (int64, error) {
	once.Do(func() {
		defaultGenerator = NewGenerator(0)
	})
	return defaultGenerator.Generate()
}

// NewString returns a new Snowflake ID as a string
func NewString() string {
	id, err := New()
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%d", id)
}

// Parse extracts components from a Snowflake ID
func Parse(id int64) (timestamp time.Time, workerID int64, sequence int64) {
	timestamp = time.UnixMilli((id >> timestampShift) + epoch)
	workerID = (id >> workerIDShift) & maxWorkerID
	sequence = id & maxSequence
	return
}
