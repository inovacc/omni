package video_test

import (
	"testing"

	"github.com/inovacc/omni/pkg/video"
)

func TestNew_API(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	c, err := video.New()
	if err != nil {
		t.Fatalf("video.New() error = %v", err)
	}
	if c == nil {
		t.Fatal("video.New() returned nil client")
	}
}

func TestWithOptions_API(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	c, err := video.New(
		video.WithFormat("best"),
		video.WithQuiet(),
		video.WithNoProgress(),
		video.WithRetries(1),
	)
	if err != nil {
		t.Fatalf("video.New(WithFormat, WithQuiet, ...) error = %v", err)
	}
	if c == nil {
		t.Fatal("video.New() returned nil client")
	}
}
