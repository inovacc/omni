package base

import (
	"bufio"
	"fmt"
	"io"
	"os"

	pkgenc "github.com/inovacc/omni/pkg/encoding"
)

// BaseOptions configures the base encode/decode command behavior
type BaseOptions struct {
	Decode        bool // -d: decode data
	Wrap          int  // -w: wrap encoded lines after N characters (0 = no wrap)
	IgnoreGarbage bool // -i: ignore non-alphabet characters when decoding
}

// RunBase64 encodes or decodes base64 data
func RunBase64(w io.Writer, args []string, opts BaseOptions) error {
	if opts.Wrap == 0 {
		opts.Wrap = 76 // default wrap for base64
	}

	var input io.Reader
	if len(args) == 0 || args[0] == "-" {
		input = os.Stdin
	} else {
		f, err := os.Open(args[0])
		if err != nil {
			return fmt.Errorf("base64: %w", err)
		}

		defer func() {
			_ = f.Close()
		}()

		input = f
	}

	data, err := io.ReadAll(input)
	if err != nil {
		return fmt.Errorf("base64: %w", err)
	}

	if opts.Decode {
		decoded, err := pkgenc.Base64Decode(string(data))
		if err != nil {
			return fmt.Errorf("base64: %w", err)
		}

		_, _ = w.Write(decoded)
	} else {
		encoded := pkgenc.Base64Encode(data)
		if opts.Wrap > 0 {
			encoded = pkgenc.WrapString(encoded, opts.Wrap)
		}

		_, _ = fmt.Fprintln(w, encoded)
	}

	return nil
}

// RunBase32 encodes or decodes base32 data
func RunBase32(w io.Writer, args []string, opts BaseOptions) error {
	if opts.Wrap == 0 {
		opts.Wrap = 76
	}

	var input io.Reader
	if len(args) == 0 || args[0] == "-" {
		input = os.Stdin
	} else {
		f, err := os.Open(args[0])
		if err != nil {
			return fmt.Errorf("base32: %w", err)
		}

		defer func() {
			_ = f.Close()
		}()

		input = f
	}

	data, err := io.ReadAll(input)
	if err != nil {
		return fmt.Errorf("base32: %w", err)
	}

	if opts.Decode {
		decoded, err := pkgenc.Base32Decode(string(data))
		if err != nil {
			return fmt.Errorf("base32: %w", err)
		}

		_, _ = w.Write(decoded)
	} else {
		encoded := pkgenc.Base32Encode(data)
		if opts.Wrap > 0 {
			encoded = pkgenc.WrapString(encoded, opts.Wrap)
		}

		_, _ = fmt.Fprintln(w, encoded)
	}

	return nil
}

// RunBase58 encodes or decodes base58 data (Bitcoin alphabet)
func RunBase58(w io.Writer, args []string, opts BaseOptions) error {
	var input io.Reader
	if len(args) == 0 || args[0] == "-" {
		input = os.Stdin
	} else {
		f, err := os.Open(args[0])
		if err != nil {
			return fmt.Errorf("base58: %w", err)
		}

		defer func() {
			_ = f.Close()
		}()

		input = f
	}

	// Read line by line for base58 (typically used for short strings)
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		line := scanner.Text()
		if opts.Decode {
			decoded := pkgenc.Base58Decode(line)
			_, _ = w.Write(decoded)
			_, _ = fmt.Fprintln(w)
		} else {
			encoded := pkgenc.Base58Encode([]byte(line))
			_, _ = fmt.Fprintln(w, encoded)
		}
	}

	return scanner.Err()
}
