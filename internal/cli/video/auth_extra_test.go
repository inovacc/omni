package video

import (
	"os"
	"path/filepath"
	"testing"
)

// TestCopyFileIfExists_Missing verifies that a missing source is silently skipped.
func TestCopyFileIfExists_Missing(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "copyfile_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	dst := filepath.Join(tmpDir, "dst.txt")
	err = copyFileIfExists("/nonexistent/source/file.txt", dst)
	if err != nil {
		t.Errorf("copyFileIfExists() missing source should not error, got: %v", err)
	}

	if _, err := os.Stat(dst); !os.IsNotExist(err) {
		t.Error("copyFileIfExists() missing source should not create destination")
	}
}

// TestCopyFileIfExists_Existing verifies that an existing source is copied.
func TestCopyFileIfExists_Existing(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "copyfile_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	src := filepath.Join(tmpDir, "src.txt")
	dst := filepath.Join(tmpDir, "dst.txt")
	content := []byte("hello copy")

	if err := os.WriteFile(src, content, 0644); err != nil {
		t.Fatal(err)
	}

	err = copyFileIfExists(src, dst)
	if err != nil {
		t.Fatalf("copyFileIfExists() error = %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("ReadFile(dst) error = %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("copyFileIfExists() dst content = %q, want %q", got, content)
	}
}

// TestWsEncodeTextFrame_Empty verifies encoding of empty payload.
func TestWsEncodeTextFrame_Empty(t *testing.T) {
	frame := wsEncodeTextFrame([]byte{})
	if len(frame) == 0 {
		t.Fatal("wsEncodeTextFrame() empty payload returned empty frame")
	}
	if frame[0] != 0x81 {
		t.Errorf("wsEncodeTextFrame() empty: first byte = 0x%02x, want 0x81", frame[0])
	}
}

// TestWsExtractPayload_TooShort verifies nil returned for too-short input.
func TestWsExtractPayload_TooShort(t *testing.T) {
	result := wsExtractPayload([]byte{0x81})
	if result != nil {
		t.Errorf("wsExtractPayload() too short should return nil, got: %v", result)
	}
}

// TestWsExtractPayload_EmptyInput verifies nil returned for empty input.
func TestWsExtractPayload_EmptyInput(t *testing.T) {
	result := wsExtractPayload([]byte{})
	if result != nil {
		t.Errorf("wsExtractPayload() empty should return nil, got: %v", result)
	}
}

// TestWsExtractPayload_UnmaskedSmallFrame decodes an unmasked small frame.
func TestWsExtractPayload_UnmaskedSmallFrame(t *testing.T) {
	payload := []byte("hello")
	frame := []byte{
		0x81,
		byte(len(payload)),
	}
	frame = append(frame, payload...)

	result := wsExtractPayload(frame)
	if string(result) != "hello" {
		t.Errorf("wsExtractPayload() = %q, want %q", result, "hello")
	}
}

// TestWsExtractPayload_126LengthTooShort verifies nil for incomplete 126-length frame.
func TestWsExtractPayload_126LengthTooShort(t *testing.T) {
	frame := []byte{0x81, 126, 0x00} // missing second length byte
	result := wsExtractPayload(frame)
	if result != nil {
		t.Errorf("wsExtractPayload() 126-length too short should return nil, got: %v", result)
	}
}

// TestWsExtractPayload_127LengthTooShort verifies nil for incomplete 127-length frame.
func TestWsExtractPayload_127LengthTooShort(t *testing.T) {
	frame := []byte{0x81, 127, 0x00, 0x00} // missing most length bytes
	result := wsExtractPayload(frame)
	if result != nil {
		t.Errorf("wsExtractPayload() 127-length too short should return nil, got: %v", result)
	}
}

// TestWsEncodeDecodeRoundtrip encodes a frame and decodes its payload.
func TestWsEncodeDecodeRoundtrip(t *testing.T) {
	original := []byte(`{"id":1,"method":"Network.getAllCookies"}`)
	frame := wsEncodeTextFrame(original)

	payload := wsExtractPayload(frame)
	if payload == nil {
		t.Fatal("wsExtractPayload() returned nil for valid frame")
	}
	if string(payload) != string(original) {
		t.Errorf("roundtrip payload = %q, want %q", payload, original)
	}
}

// TestWsExtractPayload_NegativePayloadLenNoPanic verifies that a 127-length
// frame whose 64-bit length has the high bit set (which decodes to a negative
// Go int) does not panic with a slice-bounds error and returns nil instead.
// Regression for ws-frame-neg-payloadlen-panic.
func TestWsExtractPayload_NegativePayloadLenNoPanic(t *testing.T) {
	// Byte[1]=0x7F: unmasked, 127-length marker. The 8 length bytes
	// 0xFF..0xF6 set the high bit, so the shift loop yields a negative int.
	frame := []byte{0x01, 0x7F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xF6}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("wsExtractPayload() panicked on negative payload length: %v", r)
		}
	}()

	result := wsExtractPayload(frame)
	if result != nil {
		t.Errorf("wsExtractPayload() negative length should return nil, got: %v", result)
	}
}

// TestWsExtractPayload_OversizedPayloadLen verifies that a frame declaring a
// payload larger than maxWSFrame is rejected (returns nil) rather than
// attempting a huge allocation/slice.
func TestWsExtractPayload_OversizedPayloadLen(t *testing.T) {
	// 127-length frame, 8-byte length = 0x00000000_20000000 (512 MiB),
	// well above maxWSFrame but not negative.
	frame := []byte{0x01, 0x7F, 0x00, 0x00, 0x00, 0x00, 0x20, 0x00, 0x00, 0x00}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("wsExtractPayload() panicked on oversized payload length: %v", r)
		}
	}()

	if result := wsExtractPayload(frame); result != nil {
		t.Errorf("wsExtractPayload() oversized length should return nil, got %d bytes", len(result))
	}
}

// TestIsYouTubeDomain_Extra covers additional cases not in the base test.
func TestIsYouTubeDomain_Extra(t *testing.T) {
	tests := []struct {
		domain string
		want   bool
	}{
		{"music.youtube.com", true},
		{"youtubeee.com", false},
		{"", false},
		{"notyoutube.com", false},
	}
	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			got := isYouTubeDomain(tt.domain)
			if got != tt.want {
				t.Errorf("isYouTubeDomain(%q) = %v, want %v", tt.domain, got, tt.want)
			}
		})
	}
}

// TestIsGoogleDomain_Extra covers additional cases not in the base test.
func TestIsGoogleDomain_Extra(t *testing.T) {
	tests := []struct {
		domain string
		want   bool
	}{
		{"apis.google.com", true},
		{"googlecom", false},
		{"", false},
		{"notgoogle.com", false},
	}
	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			got := isGoogleDomain(tt.domain)
			if got != tt.want {
				t.Errorf("isGoogleDomain(%q) = %v, want %v", tt.domain, got, tt.want)
			}
		})
	}
}

// TestFreePort_ValidRange verifies freePort returns a usable port number.
func TestFreePort_ValidRange(t *testing.T) {
	port, err := freePort()
	if err != nil {
		t.Fatalf("freePort() error = %v", err)
	}
	if port <= 0 || port > 65535 {
		t.Errorf("freePort() = %d, want valid port (1-65535)", port)
	}
}
