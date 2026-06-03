// Package secret provides Key, a wrapper around sensitive byte material
// (such as decrypted private keys) that must never leak into logs, errors,
// formatted output, or panics. Key implements fmt.Stringer, fmt.GoStringer,
// and slog.LogValuer, all returning a redacted placeholder; the raw bytes are
// reachable only via the explicit Bytes method, and Destroy zeroes them.
//
// This is a stable v1.0 API.
package secret
