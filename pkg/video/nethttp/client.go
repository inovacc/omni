package nethttp

import (
	"context"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

const defaultUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

// ClientOptions configures the HTTP client.
type ClientOptions struct {
	Proxy      string
	CookieFile string
	UserAgent  string
	Headers    map[string]string
	Retries    int
	Timeout    time.Duration
}

// Client wraps net/http.Client with video downloader features.
type Client struct {
	http    *http.Client
	ua      string
	headers map[string]string
	retries int
}

// NewClient creates an HTTP client configured for video downloading.
func NewClient(opts ClientOptions) (*Client, error) {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:        100,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	if opts.Proxy != "" {
		proxyURL, err := url.Parse(opts.Proxy)
		if err != nil {
			return nil, fmt.Errorf("nethttp: invalid proxy URL: %w", err)
		}

		transport.Proxy = http.ProxyURL(proxyURL)
	}

	jar, _ := cookiejar.New(nil)

	// Load cookies from file if specified.
	if opts.CookieFile != "" {
		cookies, err := LoadNetscapeCookies(opts.CookieFile)
		if err != nil {
			return nil, fmt.Errorf("nethttp: loading cookies: %w", err)
		}
		// Group cookies by domain.
		domainCookies := make(map[string][]*http.Cookie)
		for _, c := range cookies {
			domainCookies[c.Domain] = append(domainCookies[c.Domain], c)
		}

		for domain, cks := range domainCookies {
			host := strings.TrimPrefix(domain, ".")
			u := &url.URL{Scheme: "https", Host: host}
			jar.SetCookies(u, cks)
		}
	}

	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	ua := opts.UserAgent
	if ua == "" {
		ua = defaultUA
	}

	retries := opts.Retries
	if retries <= 0 {
		retries = 3
	}

	return &Client{
		http: &http.Client{
			Transport: transport,
			Jar:       jar,
			Timeout:   timeout,
		},
		ua:      ua,
		headers: opts.Headers,
		retries: retries,
	}, nil
}

// Do executes an HTTP request with retries.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	c.setDefaults(req)

	var lastErr error

	for attempt := range c.retries {
		resp, err := c.http.Do(req)
		if err == nil {
			if resp.StatusCode >= 500 && attempt < c.retries-1 {
				_ = resp.Body.Close()

				time.Sleep(backoff(attempt))

				continue
			}

			return resp, nil
		}

		lastErr = err

		if attempt < c.retries-1 {
			time.Sleep(backoff(attempt))
		}
	}

	return nil, fmt.Errorf("nethttp: all %d retries failed: %w", c.retries, lastErr)
}

// Get performs an HTTP GET request.
func (c *Client) Get(ctx context.Context, rawURL string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("nethttp: %w", err)
	}

	return c.Do(req)
}

// GetString performs GET and returns the body as a string.
func (c *Client) GetString(ctx context.Context, rawURL string) (string, error) {
	resp, err := c.Get(ctx, rawURL)
	if err != nil {
		return "", err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("nethttp: HTTP %d for %s", resp.StatusCode, rawURL)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("nethttp: reading body: %w", err)
	}

	return string(body), nil
}

// GetJSON performs GET and returns the body as bytes (for json.Unmarshal).
func (c *Client) GetJSON(ctx context.Context, rawURL string) ([]byte, error) {
	resp, err := c.Get(ctx, rawURL)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("nethttp: HTTP %d for %s", resp.StatusCode, rawURL)
	}

	return io.ReadAll(resp.Body)
}

// PostJSON sends a JSON POST request and returns the body.
// Optional extra headers can be passed as key-value pairs.
func (c *Client) PostJSON(ctx context.Context, rawURL string, body io.Reader, extraHeaders ...map[string]string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rawURL, body)
	if err != nil {
		return nil, fmt.Errorf("nethttp: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	for _, headers := range extraHeaders {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	return io.ReadAll(resp.Body)
}

// HTTPClient returns the underlying *http.Client.
func (c *Client) HTTPClient() *http.Client {
	return c.http
}

// CookieJar returns the cookie jar used by this client.
// Extractors use this to read cookies (e.g., SAPISID for authenticated requests).
func (c *Client) CookieJar() http.CookieJar {
	return c.http.Jar
}

func (c *Client) setDefaults(req *http.Request) {
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", c.ua)
	}

	if req.Header.Get("Accept-Language") == "" {
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	}

	for k, v := range c.headers {
		if req.Header.Get(k) == "" {
			req.Header.Set(k, v)
		}
	}
}

func backoff(attempt int) time.Duration {
	d := time.Duration(math.Pow(2, float64(attempt))) * time.Second

	return min(d, 30*time.Second)
}
