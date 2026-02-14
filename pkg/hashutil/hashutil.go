package hashutil

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"hash/crc32"
	"hash/crc64"
	"io"
	"os"
	"strings"
)

// Algorithm represents a hash algorithm.
type Algorithm string

const (
	MD5    Algorithm = "md5"
	SHA1   Algorithm = "sha1"
	SHA224 Algorithm = "sha224"
	SHA256 Algorithm = "sha256"
	SHA384 Algorithm = "sha384"
	SHA512 Algorithm = "sha512"
	CRC32  Algorithm = "crc32"
	CRC64  Algorithm = "crc64"
)

// HashFile computes the hash of a file at the given path.
func HashFile(path string, algo Algorithm) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("hashutil: %w", err)
	}

	defer func() { _ = f.Close() }()

	return HashReader(f, algo)
}

// HashReader computes the hash of data from an io.Reader.
func HashReader(r io.Reader, algo Algorithm) (string, error) {
	h := newHasher(algo)
	if _, err := io.Copy(h, r); err != nil {
		return "", fmt.Errorf("hashutil: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// HashString computes the hash of a string.
func HashString(s string, algo Algorithm) string {
	h := newHasher(algo)
	_, _ = io.WriteString(h, s)

	return hex.EncodeToString(h.Sum(nil))
}

// HashBytes computes the hash of a byte slice.
func HashBytes(data []byte, algo Algorithm) string {
	h := newHasher(algo)
	_, _ = h.Write(data)

	return hex.EncodeToString(h.Sum(nil))
}

func newHasher(algo Algorithm) hash.Hash {
	switch Algorithm(strings.ToLower(string(algo))) {
	case MD5:
		return md5.New()
	case SHA1:
		return sha1.New()
	case SHA224:
		return sha256.New224()
	case SHA256:
		return sha256.New()
	case SHA384:
		return sha512.New384()
	case SHA512:
		return sha512.New()
	case CRC32:
		return crc32.NewIEEE()
	case CRC64:
		return crc64.New(crc64.MakeTable(crc64.ECMA))
	default:
		return sha256.New()
	}
}
