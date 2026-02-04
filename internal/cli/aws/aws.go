// Package aws provides AWS CLI functionality for omni.
// It implements core AWS services: S3, EC2, IAM, STS, and SSM.
package aws

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

// Options configures AWS operations
type Options struct {
	Profile string // AWS profile to use
	Region  string // AWS region
	Output  string // Output format: json, text, table (default: json)
	Debug   bool   // Enable debug logging
}

// OutputFormat represents the output format type
type OutputFormat int

const (
	OutputJSON OutputFormat = iota
	OutputText
	OutputTable
)

// ParseOutputFormat parses output format string
func ParseOutputFormat(s string) OutputFormat {
	switch s {
	case "text":
		return OutputText
	case "table":
		return OutputTable
	default:
		return OutputJSON
	}
}

// LoadConfig loads AWS configuration from environment, profile, etc.
func LoadConfig(ctx context.Context, opts Options) (aws.Config, error) {
	var cfgOpts []func(*config.LoadOptions) error

	// Set profile if specified
	if opts.Profile != "" {
		cfgOpts = append(cfgOpts, config.WithSharedConfigProfile(opts.Profile))
	}

	// Set region if specified
	if opts.Region != "" {
		cfgOpts = append(cfgOpts, config.WithRegion(opts.Region))
	}

	// Load configuration
	cfg, err := config.LoadDefaultConfig(ctx, cfgOpts...)
	if err != nil {
		return aws.Config{}, fmt.Errorf("loading AWS config: %w", err)
	}

	return cfg, nil
}

// GetRegion returns the region to use, checking various sources
func GetRegion(opts Options) string {
	if opts.Region != "" {
		return opts.Region
	}

	// Check environment variables
	if region := os.Getenv("AWS_REGION"); region != "" {
		return region
	}
	if region := os.Getenv("AWS_DEFAULT_REGION"); region != "" {
		return region
	}

	// Default region
	return "us-east-1"
}

// Printer handles output formatting
type Printer struct {
	w      io.Writer
	format OutputFormat
}

// NewPrinter creates a new printer
func NewPrinter(w io.Writer, format OutputFormat) *Printer {
	return &Printer{w: w, format: format}
}

// PrintJSON outputs data as JSON
func (p *Printer) PrintJSON(v any) error {
	return PrintJSON(p.w, v)
}

// PrintText outputs data as plain text
func (p *Printer) PrintText(format string, args ...any) {
	_, _ = fmt.Fprintf(p.w, format, args...)
}

// PrintLine outputs a line
func (p *Printer) PrintLine(line string) {
	_, _ = fmt.Fprintln(p.w, line)
}

// PrintJSON outputs data as JSON to a writer
func PrintJSON(w io.Writer, v any) error {
	enc := NewJSONEncoder(w)
	return enc.Encode(v)
}

// PrintTable outputs data as a table (simplified implementation)
func PrintTable(out io.Writer, headers []string, rows [][]string) {
	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print headers
	for i, h := range headers {
		if i > 0 {
			_, _ = fmt.Fprint(out, "  ")
		}
		_, _ = fmt.Fprintf(out, "%-*s", widths[i], h)
	}
	_, _ = fmt.Fprintln(out)

	// Print header separator
	for i, width := range widths {
		if i > 0 {
			_, _ = fmt.Fprint(out, "  ")
		}
		for j := 0; j < width; j++ {
			_, _ = fmt.Fprint(out, "-")
		}
	}
	_, _ = fmt.Fprintln(out)

	// Print rows
	for _, row := range rows {
		for i, cell := range row {
			if i > 0 {
				_, _ = fmt.Fprint(out, "  ")
			}
			if i < len(widths) {
				_, _ = fmt.Fprintf(out, "%-*s", widths[i], cell)
			}
		}
		_, _ = fmt.Fprintln(out)
	}
}
