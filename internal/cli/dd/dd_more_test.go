package dd

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestPrintDdStats(t *testing.T) {
	start := time.Unix(0, 0)
	stats := DdStats{
		BytesWritten: 2048,
		BlocksIn:     2,
		BlocksOut:    2,
		PartialIn:    1,
		PartialOut:   0,
		StartTime:    start,
		EndTime:      start.Add(2 * time.Second),
	}

	tests := []struct {
		name         string
		showTransfer bool
		wantSubstr   []string
	}{
		{
			name:         "records only",
			showTransfer: false,
			wantSubstr:   []string{"2+1 records in", "2+0 records out"},
		},
		{
			name:         "with transfer",
			showTransfer: true,
			wantSubstr:   []string{"2+1 records in", "2048 bytes transferred", "secs", "/sec)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			printDdStats(&buf, stats, tt.showTransfer)
			out := buf.String()
			for _, s := range tt.wantSubstr {
				if !strings.Contains(out, s) {
					t.Errorf("output %q missing %q", out, s)
				}
			}
		})
	}
}

