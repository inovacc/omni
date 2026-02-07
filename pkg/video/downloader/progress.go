package downloader

import (
	"fmt"
	"math"
	"time"
)

// SpeedTracker tracks download speed using a sliding window.
type SpeedTracker struct {
	samples    []speedSample
	windowSize int
}

type speedSample struct {
	bytes int64
	time  time.Time
}

// NewSpeedTracker creates a speed tracker with the given window size.
func NewSpeedTracker(windowSize int) *SpeedTracker {
	if windowSize <= 0 {
		windowSize = 10
	}

	return &SpeedTracker{
		samples:    make([]speedSample, 0, windowSize),
		windowSize: windowSize,
	}
}

// Add records a data point.
func (s *SpeedTracker) Add(totalBytes int64) {
	now := time.Now()

	s.samples = append(s.samples, speedSample{bytes: totalBytes, time: now})
	if len(s.samples) > s.windowSize {
		s.samples = s.samples[1:]
	}
}

// Speed returns the current speed in bytes per second.
func (s *SpeedTracker) Speed() *float64 {
	if len(s.samples) < 2 {
		return nil
	}

	first := s.samples[0]
	last := s.samples[len(s.samples)-1]

	elapsed := last.time.Sub(first.time).Seconds()
	if elapsed <= 0 {
		return nil
	}

	speed := float64(last.bytes-first.bytes) / elapsed

	return &speed
}

// ETA returns the estimated time remaining in seconds.
func (s *SpeedTracker) ETA(downloaded, total int64) *float64 {
	speed := s.Speed()
	if speed == nil || *speed <= 0 || total <= 0 {
		return nil
	}

	remaining := float64(total - downloaded)
	if remaining <= 0 {
		zero := 0.0
		return &zero
	}

	eta := remaining / *speed

	return &eta
}

// FormatBytes formats a byte count into a human-readable string.
func FormatBytes(bytes int64) string {
	if bytes < 0 {
		return "unknown"
	}

	units := []string{"B", "KiB", "MiB", "GiB", "TiB"}
	size := float64(bytes)

	unit := 0
	for size >= 1024 && unit < len(units)-1 {
		size /= 1024
		unit++
	}

	if unit == 0 {
		return fmt.Sprintf("%d%s", bytes, units[unit])
	}

	return fmt.Sprintf("%.2f%s", size, units[unit])
}

// FormatSpeed formats bytes per second into a human-readable string.
func FormatSpeed(speed *float64) string {
	if speed == nil {
		return "unknown"
	}

	return FormatBytes(int64(*speed)) + "/s"
}

// FormatETA formats seconds into HH:MM:SS.
func FormatETA(eta *float64) string {
	if eta == nil {
		return "unknown"
	}

	secs := int(math.Round(*eta))
	if secs < 0 {
		return "unknown"
	}

	h := secs / 3600
	m := (secs % 3600) / 60
	s := secs % 60

	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}

	return fmt.Sprintf("%02d:%02d", m, s)
}

// FormatPercent formats a download progress percentage.
func FormatPercent(downloaded, total int64) string {
	if total <= 0 {
		return "unknown"
	}

	pct := float64(downloaded) / float64(total) * 100

	return fmt.Sprintf("%.1f%%", pct)
}
