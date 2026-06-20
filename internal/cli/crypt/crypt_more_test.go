package crypt

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestGetPassword(t *testing.T) {
	dir := t.TempDir()
	pwFile := filepath.Join(dir, "pw.txt")
	if err := os.WriteFile(pwFile, []byte("filepass\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	keyFile := filepath.Join(dir, "key.bin")
	if err := os.WriteFile(keyFile, []byte("keymaterial"), 0o600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		opts    CryptOptions
		env     string
		want    string
		wantErr bool
	}{
		{"direct password", CryptOptions{Password: "secret"}, "", "secret", false},
		{"password file trims newline", CryptOptions{PasswordFile: pwFile}, "", "filepass", false},
		{"key file", CryptOptions{KeyFile: keyFile}, "", "keymaterial", false},
		{"env var", CryptOptions{}, "envpass", "envpass", false},
		{"none", CryptOptions{}, "", "", true},
		{"missing password file", CryptOptions{PasswordFile: filepath.Join(dir, "nope")}, "", "", true},
		{"missing key file", CryptOptions{KeyFile: filepath.Join(dir, "nope")}, "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.env != "" {
				t.Setenv("omni_PASSWORD", tt.env)
			} else {
				t.Setenv("omni_PASSWORD", "")
			}
			got, err := getPassword(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Fatalf("getPassword err=%v wantErr=%v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Fatalf("getPassword=%q want %q", got, tt.want)
			}
		})
	}
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	dir := t.TempDir()
	plainPath := filepath.Join(dir, "plain.txt")
	plaintext := []byte("the quick brown fox\njumps over the lazy dog\n")
	if err := os.WriteFile(plainPath, plaintext, 0o600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		opts CryptOptions
	}{
		{"raw", CryptOptions{Password: "pw1"}},
		{"base64", CryptOptions{Password: "pw2", Base64: true}},
		{"armor with iterations", CryptOptions{Password: "pw3", Armor: true, Iterations: 4096}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt to an output file.
			encPath := filepath.Join(dir, tt.name+".enc")
			encOpts := tt.opts
			encOpts.Output = encPath
			if err := RunEncrypt(&bytes.Buffer{}, []string{plainPath}, encOpts); err != nil {
				t.Fatalf("RunEncrypt: %v", err)
			}

			// Decrypt from the file back to a buffer.
			var dec bytes.Buffer
			if err := RunDecrypt(&dec, []string{encPath}, tt.opts); err != nil {
				t.Fatalf("RunDecrypt: %v", err)
			}
			if !bytes.Equal(dec.Bytes(), plaintext) {
				t.Fatalf("round-trip mismatch: got %q want %q", dec.Bytes(), plaintext)
			}
		})
	}
}

func TestEncryptMissingPassword(t *testing.T) {
	t.Setenv("omni_PASSWORD", "")
	if err := RunEncrypt(&bytes.Buffer{}, []string{"x"}, CryptOptions{}); err == nil {
		t.Fatal("expected error with no password")
	}
	if err := RunDecrypt(&bytes.Buffer{}, []string{"x"}, CryptOptions{}); err == nil {
		t.Fatal("expected error with no password")
	}
}

func TestEncryptMissingInputFile(t *testing.T) {
	t.Setenv("omni_PASSWORD", "")
	err := RunEncrypt(&bytes.Buffer{}, []string{filepath.Join(t.TempDir(), "absent")}, CryptOptions{Password: "p"})
	if err == nil {
		t.Fatal("expected not-found error")
	}
}

func TestDecryptBadCiphertext(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "bad.enc")
	if err := os.WriteFile(bad, []byte("not encrypted data"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := RunDecrypt(&bytes.Buffer{}, []string{bad}, CryptOptions{Password: "p"}); err == nil {
		t.Fatal("expected decrypt failure on garbage")
	}
}

func TestGenerateKeyMore(t *testing.T) {
	var buf bytes.Buffer
	if err := GenerateKey(&buf, 32); err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected key output")
	}
}
