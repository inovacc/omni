package gopsclient

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/pkg/gopsagent"
)

// Re-export opcode constants so callers within omni don't need a second import.
const (
	OpStack           = gopsagent.OpStack
	OpGC              = gopsagent.OpGC
	OpMemStats        = gopsagent.OpMemStats
	OpVersion         = gopsagent.OpVersion
	OpHeapProfile     = gopsagent.OpHeapProfile
	OpCPUProfile      = gopsagent.OpCPUProfile
	OpStats           = gopsagent.OpStats
	OpTrace           = gopsagent.OpTrace
	OpSetGCPercent    = gopsagent.OpSetGCPercent
	OpRuntimeSnapshot = gopsagent.OpRuntimeSnapshot
	OpMetricsStream   = gopsagent.OpMetricsStream
	OpAuthChallenge   = gopsagent.OpAuthChallenge
	OpShutdown        = gopsagent.OpShutdown
)

// streamIdleTimeout bounds how long Stream waits for the next byte from the
// agent before treating the connection as hung. It mirrors the agent-side
// gopsagent idleReadTimeout (30s): a planted/silent agent that completes the
// handshake but then stops streaming must not be able to block Stream forever
// (CWE-400). The deadline is refreshed after every successful read.
const streamIdleTimeout = 30 * time.Second

// Client speaks the gops binary opcode protocol against an agent at addr.
type Client struct {
	addr             string
	timeout          time.Duration
	idleTimeout      time.Duration
	allowNonLoopback bool
}

// NewClient returns a Client with a 5s default timeout. SetTimeout to change.
func NewClient(addr string) *Client {
	return &Client{addr: addr, timeout: 5 * time.Second, idleTimeout: streamIdleTimeout}
}

// SetTimeout changes the per-connection timeout used by Call/CallProfile/Stream.
func (c *Client) SetTimeout(d time.Duration) { c.timeout = d }

// setIdleTimeout overrides the Stream idle read deadline. Test-only: it exists
// so the idle-timeout regression test can use a short bound without waiting out
// the production streamIdleTimeout.
func (c *Client) setIdleTimeout(d time.Duration) { c.idleTimeout = d }

// SetAllowNonLoopback permits dialing non-loopback addresses. By default the
// client refuses to dial anything but loopback because the target address is
// read from the same-user-writable $HOME/.config/gops/<pid> file: a planted
// file could otherwise redirect omni to an attacker host (SSRF, CWE-918).
// Mirrors gopsagent.Options.AllowNonLoopback so the escape hatch is explicit.
func (c *Client) SetAllowNonLoopback(allow bool) { c.allowNonLoopback = allow }

// validateAddr rejects c.addr unless it is a loopback host, guarding against a
// poisoned discovery file pointing off-host. Empty/garbage addresses and
// non-loopback IPs return cmderr.ErrInvalidInput unless SetAllowNonLoopback was
// called. "localhost" is accepted by name without resolution.
func (c *Client) validateAddr() error {
	if c.allowNonLoopback {
		return nil
	}
	host, _, err := net.SplitHostPort(c.addr)
	if err != nil {
		return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("agent address %q is not host:port", c.addr))
	}
	if host == "localhost" {
		return nil
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("agent address host %q is not an IP or localhost", host))
	}
	if !ip.IsLoopback() {
		return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("agent address %s is not loopback; refusing to dial (set AllowNonLoopback to override)", host))
	}
	return nil
}

// authenticate performs the HMAC challenge handshake when GOPS_AGENT_KEY is set.
// The server emits a 32-byte nonce; the client returns HMAC-SHA256(nonce, key).
// No-op when the env var is empty.
func authenticate(conn net.Conn) error {
	key := os.Getenv("GOPS_AGENT_KEY")
	if key == "" {
		return nil
	}
	var nonce [32]byte
	if _, err := io.ReadFull(conn, nonce[:]); err != nil {
		return fmt.Errorf("agent auth: read nonce: %w", err)
	}
	mac := hmac.New(sha256.New, []byte(key))
	_, _ = mac.Write(nonce[:])
	if _, err := conn.Write(mac.Sum(nil)); err != nil {
		return fmt.Errorf("agent auth: write response: %w", err)
	}
	return nil
}

