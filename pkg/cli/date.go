package cli

import (
	"fmt"
	"io"
	"time"
)

// DateOptions configures the date command behavior
type DateOptions struct {
	Format string // custom format string
	UTC    bool   // -u: use UTC time
	ISO    bool   // --iso-8601: output date/time in ISO 8601 format
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
