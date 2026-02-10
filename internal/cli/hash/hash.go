package hash

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/inovacc/omni/pkg/hashutil"
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
	JSON      bool   // --json: output as JSON
}

// HashResult represents the result of hashing a file
type HashResult struct {
	Path      string `json:"path"`
	Hash      string `json:"hash"`
	Algorithm string `json:"algorithm"`
	Size      int64  `json:"size"`
}

// HashesResult represents the JSON output for hashes
type HashesResult struct {
	Hashes []HashResult `json:"hashes"`
	Count  int          `json:"count"`
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
	algo := hashutil.Algorithm(opts.Algorithm)
	var results []HashResult

	if len(args) == 0 {
		// Read from stdin
		hashStr, err := hashutil.HashReader(os.Stdin, algo)
		if err != nil {
			return fmt.Errorf("hash: %w", err)
		}

		if opts.JSON {
			results = append(results, HashResult{Path: "-", Hash: hashStr, Algorithm: opts.Algorithm})
			return json.NewEncoder(w).Encode(HashesResult{Hashes: results, Count: len(results)})
		}

		_, _ = fmt.Fprintf(w, "%s  -\n", hashStr)

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
						if opts.JSON {
							result, hashErr := hashFileResult(p, opts)
							if hashErr == nil {
								results = append(results, result)
							}

							return nil
						}

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

		if opts.JSON {
			result, hashErr := hashFileResult(path, opts)
			if hashErr == nil {
				results = append(results, result)
			} else {
				_, _ = fmt.Fprintf(os.Stderr, "hash: %s: %v\n", path, hashErr)
			}
		} else {
			if err := hashFile(w, path, opts); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "hash: %s: %v\n", path, err)
			}
		}
	}

	if opts.JSON {
		return json.NewEncoder(w).Encode(HashesResult{Hashes: results, Count: len(results)})
	}

	return nil
}

func hashFileResult(path string, opts HashOptions) (HashResult, error) {
	algo := hashutil.Algorithm(opts.Algorithm)

	f, err := os.Open(path)
	if err != nil {
		return HashResult{}, err
	}

	defer func() { _ = f.Close() }()

	info, err := f.Stat()
	if err != nil {
		return HashResult{}, err
	}

	hashStr, err := hashutil.HashReader(f, algo)
	if err != nil {
		return HashResult{}, err
	}

	return HashResult{
		Path:      path,
		Hash:      hashStr,
		Algorithm: opts.Algorithm,
		Size:      info.Size(),
	}, nil
}

func hashFile(w io.Writer, path string, opts HashOptions) error {
	algo := hashutil.Algorithm(opts.Algorithm)

	hashStr, err := hashutil.HashFile(path, algo)
	if err != nil {
		return err
	}

	mode := " "
	if opts.Binary {
		mode = "*"
	}

	_, _ = fmt.Fprintf(w, "%s %s%s\n", hashStr, mode, path)

	return nil
}

func verifyChecksums(w io.Writer, args []string, opts HashOptions) error {
	if len(args) == 0 {
		return fmt.Errorf("hash: no checksum file specified")
	}

	algo := hashutil.Algorithm(opts.Algorithm)
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

			actualHash, err := hashutil.HashFile(filename, algo)
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

// RunCRC32Sum computes CRC32 checksums (IEEE polynomial)
func RunCRC32Sum(w io.Writer, args []string, opts HashOptions) error {
	opts.Algorithm = "crc32"
	return RunHash(w, args, opts)
}

// RunCRC64Sum computes CRC64 checksums (ECMA polynomial)
func RunCRC64Sum(w io.Writer, args []string, opts HashOptions) error {
	opts.Algorithm = "crc64"
	return RunHash(w, args, opts)
}
