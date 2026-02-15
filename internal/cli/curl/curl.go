package curl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/inovacc/omni/internal/cli/output"
)

// Options configures the curl command behavior
type Options struct {
	Method       string            // HTTP method
	Headers      map[string]string // Custom headers
	Data         string            // Request body
	Form         bool              // Send as form data
	JSON         bool              // Raw JSON output (no formatting) — deprecated, use OutputFormat
	Verbose      bool              // Show request/response details
	Timeout      time.Duration     // Request timeout
	FollowRedir  bool              // Follow redirects
	Insecure     bool              // Skip TLS verification
	OutputFormat output.Format     // global output format
}

// Response represents the HTTP response
type Response struct {
	Status     string              `json:"status"`
	StatusCode int                 `json:"statusCode"`
	Headers    map[string][]string `json:"headers"`
	Body       string              `json:"body"`
	Duration   float64             `json:"duration_ms"`
}

// Run executes an HTTP request
func Run(w io.Writer, args []string, opts Options) error {
	if len(args) == 0 {
		return fmt.Errorf("curl: URL required")
	}

	// Parse URL and arguments
	urlStr, headers, data, err := parseArgs(args, opts)
	if err != nil {
		return err
	}

	// Merge headers
	for k, v := range headers {
		if opts.Headers == nil {
			opts.Headers = make(map[string]string)
		}

		opts.Headers[k] = v
	}

	// Build request body
	var body io.Reader

	if opts.Data != "" {
		body = strings.NewReader(opts.Data)
	} else if data != "" {
		body = strings.NewReader(data)
	}

	// Create request
	req, err := http.NewRequest(opts.Method, urlStr, body)
	if err != nil {
		return fmt.Errorf("curl: %w", err)
	}

	// Set headers
	for k, v := range opts.Headers {
		req.Header.Set(k, v)
	}

	// Set default Content-Type for POST/PUT/PATCH with data
	if body != nil && req.Header.Get("Content-Type") == "" {
		if opts.Form {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		} else {
			req.Header.Set("Content-Type", "application/json")
		}
	}

	// Set User-Agent
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "omni-curl/1.0")
	}

	// Create client
	client := &http.Client{
		Timeout: opts.Timeout,
	}

	if !opts.FollowRedir {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	// Print request details if verbose
	if opts.Verbose {
		_, _ = fmt.Fprintf(w, "> %s %s\n", req.Method, req.URL.String())

		for k, v := range req.Header {
			_, _ = fmt.Fprintf(w, "> %s: %s\n", k, strings.Join(v, ", "))
		}

		_, _ = fmt.Fprintln(w, ">")
	}

	// Execute request
	start := time.Now()

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("curl: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	duration := time.Since(start)

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("curl: reading response: %w", err)
	}

	// Print response details if verbose
	if opts.Verbose {
		_, _ = fmt.Fprintf(w, "< %s\n", resp.Status)

		for k, v := range resp.Header {
			_, _ = fmt.Fprintf(w, "< %s: %s\n", k, strings.Join(v, ", "))
		}

		_, _ = fmt.Fprintln(w, "<")
	}

	// Output response
	useJSON := opts.JSON || opts.OutputFormat == output.FormatJSON
	if useJSON {
		// Structured JSON output mode
		response := Response{
			Status:     resp.Status,
			StatusCode: resp.StatusCode,
			Headers:    resp.Header,
			Body:       string(respBody),
			Duration:   float64(duration.Milliseconds()),
		}

		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")

		return enc.Encode(response)
	}

	// Pretty print JSON response if Content-Type is JSON
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, respBody, "", "  "); err == nil {
			_, _ = fmt.Fprintln(w, prettyJSON.String())

			return nil
		}
	}

	// Output raw body
	_, _ = fmt.Fprintln(w, string(respBody))

	return nil
}

// parseArgs parses command arguments for URL, headers, and data
// Supports httpie-like syntax:
//   - key:value for headers
//   - key=value for JSON data
//   - key==value for URL parameters
func parseArgs(args []string, opts Options) (string, map[string]string, string, error) {
	var urlStr string

	headers := make(map[string]string)
	data := make(map[string]any)
	params := make(map[string]string)

	for _, arg := range args {
		// Check for header (key:value)
		if idx := strings.Index(arg, ":"); idx > 0 && !strings.HasPrefix(arg, "http") {
			// Make sure it's not a URL
			if !strings.Contains(arg[:idx], "/") && !strings.Contains(arg[:idx], "=") {
				key := arg[:idx]
				value := strings.TrimSpace(arg[idx+1:])
				headers[key] = value

				continue
			}
		}

		// Check for URL parameter (key==value)
		if idx := strings.Index(arg, "=="); idx > 0 {
			key := arg[:idx]
			value := arg[idx+2:]
			params[key] = value

			continue
		}

		// Check for data (key=value)
		if idx := strings.Index(arg, "="); idx > 0 && !strings.HasPrefix(arg, "http") {
			key := arg[:idx]
			value := arg[idx+1:]

			// Try to parse as JSON value
			var jsonVal any
			if err := json.Unmarshal([]byte(value), &jsonVal); err == nil {
				data[key] = jsonVal
			} else {
				data[key] = value
			}

			continue
		}

		// Check for file upload (@file)
		if strings.HasPrefix(arg, "@") {
			filename := arg[1:]

			content, err := os.ReadFile(filename)
			if err != nil {
				return "", nil, "", fmt.Errorf("curl: reading file %s: %w", filename, err)
			}

			return urlStr, headers, string(content), nil
		}

		// Must be URL
		if urlStr == "" {
			urlStr = arg
		}
	}

	if urlStr == "" {
		return "", nil, "", fmt.Errorf("curl: URL required")
	}

	// Add URL parameters
	if len(params) > 0 {
		parsedURL, err := url.Parse(urlStr)
		if err != nil {
			return "", nil, "", fmt.Errorf("curl: invalid URL: %w", err)
		}

		q := parsedURL.Query()

		for k, v := range params {
			q.Set(k, v)
		}

		parsedURL.RawQuery = q.Encode()
		urlStr = parsedURL.String()
	}

	// Convert data map to JSON
	var dataStr string

	if len(data) > 0 {
		if opts.Form {
			// Form encoding
			values := url.Values{}

			for k, v := range data {
				values.Set(k, fmt.Sprintf("%v", v))
			}

			dataStr = values.Encode()
		} else {
			// JSON encoding
			jsonData, err := json.Marshal(data)
			if err != nil {
				return "", nil, "", fmt.Errorf("curl: encoding data: %w", err)
			}

			dataStr = string(jsonData)
		}
	}

	return urlStr, headers, dataStr, nil
}
