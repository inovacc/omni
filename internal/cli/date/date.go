package date

import (
	"fmt"
	"io"
	"time"

	"github.com/inovacc/omni/internal/cli/output"
)

// DateOptions configures the date command behavior
type DateOptions struct {
	Format       string        // custom format string
	UTC          bool          // -u: use UTC time
	ISO          bool          // --iso-8601: output date/time in ISO 8601 format
	OutputFormat output.Format // output format
}

// DateResult represents date output for JSON
type DateResult struct {
	Formatted string `json:"formatted"`
	Unix      int64  `json:"unix"`
	UnixNano  int64  `json:"unix_nano"`
	Year      int    `json:"year"`
	Month     int    `json:"month"`
	Day       int    `json:"day"`
	Hour      int    `json:"hour"`
	Minute    int    `json:"minute"`
	Second    int    `json:"second"`
	Weekday   string `json:"weekday"`
	Timezone  string `json:"timezone"`
	UTC       bool   `json:"utc"`
}

// RunDate prints the current date and time
func RunDate(w io.Writer, opts DateOptions) error {
	now := time.Now()
	if opts.UTC {
		now = now.UTC()
	}

	format := time.RFC3339
	if opts.Format != "" {
		format = opts.Format
	} else if opts.ISO {
		format = "2006-01-02T15:04:05-07:00"
	}

	f := output.New(w, opts.OutputFormat)
	if f.IsJSON() {
		zone, _ := now.Zone()
		result := DateResult{
			Formatted: now.Format(format),
			Unix:      now.Unix(),
			UnixNano:  now.UnixNano(),
			Year:      now.Year(),
			Month:     int(now.Month()),
			Day:       now.Day(),
			Hour:      now.Hour(),
			Minute:    now.Minute(),
			Second:    now.Second(),
			Weekday:   now.Weekday().String(),
			Timezone:  zone,
			UTC:       opts.UTC,
		}

		return f.Print(result)
	}

	_, _ = fmt.Fprintln(w, now.Format(format))

	return nil
}

// Date returns the current time formatted with the given layout
func Date(layout string) string {
	if layout == "" {
		layout = time.RFC3339
	}

	return time.Now().Format(layout)
}
