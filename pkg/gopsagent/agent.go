package gopsagent

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
)

// Options configures the embeddable agent.
type Options struct {
	// Addr to listen on (default "127.0.0.1:0" => random port).
	Addr string
	// ConfigDir for the pid file. Default: $HOME/.config/gops.
	// Compatible with the original github.com/inovacc/gops directory layout,
	// so any client of that protocol (including `omni gops`) can discover.
	ConfigDir string
	// AuthKey enables the HMAC challenge when non-empty.
	AuthKey []byte
	// AllowShutdown enables the OpShutdown opcode.
	AllowShutdown bool
	// AllowNonLoopback permits the agent to bind a non-loopback address.
	// By default the agent rejects any Addr that does not resolve to a loopback
	// interface (127.x.x.x / ::1 / localhost) — opcodes are unauthenticated
	// unless AuthKey is set, so exposing them off-host needs explicit opt-in.
	AllowNonLoopback bool
}

// maxConns caps the number of connections served concurrently. Beyond this the
// acceptor blocks before accepting more work, bounding goroutine/FD growth from
// a flood of clients. Kept small because the agent is a loopback-only control
// channel, not a general-purpose server.
const maxConns = 16

// Agent owns the listener + acceptor goroutine.
type Agent struct {
	opts   Options
	ln     net.Listener
	mu     sync.Mutex
	closed bool
	wg     sync.WaitGroup
	// sem is a buffered-channel semaphore bounding concurrently-served
	// connections to maxConns. A slot is acquired before handling a conn and
	// released (defer) when the handler returns.
	sem chan struct{}
}

// New constructs an Agent. Call Listen to start it.
func New(opts Options) *Agent {
	if opts.Addr == "" {
		opts.Addr = "127.0.0.1:0"
	}
	return &Agent{opts: opts, sem: make(chan struct{}, maxConns)}
}

// Addr returns the bound address (host:port) after Listen succeeds. Empty before.
func (a *Agent) Addr() string {
	if a.ln == nil {
		return ""
	}
	return a.ln.Addr().String()
}

// Listen binds the socket, writes the pid → address file, and starts the
// acceptor. Returns immediately; the agent runs until Close.
func (a *Agent) Listen() error {
	ln, err := net.Listen("tcp", a.opts.Addr)
	if err != nil {
		return fmt.Errorf("agent listen: %w", err)
	}
	if !a.opts.AllowNonLoopback {
		host, _, _ := net.SplitHostPort(ln.Addr().String())
		ip := net.ParseIP(host)
		if ip == nil || !ip.IsLoopback() {
			_ = ln.Close()
			return fmt.Errorf("agent: listen address %s is not loopback; set AllowNonLoopback to override", host)
		}
	}
	a.ln = ln
	if err := a.writePIDFile(); err != nil {
		_ = ln.Close()
		return err
	}
	// Startup notification: opt-in via config file or GOPS_AGENT_NOTIFY env var.
	// A malformed config is not fatal — we just skip the notification.
	if cfg, cfgErr := LoadConfig(); cfgErr == nil && notifyEnabled(cfg) {
		fireStartupNotification(a.Addr(), os.Getpid())
	}
	go a.acceptLoop()
	return nil
}

// Close stops accepting, removes the pid file, and waits for in-flight handlers.
// Safe to call multiple times.
func (a *Agent) Close() error {
	a.mu.Lock()
	if a.closed {
		a.mu.Unlock()
		return nil
	}
	a.closed = true
	_ = a.removePIDFile()
	var err error
	if a.ln != nil {
		err = a.ln.Close()
	}
	a.mu.Unlock()
	a.wg.Wait()
	return err
}

func (a *Agent) configDir() string {
	if a.opts.ConfigDir != "" {
		return a.opts.ConfigDir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "gops")
}

func (a *Agent) pidFile() string {
	return filepath.Join(a.configDir(), fmt.Sprintf("%d", os.Getpid()))
}

func (a *Agent) writePIDFile() error {
	dir := a.configDir()
	if dir == "" {
		return errors.New("cannot determine config dir")
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	return os.WriteFile(a.pidFile(), []byte(a.Addr()), 0o600)
}

func (a *Agent) removePIDFile() error {
	return os.Remove(a.pidFile())
}

func (a *Agent) acceptLoop() {
	for {
		// Acquire a semaphore slot before accepting more work so we cap the
		// number of connections served concurrently. When all maxConns slots
		// are in use the acceptor blocks here (clients queue in the kernel
		// backlog) until a handler returns and frees a slot.
		a.sem <- struct{}{}
		conn, err := a.ln.Accept()
		if err != nil {
			<-a.sem
			return
		}
		a.wg.Add(1)
		go func() {
			defer a.wg.Done()
			defer func() { <-a.sem }() // release the slot on handler return
			a.handle(conn)
		}()
	}
}
