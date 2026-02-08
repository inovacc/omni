package pipeline

import (
	"io"
	"os"
)

// createFile creates or truncates a file for writing.
// Extracted for testability.
func createFile(path string) (io.WriteCloser, error) {
	return os.Create(path)
}
