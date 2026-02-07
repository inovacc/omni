package encoding

import (
	"bytes"
	"strings"
	"testing"
)

func TestBase64EncodeDecode(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{"simple", []byte("hello world")},
		{"empty", []byte("")},
		{"binary", []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD}},
		{"unicode", []byte("世界🌍")},
		{"long", []byte(strings.Repeat("ABCDEFGHIJ", 1000))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := Base64Encode(tt.input)
			decoded, err := Base64Decode(encoded)
			if err != nil {
				t.Fatalf("Base64Decode() error = %v", err)
			}
			if !bytes.Equal(decoded, tt.input) {
				t.Errorf("Base64 roundtrip failed")
			}
		})
	}
}

func TestBase64DecodeKnown(t *testing.T) {
	decoded, err := Base64Decode("aGVsbG8gd29ybGQ=")
	if err != nil {
		t.Fatalf("Base64Decode() error = %v", err)
	}
	if string(decoded) != "hello world" {
		t.Errorf("Base64Decode() = %v, want 'hello world'", string(decoded))
	}
}

func TestBase64DecodeInvalid(t *testing.T) {
	_, err := Base64Decode("not valid base64!!!")
	if err == nil {
		t.Error("Base64Decode() expected error for invalid input")
	}
}

func TestBase64DecodeWithWhitespace(t *testing.T) {
	// Should strip whitespace before decoding
	decoded, err := Base64Decode("aGVs\nbG8g\nd29y\nbGQ=")
	if err != nil {
		t.Fatalf("Base64Decode() error = %v", err)
	}
	if string(decoded) != "hello world" {
		t.Errorf("Base64Decode() = %v, want 'hello world'", string(decoded))
	}
}

func TestBase32EncodeDecode(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{"simple", []byte("hello")},
		{"empty", []byte("")},
		{"binary", []byte{0x00, 0xFF, 0x7F}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := Base32Encode(tt.input)
			decoded, err := Base32Decode(encoded)
			if err != nil {
				t.Fatalf("Base32Decode() error = %v", err)
			}
			if !bytes.Equal(decoded, tt.input) {
				t.Errorf("Base32 roundtrip failed")
			}
		})
	}
}

func TestBase32DecodeKnown(t *testing.T) {
	decoded, err := Base32Decode("NBSWY3DP")
	if err != nil {
		t.Fatalf("Base32Decode() error = %v", err)
	}
	if string(decoded) != "hello" {
		t.Errorf("Base32Decode() = %v, want 'hello'", string(decoded))
	}
}

func TestBase32DecodeInvalid(t *testing.T) {
	_, err := Base32Decode("!!!invalid!!!")
	if err == nil {
		t.Error("Base32Decode() expected error for invalid input")
	}
}

func TestBase58EncodeDecode(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		encoded string
	}{
		{"simple", []byte("hello"), "Cn8eVZg"},
		{"test", []byte("test"), "3yZe7d"},
		{"longer", []byte("longer test string"), ""},
		{"binary", []byte{1, 2, 3, 4, 5}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := Base58Encode(tt.input)
			if tt.encoded != "" && encoded != tt.encoded {
				t.Errorf("Base58Encode() = %v, want %v", encoded, tt.encoded)
			}

			if len(tt.input) > 0 {
				decoded := Base58Decode(encoded)
				if !bytes.Equal(decoded, tt.input) {
					t.Errorf("Base58 roundtrip failed: got %v, want %v", decoded, tt.input)
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
		{"no wrap needed", "hello", 10, "hello"},
		{"wrap at 5", "helloworld", 5, "hello\nworld"},
		{"zero width", "hello", 0, "hello"},
		{"negative width", "hello", -1, "hello"},
		{"width equals length", "hello", 5, "hello"},
		{"width 1", "abc", 1, "a\nb\nc"},
		{"empty string", "", 10, ""},
		{"large width", "hello", 1000, "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapString(tt.input, tt.width)
			if result != tt.expect {
				t.Errorf("WrapString() = %v, want %v", result, tt.expect)
			}
		})
	}
}
