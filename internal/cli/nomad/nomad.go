// Package nomad provides a read-only HashiCorp Nomad CLI for omni.
//
// It is a thin, hand-rolled net/http + encoding/json client against Nomad's
// HTTP API. This is deliberate: a read-only MVP needs only a handful of GET
// endpoints, so hand-modeling them adds zero new dependencies and keeps the
// binary lean (see docs/spikes/023-consul-nomad-packer.md). All operations are
// GET-only; write/mutating operations are intentionally out of scope.
package nomad

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

const (
	// defaultAddress is used when neither --address nor NOMAD_ADDR is set.
	defaultAddress = "http://127.0.0.1:4646"
	// requestTimeout bounds each HTTP request so a hung server does not block CLI.
	requestTimeout = 30 * time.Second
)

// Options configures a Nomad client. Empty fields fall back to environment
// variables (and finally to documented defaults), mirroring how real Nomad
// clients read their configuration.
type Options struct {
	Address   string // Nomad HTTP address (default: NOMAD_ADDR or http://127.0.0.1:4646)
	Token     string // ACL token (default: NOMAD_TOKEN), sent as X-Nomad-Token
	Namespace string // Nomad namespace (default: NOMAD_NAMESPACE), sent as ?namespace=
	Region    string // Nomad region (default: NOMAD_REGION), sent as ?region=
	TLSSkip   bool   // Skip TLS verification
}

// Client is a read-only Nomad HTTP client.
type Client struct {
	addr      string
	token     string
	namespace string
	region    string
	http      *http.Client
}

// Job is one entry from GET /v1/jobs.
type Job struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

// Node is one entry from GET /v1/nodes.
type Node struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Drain  bool   `json:"drain"`
}

// Alloc is one entry from GET /v1/allocations.
type Alloc struct {
	ID     string `json:"id"`
	JobID  string `json:"jobID"`
	Status string `json:"status"`
}

// classifyNomadError maps an HTTP status code (and/or transport error) to a
// cmderr sentinel, mirroring internal/cli/vault's classifyVaultError shape.
func classifyNomadError(err error, statusCode int, op string) error {
	if err != nil {
		return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("nomad: %s: %v", op, err))
	}

	switch statusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		return cmderr.Wrap(cmderr.ErrPermission, fmt.Sprintf("nomad: %s: HTTP %d", op, statusCode))
	case http.StatusNotFound:
		return cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("nomad: %s: HTTP %d", op, statusCode))
	case http.StatusBadRequest:
		return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("nomad: %s: HTTP %d", op, statusCode))
	default:
		return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("nomad: %s: HTTP %d", op, statusCode))
	}
}

// New creates a Nomad client, resolving env-var defaults for empty options.
func New(opts Options) (*Client, error) {
	addr := opts.Address
	if addr == "" {
		addr = os.Getenv("NOMAD_ADDR")
	}
	if addr == "" {
		addr = defaultAddress
	}

	if !strings.Contains(addr, "://") {
		addr = "http://" + addr
	}

	if _, err := url.Parse(addr); err != nil {
		return nil, cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("nomad: invalid address %q: %v", addr, err))
	}

	token := opts.Token
	if token == "" {
		token = os.Getenv("NOMAD_TOKEN")
	}

	namespace := opts.Namespace
	if namespace == "" {
		namespace = os.Getenv("NOMAD_NAMESPACE")
	}

	region := opts.Region
	if region == "" {
		region = os.Getenv("NOMAD_REGION")
	}

	httpClient := &http.Client{Timeout: requestTimeout}
	if opts.TLSSkip {
		httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // user opted into --tls-skip-verify
		}
	}

	return &Client{
		addr:      strings.TrimRight(addr, "/"),
		token:     token,
		namespace: namespace,
		region:    region,
		http:      httpClient,
	}, nil
}

// Address returns the resolved Nomad address.
func (c *Client) Address() string { return c.addr }

// get performs a GET request against the given API path and decodes the JSON
// body into out. The path must begin with "/v1/...".
func (c *Client) get(ctx context.Context, path, op string, out any) error {
	q := url.Values{}
	if c.namespace != "" {
		q.Set("namespace", c.namespace)
	}
	if c.region != "" {
		q.Set("region", c.region)
	}

	u := c.addr + path
	if encoded := q.Encode(); encoded != "" {
		u += "?" + encoded
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return classifyNomadError(err, 0, op)
	}

	if c.token != "" {
		req.Header.Set("X-Nomad-Token", c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return classifyNomadError(err, 0, op)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return classifyNomadError(nil, resp.StatusCode, op)
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("nomad: %s: decode response: %v", op, err))
		}
	}

	return nil
}

// JobList returns job stubs from GET /v1/jobs.
func (c *Client) JobList(ctx context.Context) ([]Job, error) {
	var jobs []Job
	if err := c.get(ctx, "/v1/jobs", "job list", &jobs); err != nil {
		return nil, err
	}

	return jobs, nil
}

// NodeList returns node stubs from GET /v1/nodes.
func (c *Client) NodeList(ctx context.Context) ([]Node, error) {
	var nodes []Node
	if err := c.get(ctx, "/v1/nodes", "node list", &nodes); err != nil {
		return nil, err
	}

	return nodes, nil
}

// AllocList returns allocation stubs from GET /v1/allocations.
func (c *Client) AllocList(ctx context.Context) ([]Alloc, error) {
	var allocs []Alloc
	if err := c.get(ctx, "/v1/allocations", "alloc list", &allocs); err != nil {
		return nil, err
	}

	return allocs, nil
}

// PrintJobs writes jobs to w as JSON (when asJSON) or an aligned text table.
func PrintJobs(w io.Writer, jobs []Job, asJSON bool) error {
	if asJSON {
		return writeJSON(w, jobs)
	}

	_, _ = fmt.Fprintf(w, "%-28s %-24s %s\n", "ID", "Name", "Status")
	for _, j := range jobs {
		_, _ = fmt.Fprintf(w, "%-28s %-24s %s\n", j.ID, j.Name, j.Status)
	}

	return nil
}

// PrintNodes writes nodes to w as JSON (when asJSON) or an aligned text table.
func PrintNodes(w io.Writer, nodes []Node, asJSON bool) error {
	if asJSON {
		return writeJSON(w, nodes)
	}

	_, _ = fmt.Fprintf(w, "%-38s %-24s %-12s %s\n", "ID", "Name", "Status", "Drain")
	for _, n := range nodes {
		_, _ = fmt.Fprintf(w, "%-38s %-24s %-12s %t\n", n.ID, n.Name, n.Status, n.Drain)
	}

	return nil
}

// PrintAllocs writes allocations to w as JSON (when asJSON) or an aligned text table.
func PrintAllocs(w io.Writer, allocs []Alloc, asJSON bool) error {
	if asJSON {
		return writeJSON(w, allocs)
	}

	_, _ = fmt.Fprintf(w, "%-38s %-24s %s\n", "ID", "Job", "Status")
	for _, a := range allocs {
		_, _ = fmt.Fprintf(w, "%-38s %-24s %s\n", a.ID, a.JobID, a.Status)
	}

	return nil
}

func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("nomad: encode json: %v", err))
	}

	return nil
}
