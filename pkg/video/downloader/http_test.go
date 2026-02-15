package downloader

import (
	"bytes"
	"io"
	"testing"
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

// Speed tracker, FormatBytes, FormatPercent tests are in progress_test.go
