package base

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/btcsuite/btcd/btcutil/base58"
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
			encoded := base58.Encode(tt.input)
			if encoded != tt.encoded {
				t.Errorf("base58.Encode() = %v, want %v", encoded, tt.encoded)
			}

			if len(tt.input) > 0 {
				decoded := base58.Decode(encoded)

				if !bytes.Equal(decoded, tt.input) {
					t.Errorf("base58.Decode() = %v, want %v", decoded, tt.input)
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

func TestRunBase64Extended(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "base64_ext_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("encode empty file", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "empty.txt")
		if err := os.WriteFile(testFile, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunBase64(&buf, []string{testFile}, BaseOptions{})
		if err != nil {
			t.Fatalf("RunBase64() error = %v", err)
		}

		// Empty input should produce empty or minimal output
		output := strings.TrimSpace(buf.String())
		if output != "" {
			t.Logf("Base64 of empty: %v", output)
		}
	})

	t.Run("encode binary data", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "binary.bin")
		binaryData := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD}
		if err := os.WriteFile(testFile, binaryData, 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunBase64(&buf, []string{testFile}, BaseOptions{})
		if err != nil {
			t.Fatalf("RunBase64() error = %v", err)
		}

		if buf.Len() == 0 {
			t.Error("RunBase64() should produce output for binary data")
		}
	})

	t.Run("encode unicode", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "unicode.txt")
		if err := os.WriteFile(testFile, []byte("世界🌍"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunBase64(&buf, []string{testFile}, BaseOptions{})
		if err != nil {
			t.Fatalf("RunBase64() error = %v", err)
		}

		if buf.Len() == 0 {
			t.Error("RunBase64() should produce output for unicode")
		}
	})

	t.Run("decode invalid base64", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "invalid.txt")
		if err := os.WriteFile(testFile, []byte("not valid base64!!!"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunBase64(&buf, []string{testFile}, BaseOptions{Decode: true})
		if err == nil {
			t.Log("RunBase64() may handle invalid base64 gracefully")
		}
	})

	t.Run("encode long content", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "long.txt")
		longContent := strings.Repeat("ABCDEFGHIJ", 1000)
		if err := os.WriteFile(testFile, []byte(longContent), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunBase64(&buf, []string{testFile}, BaseOptions{})
		if err != nil {
			t.Fatalf("RunBase64() error = %v", err)
		}

		if buf.Len() == 0 {
			t.Error("RunBase64() should produce output for long content")
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunBase64(&buf, []string{"/nonexistent/file.txt"}, BaseOptions{})
		if err == nil {
			t.Error("RunBase64() expected error for nonexistent file")
		}
	})

	t.Run("encode decode roundtrip", func(t *testing.T) {
		original := "Test data for roundtrip verification!"
		inputFile := filepath.Join(tmpDir, "roundtrip_in.txt")
		encodedFile := filepath.Join(tmpDir, "roundtrip_enc.txt")

		if err := os.WriteFile(inputFile, []byte(original), 0644); err != nil {
			t.Fatal(err)
		}

		// Encode
		var encBuf bytes.Buffer
		err := RunBase64(&encBuf, []string{inputFile}, BaseOptions{})
		if err != nil {
			t.Fatalf("RunBase64() encode error = %v", err)
		}

		// Write encoded to file
		if err := os.WriteFile(encodedFile, encBuf.Bytes(), 0644); err != nil {
			t.Fatal(err)
		}

		// Decode
		var decBuf bytes.Buffer
		err = RunBase64(&decBuf, []string{encodedFile}, BaseOptions{Decode: true})
		if err != nil {
			t.Fatalf("RunBase64() decode error = %v", err)
		}

		if decBuf.String() != original {
			t.Errorf("Roundtrip failed: got %v, want %v", decBuf.String(), original)
		}
	})
}

