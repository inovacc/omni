package command

import (
	"context"
	"io"
)

// AdaptWriterArgs adapts a func(io.Writer, []string) error to Command.
// This covers commands like: RunHash(w io.Writer, args []string, opts HashOptions)
// where opts are pre-bound.
func AdaptWriterArgs(fn func(io.Writer, []string) error) Command {
	return CommandFunc(func(_ context.Context, w io.Writer, _ io.Reader, args []string) error {
		return fn(w, args)
	})
}

// AdaptWriterReaderArgs adapts a func(io.Writer, io.Reader, []string) error to Command.
// This covers commands like: RunHead(w io.Writer, r io.Reader, args []string, opts HeadOptions)
// where opts are pre-bound.
func AdaptWriterReaderArgs(fn func(io.Writer, io.Reader, []string) error) Command {
	return CommandFunc(func(_ context.Context, w io.Writer, r io.Reader, args []string) error {
		return fn(w, r, args)
	})
}

// AdaptFull adapts a func(context.Context, io.Writer, io.Reader, []string) error to Command.
func AdaptFull(fn func(context.Context, io.Writer, io.Reader, []string) error) Command {
	return CommandFunc(fn)
}
