package s3

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

// TestS3InvalidURI_IsInvalidInput verifies that supplying a non-S3 URI to an
// operation that requires one is a usage/argument error (exit 2), classified as
// cmderr.ErrInvalidInput. The "not an S3 URI" guard fires before any AWS client
// call, so a zero-value *Client (nil underlying client) is safe here.
func TestS3InvalidURI_IsInvalidInput(t *testing.T) {
	c := &Client{}
	ctx := context.Background()
	var buf bytes.Buffer

	// Rm with a plain (non-s3://) path -> invalid S3 URI usage error.
	err := c.Rm(ctx, &buf, "not-an-s3-uri", false, false, true)
	if !errors.Is(err, cmderr.ErrInvalidInput) {
		t.Fatalf("Rm invalid URI: want ErrInvalidInput, got %v", err)
	}

	// Cp with neither argument an S3 URI -> "at least one argument must be an
	// S3 URI" usage error. Both ParseS3URI calls return IsS3:false before any
	// client use, so the default switch arm is reached.
	err = c.Cp(ctx, &buf, "local-src", "local-dst", CpOptions{})
	if !errors.Is(err, cmderr.ErrInvalidInput) {
		t.Fatalf("Cp no-S3-arg: want ErrInvalidInput, got %v", err)
	}
}
