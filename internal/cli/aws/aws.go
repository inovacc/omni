// Package aws provides AWS CLI functionality for omni.
// It implements core AWS services: S3, EC2, IAM, STS, and SSM.
package aws

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/inovacc/omni/internal/cli/cloud/profile"
)

// Options configures AWS operations
type Options struct {
	Profile     string // AWS profile to use
	Region      string // AWS region
	Output      string // Output format: json, text, table (default: json)
	Debug       bool   // Enable debug logging
	EndpointURL string // Custom endpoint URL (for LocalStack, etc.)
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
// Supports omni cloud profiles via:
//   - OMNI_CLOUD_PROFILE environment variable
//   - --profile omni:name flag prefix
func LoadConfig(ctx context.Context, opts Options) (aws.Config, error) {
	// 1. Check OMNI_CLOUD_PROFILE env var
	if omniProfile := os.Getenv("OMNI_CLOUD_PROFILE"); omniProfile != "" {
		return loadWithOmniProfile(ctx, omniProfile, opts)
	}

	// 2. Check --profile with "omni:" prefix
	if strings.HasPrefix(opts.Profile, "omni:") {
		name := strings.TrimPrefix(opts.Profile, "omni:")
		return loadWithOmniProfile(ctx, name, opts)
	}

	// 3. Fallback to standard AWS SDK config
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

// loadWithOmniProfile loads AWS configuration using an omni cloud profile.
func loadWithOmniProfile(ctx context.Context, name string, opts Options) (aws.Config, error) {
	svc, err := profile.NewService()
	if err != nil {
		return aws.Config{}, fmt.Errorf("initializing profile service: %w", err)
	}

	// Get the profile to check region
	p, err := svc.GetProfile(profile.ProviderAWS, name)
	if err != nil {
		return aws.Config{}, fmt.Errorf("loading omni profile: %w", err)
	}

	// Get credentials
	creds, err := svc.GetAWSCredentials(name)
	if err != nil {
		return aws.Config{}, fmt.Errorf("loading credentials: %w", err)
	}

	// Create static credentials provider
	staticCreds := credentials.NewStaticCredentialsProvider(
		creds.AccessKeyID,
		creds.SecretAccessKey,
		creds.SessionToken,
	)

	// Determine region: explicit opts > profile > default
	region := opts.Region
	if region == "" && p.Region != "" {
		region = p.Region
	}
	if region == "" {
		region = GetRegion(opts)
	}

	// Load config with static credentials
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(staticCreds),
		config.WithRegion(region),
	)
	if err != nil {
		return aws.Config{}, fmt.Errorf("loading AWS config: %w", err)
	}

	return cfg, nil
}

// GetEndpointURL returns the endpoint URL, checking environment variable
func GetEndpointURL(opts Options) string {
	if opts.EndpointURL != "" {
		return opts.EndpointURL
	}

	// Check environment variable (commonly used for LocalStack)
	if endpoint := os.Getenv("AWS_ENDPOINT_URL"); endpoint != "" {
		return endpoint
	}

	return ""
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

// PrintTextf outputs data as plain text with formatting
func (p *Printer) PrintTextf(format string, args ...any) {
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

		for range width {
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
