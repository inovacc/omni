package downloader

import (
	"strings"
	"testing"
	"time"
)

func TestNewSpeedTracker(t *testing.T) {
	tests := []struct {
		name       string
		windowSize int
		wantSize   int
	}{
		{"positive window", 5, 5},
		{"zero defaults to 10", 0, 10},
		{"negative defaults to 10", -1, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := NewSpeedTracker(tt.windowSize)
			if st.windowSize != tt.wantSize {
				t.Errorf("windowSize = %d, want %d", st.windowSize, tt.wantSize)
			}
		})
	}
}

func TestSpeedTracker_Speed(t *testing.T) {
	t.Run("nil with zero samples", func(t *testing.T) {
		st := NewSpeedTracker(10)
		if st.Speed() != nil {
			t.Error("expected nil speed with no samples")
		}
	})

	t.Run("nil with one sample", func(t *testing.T) {
		st := NewSpeedTracker(10)
		st.Add(100)
		if st.Speed() != nil {
			t.Error("expected nil speed with one sample")
		}
	})

	t.Run("positive with multiple samples", func(t *testing.T) {
		st := NewSpeedTracker(10)
		// Manually inject samples with known timestamps to get deterministic speed.
		now := time.Now()
		st.samples = []speedSample{
			{bytes: 0, time: now},
			{bytes: 1000, time: now.Add(time.Second)},
		}
		speed := st.Speed()
		if speed == nil {
			t.Fatal("expected non-nil speed")
		}
		if *speed < 999 || *speed > 1001 {
			t.Errorf("speed = %f, want ~1000", *speed)
		}
	})
}

func TestSpeedTracker_ETA(t *testing.T) {
	now := time.Now()

	t.Run("valid eta", func(t *testing.T) {
		st := NewSpeedTracker(10)
		st.samples = []speedSample{
			{bytes: 0, time: now},
			{bytes: 500, time: now.Add(time.Second)},
		}
		eta := st.ETA(500, 1000)
		if eta == nil {
			t.Fatal("expected non-nil ETA")
		}
		if *eta < 0.9 || *eta > 1.1 {
			t.Errorf("ETA = %f, want ~1.0", *eta)
		}
	})

	t.Run("completed download", func(t *testing.T) {
		st := NewSpeedTracker(10)
		st.samples = []speedSample{
			{bytes: 0, time: now},
			{bytes: 1000, time: now.Add(time.Second)},
		}
		eta := st.ETA(1000, 1000)
		if eta == nil || *eta != 0 {
			t.Errorf("expected ETA=0 for completed download, got %v", eta)
		}
	})

	t.Run("zero speed returns nil", func(t *testing.T) {
		st := NewSpeedTracker(10)
		// No samples means nil speed.
		eta := st.ETA(0, 1000)
		if eta != nil {
			t.Error("expected nil ETA with no speed data")
		}
	})
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name  string
		bytes int64
		want  string
	}{
		{"zero", 0, "0B"},
		{"small", 512, "512B"},
		{"kib", 1024, "1.00KiB"},
		{"mib", 1048576, "1.00MiB"},
		{"gib", 1073741824, "1.00GiB"},
		{"negative", -1, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatBytes(tt.bytes)
			if got != tt.want {
				t.Errorf("FormatBytes(%d) = %q, want %q", tt.bytes, got, tt.want)
			}
		})
	}
}

func TestFormatSpeed(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		if got := FormatSpeed(nil); got != "unknown" {
			t.Errorf("FormatSpeed(nil) = %q, want %q", got, "unknown")
		}
	})

	t.Run("valid", func(t *testing.T) {
		speed := 1024.0
		got := FormatSpeed(&speed)
		if !strings.HasSuffix(got, "/s") {
			t.Errorf("FormatSpeed(%f) = %q, want suffix '/s'", speed, got)
		}
	})
}

func TestFormatETA(t *testing.T) {
	tests := []struct {
		name string
		eta  *float64
		want string
	}{
		{"nil", nil, "unknown"},
		{"zero", ptr(0.0), "00:00"},
		{"65 seconds", ptr(65.0), "01:05"},
		{"3661 seconds", ptr(3661.0), "1:01:01"},
		{"negative", ptr(-5.0), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatETA(tt.eta)
			if got != tt.want {
				t.Errorf("FormatETA(%v) = %q, want %q", tt.eta, got, tt.want)
			}
		})
	}
}

func TestFormatPercent(t *testing.T) {
	tests := []struct {
		name       string
		downloaded int64
		total      int64
		want       string
	}{
		{"zero total", 0, 0, "unknown"},
		{"half", 50, 100, "50.0%"},
		{"complete", 100, 100, "100.0%"},
		{"negative total", 0, -1, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatPercent(tt.downloaded, tt.total)
			if got != tt.want {
				t.Errorf("FormatPercent(%d, %d) = %q, want %q", tt.downloaded, tt.total, got, tt.want)
			}
		})
	}
}

func ptr(f float64) *float64 { return &f }
