package timecmd

import (
	"fmt"
	"io"
	"time"

	"github.com/inovacc/omni/internal/cli/output"
)

// TimeResult represents the timing result of an operation
type TimeResult struct {
	Real     time.Duration `json:"real"`
	User     time.Duration `json:"user"` // Approximated as real for Go
	Sys      time.Duration `json:"sys"`  // Not available in pure Go
	ExitCode int           `json:"exitCode"`
}

// RunTimeJSON measures execution time and outputs in the specified format.
func RunTimeJSON(w io.Writer, format output.Format, fn func() error) (TimeResult, error) {
	start := time.Now()

	err := fn()

	elapsed := time.Since(start)

	result := TimeResult{
		Real:     elapsed,
		User:     elapsed, // Approximation
		Sys:      0,
		ExitCode: 0,
	}

	if err != nil {
		result.ExitCode = 1
	}

	f := output.New(w, format)
	if f.IsJSON() {
		jsonResult := map[string]any{
			"real_ms":   result.Real.Milliseconds(),
			"user_ms":   result.User.Milliseconds(),
			"sys_ms":    result.Sys.Milliseconds(),
			"real":      formatDuration(result.Real),
			"user":      formatDuration(result.User),
			"sys":       formatDuration(result.Sys),
			"exit_code": result.ExitCode,
		}

		_ = f.Print(jsonResult)
	} else {
		_, _ = fmt.Fprintf(w, "\nreal\t%s\n", formatDuration(result.Real))
		_, _ = fmt.Fprintf(w, "user\t%s\n", formatDuration(result.User))
		_, _ = fmt.Fprintf(w, "sys\t%s\n", formatDuration(result.Sys))
	}

	return result, err
}

// RunTime measures execution time of a function
// Since omni doesn't exec external commands, this is used as a timing utility
func RunTime(w io.Writer, fn func() error) (TimeResult, error) {
	start := time.Now()

	err := fn()

	elapsed := time.Since(start)

	result := TimeResult{
		Real:     elapsed,
		User:     elapsed, // Approximation
		Sys:      0,
		ExitCode: 0,
	}

	if err != nil {
		result.ExitCode = 1
	}

	// Print timing info to stderr (like bash time)
	_, _ = fmt.Fprintf(w, "\nreal\t%s\n", formatDuration(result.Real))
	_, _ = fmt.Fprintf(w, "user\t%s\n", formatDuration(result.User))
	_, _ = fmt.Fprintf(w, "sys\t%s\n", formatDuration(result.Sys))

	return result, err
}

// formatDuration formats a duration in the style of bash's time command
func formatDuration(d time.Duration) string {
	minutes := int(d.Minutes())
	seconds := d.Seconds() - float64(minutes*60)

	return fmt.Sprintf("%dm%.3fs", minutes, seconds)
}

// Stopwatch provides a simple timing utility
type Stopwatch struct {
	start time.Time
	laps  []time.Duration
}

// NewStopwatch creates a new stopwatch
func NewStopwatch() *Stopwatch {
	return &Stopwatch{
		start: time.Now(),
	}
}

// Lap records a lap time
func (s *Stopwatch) Lap() time.Duration {
	lap := time.Since(s.start)
	s.laps = append(s.laps, lap)

	return lap
}

// Elapsed returns the total elapsed time
func (s *Stopwatch) Elapsed() time.Duration {
	return time.Since(s.start)
}

// Reset resets the stopwatch
func (s *Stopwatch) Reset() {
	s.start = time.Now()
	s.laps = nil
}

// Laps returns all recorded laps
func (s *Stopwatch) Laps() []time.Duration {
	return s.laps
}

// Sleep pauses execution for the specified duration
func Sleep(d time.Duration) {
	time.Sleep(d)
}

// SleepSeconds pauses execution for the specified number of seconds
func SleepSeconds(seconds float64) {
	time.Sleep(time.Duration(seconds * float64(time.Second)))
}
