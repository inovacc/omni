package cmd

import (
	"os"
	"strings"
	"time"

	"github.com/inovacc/omni/internal/cli/curl"
	"github.com/spf13/cobra"
)

var curlCmd = &cobra.Command{
	Use:   "curl [METHOD] URL [ITEM...]",
	Short: "HTTP client with httpie-like syntax",
	Long: `HTTP client inspired by curlie/httpie.

Supports httpie-like syntax for headers and data:
  key:value     HTTP header
  key=value     JSON data field
  key==value    URL query parameter
  @file         Request body from file

Methods: GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS

Examples:
  omni curl https://api.example.com/users
  omni curl POST https://api.example.com/users name=John email=john@example.com
  omni curl https://api.example.com/users Authorization:"Bearer token"
  omni curl https://api.example.com/search q==hello
  omni curl POST https://api.example.com/upload @data.json
  omni curl -v https://api.example.com/users
  omni curl --json https://api.example.com/users`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}

		opts := curl.Options{}
		opts.Verbose, _ = cmd.Flags().GetBool("verbose")
		opts.JSON, _ = cmd.Flags().GetBool("json")
		opts.Form, _ = cmd.Flags().GetBool("form")
		opts.Insecure, _ = cmd.Flags().GetBool("insecure")
		opts.Data, _ = cmd.Flags().GetString("data")

		followRedir, _ := cmd.Flags().GetBool("location")
		opts.FollowRedir = followRedir

		timeoutSec, _ := cmd.Flags().GetInt("timeout")
		opts.Timeout = time.Duration(timeoutSec) * time.Second

		// Parse headers from -H flags
		headerFlags, _ := cmd.Flags().GetStringArray("header")
		if len(headerFlags) > 0 {
			opts.Headers = make(map[string]string)

			for _, h := range headerFlags {
				if idx := strings.Index(h, ":"); idx > 0 {
					opts.Headers[h[:idx]] = strings.TrimSpace(h[idx+1:])
				}
			}
		}

		// Determine method from args
		opts.Method = "GET"

		if len(args) > 0 {
			method := strings.ToUpper(args[0])

			switch method {
			case "GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS":
				opts.Method = method
				args = args[1:]
			default:
				// Check if we have data, default to POST
				for _, arg := range args[1:] {
					if strings.Contains(arg, "=") && !strings.HasPrefix(arg, "http") {
						opts.Method = "POST"
						break
					}
				}
			}
		}

		return curl.Run(cmd.OutOrStdout(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(curlCmd)

	curlCmd.Flags().BoolP("verbose", "v", false, "show request/response details")
	curlCmd.Flags().Bool("json", false, "output response as structured JSON")
	curlCmd.Flags().BoolP("form", "f", false, "send as form data instead of JSON")
	curlCmd.Flags().BoolP("location", "L", true, "follow redirects")
	curlCmd.Flags().BoolP("insecure", "k", false, "skip TLS verification")
	curlCmd.Flags().StringP("data", "d", "", "request body data")
	curlCmd.Flags().StringArrayP("header", "H", nil, "custom header (can be used multiple times)")
	curlCmd.Flags().IntP("timeout", "t", 30, "request timeout in seconds")
}
