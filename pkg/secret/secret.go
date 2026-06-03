package secret

import "log/slog"

const placeholder = "<REDACTED>"

// Key holds sensitive bytes that must never appear in logs, errors, %v/%#v, or panics.
type Key struct{ b []byte }

// New wraps raw secret bytes. The caller relinquishes the slice; do not retain it.
func New(b []byte) Key { return Key{b: b} }

// Bytes returns the underlying secret bytes for controlled cryptographic use ONLY.
func (k Key) Bytes() []byte { return k.b }

// String implements fmt.Stringer with a redacted placeholder.
func (k Key) String() string { return placeholder }

// GoString implements fmt.GoStringer so %#v never reveals the bytes.
func (k Key) GoString() string { return "secret.Key{" + placeholder + "}" }

// LogValue implements slog.LogValuer so structured logs never reveal the bytes.
func (k Key) LogValue() slog.Value { return slog.StringValue(placeholder) }

// Destroy zeroes the underlying bytes.
func (k Key) Destroy() {
	for i := range k.b {
		k.b[i] = 0
	}
}