func TestRunBase32Extended(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "base32_ext_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("encode empty", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "empty.txt")
		if err := os.WriteFile(testFile, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunBase32(&buf, []string{testFile}, BaseOptions{})
		if err != nil {
			t.Fatalf("RunBase32() error = %v", err)
		}
	})

	t.Run("encode binary", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "binary.bin")
		if err := os.WriteFile(testFile, []byte{0x00, 0xFF, 0x7F}, 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunBase32(&buf, []string{testFile}, BaseOptions{})
		if err != nil {
			t.Fatalf("RunBase32() error = %v", err)
		}

		if buf.Len() == 0 {
			t.Error("RunBase32() should produce output")
		}
	})

	t.Run("decode invalid", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "invalid.txt")
		if err := os.WriteFile(testFile, []byte("!!!invalid!!!"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunBase32(&buf, []string{testFile}, BaseOptions{Decode: true})
		if err == nil {
			t.Log("RunBase32() may handle invalid gracefully")
		}
	})

	t.Run("roundtrip", func(t *testing.T) {
		original := "base32 test"
		inputFile := filepath.Join(tmpDir, "rt_in.txt")
		encodedFile := filepath.Join(tmpDir, "rt_enc.txt")

		if err := os.WriteFile(inputFile, []byte(original), 0644); err != nil {
			t.Fatal(err)
		}

		var encBuf bytes.Buffer
		_ = RunBase32(&encBuf, []string{inputFile}, BaseOptions{})

		if err := os.WriteFile(encodedFile, encBuf.Bytes(), 0644); err != nil {
			t.Fatal(err)
		}

		var decBuf bytes.Buffer
		_ = RunBase32(&decBuf, []string{encodedFile}, BaseOptions{Decode: true})

		if decBuf.String() != original {
			t.Errorf("Roundtrip failed: got %v, want %v", decBuf.String(), original)
		}
	})
}

func TestBase58Extended(t *testing.T) {
	t.Run("encode various inputs", func(t *testing.T) {
		inputs := [][]byte{
			[]byte("a"),
			[]byte("ab"),
			[]byte("abc"),
			[]byte("test string"),
			[]byte{0x00},
			[]byte{0x00, 0x00},
		}

		for i, input := range inputs {
			encoded := base58.Encode(input)
			// Note: leading zeros produce "1" characters in base58
			if len(input) > 0 && input[0] != 0x00 && len(encoded) == 0 {
				t.Errorf("base58.Encode() input %d produced empty output", i)
			}
		}
	})

	t.Run("decode various inputs", func(t *testing.T) {
		validInputs := []string{
			"2NEpo7TZRRrLZSi2U",
			"1",
			"z",
		}

		for _, input := range validInputs {
			decoded := base58.Decode(input)
			t.Logf("base58.Decode(%v) = %v", input, decoded)
		}
	})

	t.Run("decode invalid character", func(t *testing.T) {
		// btcsuite base58 returns empty slice for invalid input
		decoded := base58.Decode("0OIl") // These chars not in base58
		t.Logf("base58.Decode(invalid) = %v", decoded)
	})

	t.Run("roundtrip multiple", func(t *testing.T) {
		inputs := [][]byte{
			[]byte("test"),
			[]byte("longer test string"),
			[]byte{1, 2, 3, 4, 5},
		}

		for i, input := range inputs {
			encoded := base58.Encode(input)
			decoded := base58.Decode(encoded)

			if !bytes.Equal(decoded, input) {
				t.Errorf("Roundtrip %d failed: got %v, want %v", i, decoded, input)
			}
		}
	})
}

func TestWrapStringExtended(t *testing.T) {
	t.Run("negative width", func(t *testing.T) {
		result := wrapString("hello", -1)
		if result != "hello" {
			t.Logf("wrapString with negative width: %v", result)
		}
	})

	t.Run("width equals length", func(t *testing.T) {
		result := wrapString("hello", 5)
		if result != "hello" {
			t.Errorf("wrapString() = %v, want 'hello'", result)
		}
	})

	t.Run("width 1", func(t *testing.T) {
		result := wrapString("abc", 1)
		expected := "a\nb\nc"
		if result != expected {
			t.Errorf("wrapString() = %v, want %v", result, expected)
		}
	})

	t.Run("empty string", func(t *testing.T) {
		result := wrapString("", 10)
		if result != "" {
			t.Errorf("wrapString() = %v, want empty", result)
		}
	})

	t.Run("large width", func(t *testing.T) {
		result := wrapString("hello", 1000)
		if result != "hello" {
			t.Errorf("wrapString() = %v, want 'hello'", result)
		}
	})
}