// Call sends a single opcode and returns the full response (capped at 64 MiB
// to prevent OOM from a rogue agent).
func (c *Client) Call(ctx context.Context, op byte) ([]byte, error) {
	if err := c.validateAddr(); err != nil {
		return nil, err
	}
	d := net.Dialer{Timeout: c.timeout}
	conn, err := d.DialContext(ctx, "tcp", c.addr)
	if err != nil {
		return nil, fmt.Errorf("dial agent %s: %w", c.addr, err)
	}
	defer func() { _ = conn.Close() }()
	hardDeadline := time.Now().Add(c.timeout)
	_ = conn.SetDeadline(hardDeadline)
	if dl, ok := ctx.Deadline(); ok && dl.Before(hardDeadline) {
		_ = conn.SetDeadline(dl)
	}
	if err := authenticate(conn); err != nil {
		return nil, err
	}
	if _, err := conn.Write([]byte{op}); err != nil {
		return nil, err
	}
	const maxResponseBytes = 64 << 20
	return io.ReadAll(io.LimitReader(conn, maxResponseBytes))
}

// CallProfile sends a duration-bearing opcode (OpCPUProfile or OpTrace) and
// streams the response to w. Protocol: 1-byte opcode + 4-byte LE uint32 secs.
// Deadline is duration + 30s grace; data cap 512 MiB.
func (c *Client) CallProfile(ctx context.Context, op byte, dur time.Duration, w io.Writer) error {
	if err := c.validateAddr(); err != nil {
		return err
	}
	d := net.Dialer{Timeout: c.timeout}
	conn, err := d.DialContext(ctx, "tcp", c.addr)
	if err != nil {
		return fmt.Errorf("dial agent %s: %w", c.addr, err)
	}
	defer func() { _ = conn.Close() }()
	deadline := time.Now().Add(dur + 30*time.Second)
	_ = conn.SetDeadline(deadline)
	if err := authenticate(conn); err != nil {
		return err
	}
	secs := uint32(dur.Seconds())
	if secs == 0 {
		secs = 30
	}
	var buf [5]byte
	buf[0] = op
	binary.LittleEndian.PutUint32(buf[1:], secs)
	if _, err := conn.Write(buf[:]); err != nil {
		return err
	}
	const maxBytes = 512 << 20
	_, err = io.Copy(w, io.LimitReader(conn, maxBytes))
	return err
}

// Stream opens OpMetricsStream and copies NDJSON snapshots to w until ctx
// cancels or the connection closes. intervalMs sent as 4-byte LE uint32.
func (c *Client) Stream(ctx context.Context, intervalMs uint32, w io.Writer) error {
	if err := c.validateAddr(); err != nil {
		return err
	}
	d := net.Dialer{Timeout: c.timeout}
	conn, err := d.DialContext(ctx, "tcp", c.addr)
	if err != nil {
		return fmt.Errorf("dial agent %s: %w", c.addr, err)
	}
	defer func() { _ = conn.Close() }()
	// Handshake (auth + opcode write) must complete within the dial timeout so a
	// silent agent cannot stall before streaming even begins.
	_ = conn.SetDeadline(time.Now().Add(c.timeout))
	if err := authenticate(conn); err != nil {
		return err
	}
	var buf [5]byte
	buf[0] = OpMetricsStream
	binary.LittleEndian.PutUint32(buf[1:], intervalMs)
	if _, err := conn.Write(buf[:]); err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		_ = conn.Close()
	}()
	idle := c.idleTimeout
	if idle <= 0 {
		idle = streamIdleTimeout
	}
	// Copy NDJSON frames with a per-read idle deadline (refreshed after every
	// successful read) so a no-data stall is bounded rather than infinite. A
	// pure timeout with no bytes received is reported as cmderr.ErrTimeout; any
	// other read error (incl. ctx cancel closing the conn) is returned as-is.
	rbuf := make([]byte, 32<<10)
	for {
		_ = conn.SetReadDeadline(time.Now().Add(idle))
		n, rerr := conn.Read(rbuf)
		if n > 0 {
			if _, werr := w.Write(rbuf[:n]); werr != nil {
				return werr
			}
		}
		if rerr != nil {
			if rerr == io.EOF {
				return nil
			}
			var nerr net.Error
			if errors.As(rerr, &nerr) && nerr.Timeout() {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				return cmderr.Wrap(cmderr.ErrTimeout, fmt.Sprintf("agent %s idle for %s; stream stalled", c.addr, idle))
			}
			return rerr
		}
	}
}
