package sqlite

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/inovacc/omni/internal/logger"
)

func TestReadOnlyDSN(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		contains []string
	}{
		{
			name:     "plain path",
			path:     "data.db",
			contains: []string{"file:", "mode=ro", "query_only"},
		},
		{
			name:     "path with question mark",
			path:     "weird?name.db",
			contains: []string{"%3F", "mode=ro"},
		},
		{
			name:     "path with hash",
			path:     "tag#1.db",
			contains: []string{"%23", "mode=ro"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsn := readOnlyDSN(tt.path)
			for _, want := range tt.contains {
				if !strings.Contains(dsn, want) {
					t.Errorf("readOnlyDSN(%q) = %q missing %q", tt.path, dsn, want)
				}
			}
		})
	}
}

func TestLogQueryError(t *testing.T) {
	// nil logger → no-op, must not panic.
	logQueryError(nil, "x.db", "SELECT 1", time.Millisecond, errors.New("boom"))

	// Active logger → full branch.
	logPath := filepath.Join(t.TempDir(), "q.log")
	l, err := logger.NewWithExactPath(logPath)
	if err != nil {
		t.Skipf("logger unavailable: %v", err)
	}
	defer func() { _ = l.Close() }()

	if !l.IsActive() {
		t.Skip("logger reports inactive; cannot exercise active branch")
	}
	logQueryError(l, "x.db", "SELECT 1", time.Millisecond, errors.New("boom"))
}
