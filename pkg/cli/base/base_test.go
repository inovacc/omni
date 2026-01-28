package base

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunBase64(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "base64_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("encode", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "input.txt")
		if err := os.WriteFile(testFile, []byte("hello world"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunBase64(&buf, []string{testFile}, BaseOptions{})
		if err != nil {
			t.Fatalf("RunBase64() error = %v", err)
		}

		// "hello world" in base64 is "aGVsbG8gd29ybGQ="
		if !strings.Contains(buf.String(), "aGVsbG8gd29ybGQ=") {
			t.Errorf("RunBase64() got = %v, want base64 of 'hello world'", buf.String())
		}
	})

	t.Run("decode", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "encoded.txt")
		if err := os.WriteFile(testFile, []byte("aGVsbG8gd29ybGQ="), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunBase64(&buf, []string{testFile}, BaseOptions{Decode: true})
		if err != nil {
			t.Fatalf("RunBase64() error = %v", err)
		}

		if buf.String() != "hello world" {
			t.Errorf("RunBase64() got = %v, want 'hello world'", buf.String())
		}
	})
}

func TestRunBase32(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "base32_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("encode", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "input.txt")
		if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunBase32(&buf, []string{testFile}, BaseOptions{})
		if err != nil {
			t.Fatalf("RunBase32() error = %v", err)
		}

		// "hello" in base32 is "NBSWY3DP"
		if !strings.Contains(buf.String(), "NBSWY3DP") {
			t.Errorf("RunBase32() got = %v, want base32 of 'hello'", buf.String())
		}
	})

	t.Run("decode", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "encoded.txt")
		if err := os.WriteFile(testFile, []byte("NBSWY3DP"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunBase32(&buf, []string{testFile}, BaseOptions{Decode: true})
		if err != nil {
			t.Fatalf("RunBase32() error = %v", err)
		}

		if buf.String() != "hello" {
			t.Errorf("RunBase32() got = %v, want 'hello'", buf.String())
		}
	})
}

func TestBase58EncodeDecode(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		encoded string
	}{
		{
			name:    "simple",
			input:   []byte("hello"),
			encoded: "Cn8eVZg",
		},
		{
			name:    "empty",
			input:   []byte{},
			encoded: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := base58Encode(tt.input)
			if encoded != tt.encoded {
				t.Errorf("base58Encode() = %v, want %v", encoded, tt.encoded)
			}

			if len(tt.input) > 0 {
				decoded, err := base58Decode(encoded)
				if err != nil {
					t.Fatalf("base58Decode() error = %v", err)
				}

				if !bytes.Equal(decoded, tt.input) {
					t.Errorf("base58Decode() = %v, want %v", decoded, tt.input)
				}
			}
		})
	}
}

func TestWrapString(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		width  int
		expect string
	}{
		{
			name:   "no wrap needed",
			input:  "hello",
			width:  10,
			expect: "hello",
		},
		{
			name:   "wrap at 5",
			input:  "helloworld",
			width:  5,
			expect: "hello\nworld",
		},
		{
			name:   "zero width",
			input:  "hello",
			width:  0,
			expect: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wrapString(tt.input, tt.width)
			if result != tt.expect {
				t.Errorf("wrapString() = %v, want %v", result, tt.expect)
			}
		})
	}
}
