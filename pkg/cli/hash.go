package cli

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// HashOptions configures the hash command behavior
type HashOptions struct {
	Algorithm string // md5, sha1, sha256, sha512
	Check     bool   // -c: read checksums from FILE and check them
	Binary    bool   // -b: read in binary mode
	Text      bool   // -t: read in text mode (default)
	Quiet     bool   // --quiet: don't print OK for each verified file
	Status    bool   // --status: don't output anything, status code shows success
	Warn      bool   // -w: warn about improperly formatted checksum lines
	Recursive bool   // -r: hash files recursively in directories
}

// HashResult represents the result of hashing a file
type HashResult struct {
	Path      string `json:"path"`
	Hash      string `json:"hash"`
	Algorithm string `json:"algorithm"`
	Size      int64  `json:"size"`
}

// RunHash computes or verifies file hashes
func RunHash(w io.Writer, args []string, opts HashOptions) error {
	if opts.Algorithm == "" {
		opts.Algorithm = "sha256"
	}

	if opts.Check {
		return verifyChecksums(w, args, opts)
	}

	return computeHashes(w, args, opts)
}

func computeHashes(w io.Writer, args []string, opts HashOptions) error {
	if len(args) == 0 {
		// Read from stdin
		h := getHasher(opts.Algorithm)
		if _, err := io.Copy(h, os.Stdin); err != nil {
			return fmt.Errorf("hash: %w", err)
		}

		_, _ = fmt.Fprintf(w, "%s  -\n", hex.EncodeToString(h.Sum(nil)))

		return nil
	}

	for _, path := range args {
		info, err := os.Stat(path)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "hash: %s: %v\n", path, err)
			continue
		}

		if info.IsDir() {
			if opts.Recursive {
				err = filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
					if err != nil {
						return err
					}

					if !d.IsDir() {
						return hashFile(w, p, opts)
					}

					return nil
				})
				if err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "hash: %v\n", err)
				}
			} else {
				_, _ = fmt.Fprintf(os.Stderr, "hash: %s: Is a directory\n", path)
			}

			continue
		}

		if err := hashFile(w, path, opts); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "hash: %s: %v\n", path, err)
		}
	}

	return nil
}

func hashFile(w io.Writer, path string, opts HashOptions) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	defer func() {
		_ = f.Close()
	}()

	h := getHasher(opts.Algorithm)
	if _, err := io.Copy(h, f); err != nil {
		return err
	}

	sum := hex.EncodeToString(h.Sum(nil))

	mode := " "
	if opts.Binary {
		mode = "*"
	}

	_, _ = fmt.Fprintf(w, "%s %s%s\n", sum, mode, path)

	return nil
}

func verifyChecksums(w io.Writer, args []string, opts HashOptions) error {
	if len(args) == 0 {
		return fmt.Errorf("hash: no checksum file specified")
	}

	var failed, notFound, malformed int

	for _, checksumFile := range args {
		f, err := os.Open(checksumFile)
		if err != nil {
			return fmt.Errorf("hash: %w", err)
		}

		content, err := io.ReadAll(f)
		_ = f.Close()

		if err != nil {
			return fmt.Errorf("hash: %w", err)
		}

		lines := strings.SplitSeq(string(content), "\n")
		for line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			// Parse checksum line: HASH  FILENAME or HASH *FILENAME
			parts := strings.SplitN(line, " ", 2)
			if len(parts) != 2 {
				if opts.Warn {
					_, _ = fmt.Fprintf(os.Stderr, "hash: %s: improperly formatted checksum line\n", line)
				}

				malformed++

				continue
			}

			expectedHash := parts[0]
			filename := strings.TrimLeft(parts[1], " *")

			// Compute actual hash
			actualHash, err := computeFileHash(filename, opts.Algorithm)
			if err != nil {
				if !opts.Status {
					_, _ = fmt.Fprintf(w, "%s: FAILED open or read\n", filename)
				}

				notFound++

				continue
			}

			if actualHash == expectedHash {
				if !opts.Quiet && !opts.Status {
					_, _ = fmt.Fprintf(w, "%s: OK\n", filename)
				}
			} else {
				if !opts.Status {
					_, _ = fmt.Fprintf(w, "%s: FAILED\n", filename)
				}

				failed++
			}
		}
	}

	if failed > 0 || notFound > 0 {
		if !opts.Status {
			if failed > 0 {
				_, _ = fmt.Fprintf(os.Stderr, "hash: WARNING: %d computed checksum did NOT match\n", failed)
			}

			if notFound > 0 {
				_, _ = fmt.Fprintf(os.Stderr, "hash: WARNING: %d listed file could not be read\n", notFound)
			}
		}

		return fmt.Errorf("verification failed")
	}

	return nil
}

func computeFileHash(path string, algorithm string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}

	defer func() {
		_ = f.Close()
	}()

	h := getHasher(algorithm)
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func getHasher(algorithm string) hash.Hash {
	switch strings.ToLower(algorithm) {
	case "md5":
		return md5.New()
	case "sha1":
		return sha1.New()
	case "sha256":
		return sha256.New()
	case "sha512":
		return sha512.New()
	case "sha384":
		return sha512.New384()
	case "sha224":
		return sha256.New224()
	default:
		return sha256.New()
	}
}

// Convenience functions for specific algorithms

// RunMD5Sum computes MD5 hashes (md5sum compatibility)
func RunMD5Sum(w io.Writer, args []string, opts HashOptions) error {
	opts.Algorithm = "md5"
	return RunHash(w, args, opts)
}

// RunSHA1Sum computes SHA1 hashes (sha1sum compatibility)
func RunSHA1Sum(w io.Writer, args []string, opts HashOptions) error {
	opts.Algorithm = "sha1"
	return RunHash(w, args, opts)
}

// RunSHA256Sum computes SHA256 hashes (sha256sum compatibility)
func RunSHA256Sum(w io.Writer, args []string, opts HashOptions) error {
	opts.Algorithm = "sha256"
	return RunHash(w, args, opts)
}

// RunSHA512Sum computes SHA512 hashes (sha512sum compatibility)
func RunSHA512Sum(w io.Writer, args []string, opts HashOptions) error {
	opts.Algorithm = "sha512"
	return RunHash(w, args, opts)
}
