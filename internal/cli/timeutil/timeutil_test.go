package timeutil

import (
	"testing"
	"time"
)

func TestNow(t *testing.T) {
	before := time.Now()
	result := Now()
	after := time.Now()

	if result.Before(before) {
		t.Errorf("Now() returned time before the call: %v < %v", result, before)
	}

	if result.After(after) {
		t.Errorf("Now() returned time after the call: %v > %v", result, after)
	}
}
