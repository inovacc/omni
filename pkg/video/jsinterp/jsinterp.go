package jsinterp

import (
	"fmt"
	"time"

	"github.com/dop251/goja"
)

// defaultExecTimeout bounds the wall-clock time any single JavaScript
// evaluation may run before the runtime is interrupted. The player JS is
// scraped from a remote (untrusted) source, so an infinite/expensive loop
// must not be able to hang the process indefinitely.
const defaultExecTimeout = 10 * time.Second

// defaultMaxCallStackSize caps goja's call-stack depth so pathological
// recursion in untrusted JS cannot exhaust the goroutine stack.
const defaultMaxCallStackSize = 2048

// errInterrupted is the sentinel value passed to vm.Interrupt when the
// watchdog fires. It surfaces as a goja *InterruptedError from RunString/fn.
const errInterrupted = "jsinterp: execution timed out"

// Interpreter wraps a goja JavaScript runtime for executing JS code.
// Used primarily for YouTube signature decryption.
type Interpreter struct {
	vm          *goja.Runtime
	execTimeout time.Duration
}

// New creates a new JavaScript interpreter.
func New() *Interpreter {
	vm := goja.New()
	vm.SetFieldNameMapper(goja.UncapFieldNameMapper())
	vm.SetMaxCallStackSize(defaultMaxCallStackSize)

	return &Interpreter{vm: vm, execTimeout: defaultExecTimeout}
}

// SetExecTimeout overrides the per-evaluation wall-clock budget. A value <= 0
// disables the watchdog (no timeout). The default is 10s.
func (i *Interpreter) SetExecTimeout(d time.Duration) {
	i.execTimeout = d
}

// runWithWatchdog runs fn while a timer arms vm.Interrupt after the configured
// budget, guaranteeing that synchronous JS cannot wedge the caller forever.
// goja can only be cancelled via a pre-armed Interrupt, so the watchdog must be
// in place before fn touches the runtime. The interrupt is always cleared on
// return so the runtime stays reusable.
func (i *Interpreter) runWithWatchdog(fn func() (goja.Value, error)) (goja.Value, error) {
	if i.execTimeout <= 0 {
		return fn()
	}

	timer := time.AfterFunc(i.execTimeout, func() {
		i.vm.Interrupt(errInterrupted)
	})
	defer func() {
		timer.Stop()
		i.vm.ClearInterrupt()
	}()

	return fn()
}

// Execute runs JavaScript code and returns the result.
func (i *Interpreter) Execute(code string) (goja.Value, error) {
	v, err := i.runWithWatchdog(func() (goja.Value, error) {
		return i.vm.RunString(code)
	})
	if err != nil {
		return nil, fmt.Errorf("jsinterp: %w", err)
	}

	return v, nil
}

// CallFunction executes JS code that defines functions, then calls
// the named function with the given arguments.
func (i *Interpreter) CallFunction(code, funcName string, args ...any) (goja.Value, error) {
	if _, err := i.runWithWatchdog(func() (goja.Value, error) {
		return i.vm.RunString(code)
	}); err != nil {
		return nil, fmt.Errorf("jsinterp: loading code: %w", err)
	}

	fn, ok := goja.AssertFunction(i.vm.Get(funcName))
	if !ok {
		return nil, fmt.Errorf("jsinterp: %s is not a function", funcName)
	}

	gojaArgs := make([]goja.Value, len(args))
	for idx, arg := range args {
		gojaArgs[idx] = i.vm.ToValue(arg)
	}

	result, err := i.runWithWatchdog(func() (goja.Value, error) {
		return fn(goja.Undefined(), gojaArgs...)
	})
	if err != nil {
		return nil, fmt.Errorf("jsinterp: calling %s: %w", funcName, err)
	}

	return result, nil
}

// ExtractFunction extracts a named function from JS code and returns it
// as a callable Go function that takes a string and returns a string.
func (i *Interpreter) ExtractFunction(code, funcName string) (func(string) (string, error), error) {
	if _, err := i.runWithWatchdog(func() (goja.Value, error) {
		return i.vm.RunString(code)
	}); err != nil {
		return nil, fmt.Errorf("jsinterp: loading code: %w", err)
	}

	fn, ok := goja.AssertFunction(i.vm.Get(funcName))
	if !ok {
		return nil, fmt.Errorf("jsinterp: %s is not a function", funcName)
	}

	return func(input string) (string, error) {
		result, err := i.runWithWatchdog(func() (goja.Value, error) {
			return fn(goja.Undefined(), i.vm.ToValue(input))
		})
		if err != nil {
			return "", fmt.Errorf("jsinterp: %s(%q): %w", funcName, input, err)
		}

		return result.String(), nil
	}, nil
}

// Set sets a global variable in the JS runtime.
func (i *Interpreter) Set(name string, value any) error {
	return i.vm.Set(name, value)
}

// GetString gets a global variable as a string.
func (i *Interpreter) GetString(name string) string {
	v := i.vm.Get(name)
	if v == nil || goja.IsUndefined(v) || goja.IsNull(v) {
		return ""
	}

	return v.String()
}
