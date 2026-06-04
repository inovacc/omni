package attest

import (
	"bytes"
	"testing"
)

func TestPAEReferenceVector(t *testing.T) {
	// From the DSSE protocol spec test vectors.
	got := PAE("http://example.com/HelloWorld", []byte("hello world"))
	want := []byte("DSSEv1 29 http://example.com/HelloWorld 11 hello world")
	if !bytes.Equal(got, want) {
		t.Fatalf("PAE mismatch:\n got=%q\nwant=%q", got, want)
	}
}

func TestPAEEmptyBody(t *testing.T) {
	got := PAE("t", []byte{})
	want := []byte("DSSEv1 1 t 0 ")
	if !bytes.Equal(got, want) {
		t.Fatalf("PAE empty body:\n got=%q\nwant=%q", got, want)
	}
}
