// Package consul provides a read-only HashiCorp Consul CLI for omni.
//
// It is a thin, hand-rolled net/http + encoding/json client against Consul's
// HTTP API. This is deliberate: a read-only MVP needs only a handful of GET
// endpoints, so hand-modeling them adds zero new dependencies and keeps the
// binary lean (see docs/spikes/023-consul-nomad-packer.md). All operations are
// GET-only; write/mutating operations are intentionally out of scope.
package consul

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

const (
	// defaultAddress is used when neither --address nor CONSUL_HTTP_ADDR is set.
	defaultAddress = "http://127.0.0.1:8500"
	// requestTimeout bounds each HTTP request so a hung agent does not block CLI.
	requestTimeout = 30 * time.Second
)

// Options configures a Consul client. Empty fields fall back to environment
// variables (and finally to documented defaults), mirroring how real Consul
// clients read their configuration.
type Options struct {
	Address   string // Consul HTTP address (default: CONSUL_HTTP_ADDR or http://127.0.0.1:8500)
	Token     string // ACL token (default: CONSUL_HTTP_TOKEN), sent as X-Consul-Token
	Namespace string // Consul namespace (default: CONSUL_NAMESPACE), sent as ?ns=
	TLSSkip   bool   // Skip TLS verification
}

// Client is a read-only Consul HTTP client.
type Client struct {
	addr      string
	token     string
	namespace string
	http      *http.Client
}

// Member is one entry from GET /v1/agent/members.
//
// Consul reports member Status as a numeric serf enum, so it is decoded as an
// int and surfaced as a human-readable name via StatusName.
type Member struct {
	Name   string `json:"name"`
	Addr   string `json:"addr"`
	Status int    `json:"status"`
}

// memberStatusNames maps Consul's serf member-status enum to readable names.
var memberStatusNames = map[int]string{
	0: "none",
	1: "alive",
	2: "leaving",
	3: "left",
	4: "failed",
}

// StatusName returns the human-readable name for the member's serf status enum.
func (m Member) StatusName() string {
	if name, ok := memberStatusNames[m.Status]; ok {
		return name
	}

	return fmt.Sprintf("unknown(%d)", m.Status)
}

// classifyConsulError maps an HTTP status code (and/or transport error) to a
// cmderr sentinel, mirroring internal/cli/vault's classifyVaultError shape.
func classifyConsulError(err error, statusCode int, op string) error {
	if err != nil {
		return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("consul: %s: %v", op, err))
	}

	switch statusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		return cmderr.Wrap(cmderr.ErrPermission, fmt.Sprintf("consul: %s: HTTP %d", op, statusCode))
	case http.StatusNotFound:
		return cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("consul: %s: HTTP %d", op, statusCode))
	case http.StatusBadRequest:
		return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("consul: %s: HTTP %d", op, statusCode))
	default:
		return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("consul: %s: HTTP %d", op, statusCode))
	}
}

// New creates a Consul client, resolving env-var defaults for empty options.
func New(opts Options) (*Client, error) {
	addr := opts.Address
	if addr == "" {
		addr = os.Getenv("CONSUL_HTTP_ADDR")
	}
	if addr == "" {
		addr = defaultAddress
	}

	if !strings.Contains(addr, "://") {
		addr = "http://" + addr
	}

	if _, err := url.Parse(addr); err != nil {
		return nil, cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("consul: invalid address %q: %v", addr, err))
	}

	token := opts.Token
	if token == "" {
		token = os.Getenv("CONSUL_HTTP_TOKEN")
	}

	namespace := opts.Namespace
	if namespace == "" {
		namespace = os.Getenv("CONSUL_NAMESPACE")
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
		http:      httpClient,
	}, nil
}

// Address returns the resolved Consul address.
func (c *Client) Address() string { return c.addr }

// get performs a GET request against the given API path and decodes the JSON
// body into out. The path must begin with "/v1/...".
func (c *Client) get(ctx context.Context, path, op string, out any) error {
	u := c.addr + path

	sep := "?"
	if strings.Contains(path, "?") {
		sep = "&"
	}
	if c.namespace != "" {
		u += sep + "ns=" + url.QueryEscape(c.namespace)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return classifyConsulError(err, 0, op)
	}

	if c.token != "" {
		req.Header.Set("X-Consul-Token", c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return classifyConsulError(err, 0, op)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return classifyConsulError(nil, resp.StatusCode, op)
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("consul: %s: decode response: %v", op, err))
		}
	}

	return nil
}

// Members returns the cluster members reported by GET /v1/agent/members.
func (c *Client) Members(ctx context.Context) ([]Member, error) {
	var members []Member
	if err := c.get(ctx, "/v1/agent/members", "members", &members); err != nil {
		return nil, err
	}

	return members, nil
}

// KVGet returns the decoded value for a key from GET /v1/kv/<key>.
func (c *Client) KVGet(ctx context.Context, key string) (string, error) {
	key = strings.TrimPrefix(key, "/")
	if key == "" {
		return "", cmderr.Wrap(cmderr.ErrInvalidInput, "consul: kv get: key is required")
	}

	var entries []struct {
		Key   string `json:"Key"`
		Value string `json:"Value"` // base64-encoded
	}
	if err := c.get(ctx, "/v1/kv/"+key, "kv get "+key, &entries); err != nil {
		return "", err
	}

	if len(entries) == 0 {
		return "", cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("consul: kv get: key not found: %s", key))
	}

	decoded, err := base64.StdEncoding.DecodeString(entries[0].Value)
	if err != nil {
		return "", cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("consul: kv get: decode value: %v", err))
	}

	return string(decoded), nil
}

// Services returns the service-catalog map from GET /v1/catalog/services
// (service name → tags).
func (c *Client) Services(ctx context.Context) (map[string][]string, error) {
	services := map[string][]string{}
	if err := c.get(ctx, "/v1/catalog/services", "services", &services); err != nil {
		return nil, err
	}

	return services, nil
}

// PrintMembers writes members to w as JSON (when asJSON) or an aligned text list.
func PrintMembers(w io.Writer, members []Member, asJSON bool) error {
	if asJSON {
		return writeJSON(w, members)
	}

	_, _ = fmt.Fprintf(w, "%-24s %-22s %s\n", "Node", "Address", "Status")
	for _, m := range members {
		_, _ = fmt.Fprintf(w, "%-24s %-22s %s\n", m.Name, m.Addr, m.StatusName())
	}

	return nil
}

// PrintKV writes a single KV value to w.
func PrintKV(w io.Writer, key, value string, asJSON bool) error {
	if asJSON {
		return writeJSON(w, map[string]string{"key": key, "value": value})
	}

	_, _ = fmt.Fprintln(w, value)

	return nil
}

// PrintServices writes the service→tags map to w.
func PrintServices(w io.Writer, services map[string][]string, asJSON bool) error {
	if asJSON {
		return writeJSON(w, services)
	}

	names := make([]string, 0, len(services))
	for name := range services {
		names = append(names, name)
	}
	sort.Strings(names)

	_, _ = fmt.Fprintf(w, "%-32s %s\n", "Service", "Tags")
	for _, name := range names {
		_, _ = fmt.Fprintf(w, "%-32s %s\n", name, strings.Join(services[name], ","))
	}

	return nil
}

func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("consul: encode json: %v", err))
	}

	return nil
}
