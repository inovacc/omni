package downloader

import (
	"bytes"
	"io"
	"testing"
	"time"
)

func TestRateLimitedReader(t *testing.T) {
	data := bytes.Repeat([]byte("x"), 1000)
	r := newRateLimitedReader(bytes.NewReader(data), 10000)

	buf := make([]byte, 500)
	n, err := r.Read(buf)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if n != 500 {
		t.Errorf("read %d bytes, want 500", n)
	}
}

func TestRateLimitedReaderFull(t *testing.T) {
	data := []byte("hello world")
	r := newRateLimitedReader(bytes.NewReader(data), 1000000)

	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}

	if string(got) != string(data) {
		t.Errorf("got %q, want %q", got, data)
	}
}

func TestPtrOr(t *testing.T) {
	val := int64(42)
	if ptrOr(&val, 0) != 42 {
		t.Error("expected 42")
	}
	if ptrOr(nil, 99) != 99 {
		t.Error("expected 99")
	}
}

func TestSpeedTracker(t *testing.T) {
	tracker := NewSpeedTracker(10)

	if tracker.Speed() != nil {
		t.Error("expected nil speed with no samples")
	}

	tracker.samples = append(tracker.samples, speedSample{bytes: 0, time: time.Now().Add(-time.Second)})
	tracker.samples = append(tracker.samples, speedSample{bytes: 1000, time: time.Now()})

	speed := tracker.Speed()
	if speed == nil {
		t.Fatal("expected non-nil speed")
	}

	if *speed < 900 || *speed > 1100 {
		t.Errorf("speed = %.0f, want ~1000", *speed)
	}
}

func TestSpeedTrackerETA(t *testing.T) {
	tracker := NewSpeedTracker(10)
	tracker.samples = append(tracker.samples, speedSample{bytes: 0, time: time.Now().Add(-time.Second)})
	tracker.samples = append(tracker.samples, speedSample{bytes: 100, time: time.Now()})

	eta := tracker.ETA(100, 200)
	if eta == nil {
		t.Fatal("expected non-nil ETA")
	}

	if *eta < 0.5 || *eta > 1.5 {
		t.Errorf("ETA = %.1f, want ~1.0", *eta)
	}

	// Already done.
	eta = tracker.ETA(200, 200)
	if eta == nil || *eta != 0 {
		t.Error("expected 0 ETA when done")
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{0, "0B"},
		{1023, "1023B"},
		{1024, "1.00KiB"},
		{1048576, "1.00MiB"},
	}
	for _, tt := range tests {
		got := FormatBytes(tt.input)
		if got != tt.want {
			t.Errorf("FormatBytes(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatPercent(t *testing.T) {
	got := FormatPercent(50, 100)
	if got != "50.0%" {
		t.Errorf("got %q, want 50.0%%", got)
	}

	got = FormatPercent(0, 0)
	if got != "unknown" {
		t.Errorf("got %q, want unknown", got)
	}
}
