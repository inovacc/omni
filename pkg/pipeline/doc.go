// Package pipeline provides a streaming text processing engine with built-in
// transform stages connected via io.Pipe goroutines. It supports grep, sort,
// uniq, head, tail, cut, tr, sed, and other stages with constant memory usage
// for streaming operations.
package pipeline
