package jsinterp

import (
	"fmt"

	"github.com/dop251/goja"
)

// Interpreter wraps a goja JavaScript runtime for executing JS code.
// Used primarily for YouTube signature decryption.
type Interpreter struct {
	vm *goja.Runtime
}

// New creates a new JavaScript interpreter.
func New() *Interpreter {
	vm := goja.New()
	vm.SetFieldNameMapper(goja.UncapFieldNameMapper())

	return &Interpreter{vm: vm}
}

// Execute runs JavaScript code and returns the result.
func (i *Interpreter) Execute(code string) (goja.Value, error) {
	v, err := i.vm.RunString(code)
	if err != nil {
		return nil, fmt.Errorf("jsinterp: %w", err)
	}

	return v, nil
}

// CallFunction executes JS code that defines functions, then calls
// the named function with the given arguments.
func (i *Interpreter) CallFunction(code, funcName string, args ...any) (goja.Value, error) {
	_, err := i.vm.RunString(code)
	if err != nil {
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

	result, err := fn(goja.Undefined(), gojaArgs...)
	if err != nil {
		return nil, fmt.Errorf("jsinterp: calling %s: %w", funcName, err)
	}

	return result, nil
}

// ExtractFunction extracts a named function from JS code and returns it
// as a callable Go function that takes a string and returns a string.
func (i *Interpreter) ExtractFunction(code, funcName string) (func(string) (string, error), error) {
	_, err := i.vm.RunString(code)
	if err != nil {
		return nil, fmt.Errorf("jsinterp: loading code: %w", err)
	}

	fn, ok := goja.AssertFunction(i.vm.Get(funcName))
	if !ok {
		return nil, fmt.Errorf("jsinterp: %s is not a function", funcName)
	}

	return func(input string) (string, error) {
		result, err := fn(goja.Undefined(), i.vm.ToValue(input))
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
