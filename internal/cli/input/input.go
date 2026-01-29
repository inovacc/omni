package input

import (
	"fmt"
	"io"
	"os"
)

// Source represents an input source (file or reader)
type Source struct {
	Reader io.Reader
	Name   string
	Closer func() error
}

// Close closes the source if it has a closer function
func (s *Source) Close() error {
	if s.Closer != nil {
		return s.Closer()
	}
	return nil
}

// Open returns Sources from args (files or defaultReader if empty/"-")
// If args is empty, returns a single source using defaultReader with name "standard input"
// If an arg is "-", uses defaultReader for that position
// Otherwise opens the file
func Open(args []string, defaultReader io.Reader) ([]Source, error) {
	if len(args) == 0 {
		return []Source{{
			Reader: defaultReader,
			Name:   "standard input",
		}}, nil
	}

	sources := make([]Source, 0, len(args))
	for _, arg := range args {
		src, err := openOne(arg, defaultReader)
		if err != nil {
			// Close any already opened sources
			for _, s := range sources {
				_ = s.Close()
			}
			return nil, err
		}
		sources = append(sources, src)
	}

	return sources, nil
}

// OpenOne returns a single Source (first arg or defaultReader if empty/"-")
func OpenOne(args []string, defaultReader io.Reader) (Source, error) {
	if len(args) == 0 || args[0] == "-" {
		return Source{
			Reader: defaultReader,
			Name:   "standard input",
		}, nil
	}

	return openOne(args[0], defaultReader)
}

// openOne opens a single file or returns the defaultReader for "-"
func openOne(path string, defaultReader io.Reader) (Source, error) {
	if path == "-" {
		return Source{
			Reader: defaultReader,
			Name:   "standard input",
		}, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return Source{}, fmt.Errorf("cannot open '%s': %w", path, err)
	}

	return Source{
		Reader: f,
		Name:   path,
		Closer: f.Close,
	}, nil
}

// OpenWithDefault is a convenience function that opens sources,
// defaulting to os.Stdin if no args provided
func OpenWithDefault(args []string) ([]Source, error) {
	return Open(args, os.Stdin)
}

// OpenOneWithDefault is a convenience function that opens a single source,
// defaulting to os.Stdin if no args provided
func OpenOneWithDefault(args []string) (Source, error) {
	return OpenOne(args, os.Stdin)
}

// MustClose closes a source and ignores any error (for defer)
func MustClose(s *Source) {
	_ = s.Close()
}

// CloseAll closes all sources in the slice
func CloseAll(sources []Source) {
	for i := range sources {
		_ = sources[i].Close()
	}
}
