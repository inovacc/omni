// Package command defines the unified Command interface for CLI commands.
package command

import (
	"context"
	"io"
	"sort"
	"sync"
)

// Command is the interface all CLI commands should implement.
type Command interface {
	Run(ctx context.Context, w io.Writer, r io.Reader, args []string) error
}

// CommandFunc adapts a function to the Command interface.
type CommandFunc func(ctx context.Context, w io.Writer, r io.Reader, args []string) error

// Run implements Command.
func (f CommandFunc) Run(ctx context.Context, w io.Writer, r io.Reader, args []string) error {
	return f(ctx, w, r, args)
}

// Registry maps command names to Command implementations.
type Registry struct {
	mu   sync.RWMutex
	cmds map[string]Command
}

// NewRegistry creates a new command registry.
func NewRegistry() *Registry {
	return &Registry{
		cmds: make(map[string]Command),
	}
}

// Register adds a command to the registry.
func (r *Registry) Register(name string, cmd Command) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cmds[name] = cmd
}

// Get returns a command by name.
func (r *Registry) Get(name string) (Command, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cmd, ok := r.cmds[name]
	return cmd, ok
}

// Names returns all registered command names sorted alphabetically.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.cmds))
	for name := range r.cmds {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Len returns the number of registered commands.
func (r *Registry) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.cmds)
}
