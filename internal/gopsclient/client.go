package gopsclient

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"time"

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

// Client speaks the gops binary opcode protocol against an agent at addr.
type Client struct {
	addr    string
	timeout time.Duration
}

// NewClient returns a Client with a 5s default timeout. SetTimeout to change.
func NewClient(addr string) *Client { return &Client{addr: addr, timeout: 5 * time.Second} }

// SetTimeout changes the per-connection timeout used by Call/CallProfile/Stream.
func (c *Client) SetTimeout(d time.Duration) { c.timeout = d }

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
	d := net.Dialer{Timeout: c.timeout}
	conn, err := d.DialContext(ctx, "tcp", c.addr)
	if err != nil {
		return fmt.Errorf("dial agent %s: %w", c.addr, err)
	}
	defer func() { _ = conn.Close() }()
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
	_, err = io.Copy(w, conn)
	return err
}
