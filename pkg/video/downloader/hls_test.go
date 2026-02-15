package downloader

import (
	"crypto/aes"
	"crypto/cipher"
	"testing"

	"github.com/inovacc/omni/pkg/video/m3u8"
)

func TestPkcs7Unpad(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  []byte
	}{
		{"empty", nil, nil},
		{"valid padding 4", append([]byte("hello world!"), 4, 4, 4, 4), []byte("hello world!")},
		{"valid padding 1", append([]byte("hello world!!!!"), 1), []byte("hello world!!!!")},
		{"full block padding", []byte{16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16}, []byte{}},
		{"invalid padding too large", []byte{1, 2, 3, 20}, []byte{1, 2, 3, 20}},
		{"invalid padding mismatch", []byte{1, 2, 3, 2}, []byte{1, 2, 3, 2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pkcs7Unpad(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("len = %d, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("byte[%d] = %d, want %d", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestSelectVariant(t *testing.T) {
	tests := []struct {
		name     string
		playlist *m3u8.Playlist
		baseURL  string
		want     string
	}{
		{
			"empty variants",
			&m3u8.Playlist{Variants: nil},
			"http://example.com/master.m3u8",
			"",
		},
		{
			"single variant",
			&m3u8.Playlist{Variants: []m3u8.Variant{
				{URL: "low.m3u8", Bandwidth: 500000},
			}},
			"http://example.com/master.m3u8",
			"http://example.com/low.m3u8",
		},
		{
			"selects highest bandwidth",
			&m3u8.Playlist{Variants: []m3u8.Variant{
				{URL: "low.m3u8", Bandwidth: 500000},
				{URL: "high.m3u8", Bandwidth: 2000000},
				{URL: "mid.m3u8", Bandwidth: 1000000},
			}},
			"http://example.com/master.m3u8",
			"http://example.com/high.m3u8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := selectVariant(tt.playlist, tt.baseURL)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAES128RoundTrip(t *testing.T) {
	key := []byte("0123456789abcdef") // 16 bytes
	iv := make([]byte, aes.BlockSize)

	plaintext := []byte("hello world!!!!") // 15 bytes
	padLen := aes.BlockSize - (len(plaintext) % aes.BlockSize)
	padded := make([]byte, len(plaintext)+padLen)
	copy(padded, plaintext)
	for i := len(plaintext); i < len(padded); i++ {
		padded[i] = byte(padLen)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}

	encrypted := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(encrypted, padded)

	// Decrypt.
	block2, _ := aes.NewCipher(key)
	decrypted := make([]byte, len(encrypted))
	copy(decrypted, encrypted)
	cipher.NewCBCDecrypter(block2, iv).CryptBlocks(decrypted, decrypted)
	result := pkcs7Unpad(decrypted)

	if string(result) != string(plaintext) {
		t.Errorf("got %q, want %q", result, plaintext)
	}
}
