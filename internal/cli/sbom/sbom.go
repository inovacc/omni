// Package sbom implements the I/O glue for the `omni sbom` command. It bridges
// Cobra to pkg/sbom: it picks a collector (Go module directory or built Go
// binary), resolves the source SBOM into a deterministic Document, and emits
// SPDX 2.3 or CycloneDX 1.5 JSON. Output is byte-identical for identical input.
//
// Optional detached signing reuses pkg/sign over the emitted bytes; signing is
// strictly file-based (requires both --out and --key) and the secret-key
// passphrase is read from OMNI_SIGN_PASSPHRASE — NEVER a flag.
package sbom

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/pkg/sbom/collect"
	"github.com/inovacc/omni/pkg/sbom/format"
	"github.com/inovacc/omni/pkg/sbom/model"
	"github.com/inovacc/omni/pkg/sign"
)

// passphraseEnv is the environment variable that may carry the secret-key
// passphrase used for --sign. The passphrase is never accepted as a flag value,
// matching the convention established by the `omni sign` command.
const passphraseEnv = "OMNI_SIGN_PASSPHRASE"

// outputPerm is the on-disk permission for the emitted SBOM and signature files.
const outputPerm os.FileMode = 0o644

// SBOMOptions configures RunSBOM.
type SBOMOptions struct {
	// Format selects the output document format: "spdx" (default) or
	// "cyclonedx"/"cdx".
	Format string
	// From selects the source kind: "auto" (default), "module", or "binary".
	From string
	// SourceDate is a fixed RFC-3339 creation timestamp. Empty means the epoch
	// default, which keeps output deterministic with no flag.
	SourceDate string
	// OmniVersion labels the generating tool in the emitted document.
	OmniVersion string
	// OutPath, when set, writes the document to that file instead of the writer.
	OutPath string
	// Sign produces a detached minisign signature next to OutPath. It requires
	// both OutPath and KeyPath.
	Sign bool
	// KeyPath is the secret-key file used when Sign is set.
	KeyPath string
	// Validate runs the emitted document through the upstream schema validator
	// (only available when built with -tags omni_sbomvalidate).
	Validate bool
}

// RunSBOM generates an SBOM for the single PATH in args and writes it to w (or
// to opts.OutPath when set). A directory is treated as a Go module directory; a
// regular file is treated as a built Go binary. An explicit opts.From that
// conflicts with the path kind is rejected with ErrInvalidInput.
func RunSBOM(w io.Writer, args []string, opts SBOMOptions) error {
	kind, err := parseFormat(opts.Format)
	if err != nil {
		return err
	}

	if len(args) != 1 {
		return cmderr.Wrap(cmderr.ErrInvalidInput, "sbom: exactly one PATH argument is required")
	}
	path := args[0]

	info, err := os.Stat(path)
	if err != nil {
		switch {
		case errors.Is(err, os.ErrNotExist):
			return cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("sbom: %s", path))
		case errors.Is(err, os.ErrPermission):
			return cmderr.Wrap(cmderr.ErrPermission, fmt.Sprintf("sbom: %s", path))
		default:
			return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("sbom: %s: %v", path, err))
		}
	}

	sb, err := collectSBOM(path, info.IsDir(), opts.From)
	if err != nil {
		return err
	}

	doc := format.From(sb, format.Options{OmniVersion: opts.OmniVersion, SourceDate: opts.SourceDate})

	// Encode to an in-memory buffer when we need the bytes for validation,
	// signing, or file output; otherwise stream straight to w.
	if opts.Validate || opts.Sign || opts.OutPath != "" {
		var buf bytes.Buffer
		if err := doc.Encode(&buf, kind); err != nil {
			return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("sbom: encode: %v", err))
		}
		if opts.Validate {
			if err := validateDocument(buf.Bytes(), opts.Format); err != nil {
				return err
			}
		}
		return emit(w, buf.Bytes(), opts)
	}

	if err := doc.Encode(w, kind); err != nil {
		return cmderr.Wrap(cmderr.ErrIO, "sbom: write sbom")
	}
	return nil
}

// parseFormat maps the --format string to a format.Kind. An empty value
// defaults to SPDX; any unrecognized value is an ErrInvalidInput.
func parseFormat(f string) (format.Kind, error) {
	switch f {
	case "", "spdx":
		return format.SPDX, nil
	case "cyclonedx", "cdx":
		return format.CycloneDX, nil
	default:
		return format.SPDX, cmderr.Wrap(cmderr.ErrInvalidInput, "sbom: unknown --format (want spdx|cyclonedx)")
	}
}

// collectSBOM picks a collector from the resolved path kind and an explicit
// --from selector, rejecting conflicts and translating collector errors into
// cmderr sentinels.
func collectSBOM(path string, isDir bool, from string) (*model.SBOM, error) {
	switch from {
	case "", "auto":
		// fall through to kind-based auto-detection
	case "module":
		if !isDir {
			return nil, cmderr.Wrap(cmderr.ErrInvalidInput, "sbom: --from module requires a directory PATH")
		}
	case "binary":
		if isDir {
			return nil, cmderr.Wrap(cmderr.ErrInvalidInput, "sbom: --from binary requires a file PATH")
		}
	default:
		return nil, cmderr.Wrap(cmderr.ErrInvalidInput, "sbom: unknown --from (want auto|module|binary)")
	}

	var (
		sb  *model.SBOM
		err error
	)
	if isDir {
		sb, err = collect.ModuleDir(path)
	} else {
		sb, err = collect.BinaryFile(path)
	}
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("sbom: %s", path))
		}
		return nil, cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("sbom: %v", err))
	}
	return sb, nil
}

// emit writes the encoded document either to w (no --out) or to opts.OutPath,
// then signs the file when opts.Sign is set.
func emit(w io.Writer, data []byte, opts SBOMOptions) error {
	if opts.OutPath == "" {
		if opts.Sign {
			return cmderr.Wrap(cmderr.ErrInvalidInput, "sbom: --sign requires --out")
		}
		if _, err := w.Write(data); err != nil {
			return cmderr.Wrap(cmderr.ErrIO, "sbom: write sbom")
		}
		return nil
	}

	if err := os.WriteFile(opts.OutPath, data, outputPerm); err != nil {
		return classifyFileErr(err, opts.OutPath)
	}
	if !opts.Sign {
		return nil
	}
	return signFile(w, opts)
}

// signFile produces a detached minisign signature for opts.OutPath using the
// secret key at opts.KeyPath. The passphrase is read from OMNI_SIGN_PASSPHRASE.
func signFile(w io.Writer, opts SBOMOptions) error {
	if opts.KeyPath == "" {
		return cmderr.Wrap(cmderr.ErrInvalidInput, "sbom: --sign requires --key")
	}

	keyText, err := os.ReadFile(opts.KeyPath)
	if err != nil {
		return classifyFileErr(err, opts.KeyPath)
	}

	passphrase, ok := os.LookupEnv(passphraseEnv)
	if !ok {
		return cmderr.Wrap(cmderr.ErrInvalidInput,
			fmt.Sprintf("sbom: --sign requires the secret-key passphrase via %s", passphraseEnv))
	}

	sk, err := sign.ParseSecretKey(keyText, passphrase)
	if err != nil {
		if errors.Is(err, sign.ErrVerification) {
			return cmderr.Wrap(cmderr.ErrInvalidInput, "sbom: wrong passphrase or corrupt secret key")
		}
		return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("sbom: parse secret key: %v", err))
	}

	data, err := os.ReadFile(opts.OutPath)
	if err != nil {
		return classifyFileErr(err, opts.OutPath)
	}

	sigText, err := sign.Sign(data, sk)
	if err != nil {
		return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("sbom: %v", err))
	}

	sigPath := opts.OutPath + ".minisig"
	if err := os.WriteFile(sigPath, sigText, outputPerm); err != nil {
		return classifyFileErr(err, sigPath)
	}
	_, _ = fmt.Fprintf(w, "SBOM written to %s\n", opts.OutPath)
	_, _ = fmt.Fprintf(w, "Signature written to %s\n", sigPath)
	return nil
}

// classifyFileErr maps an os file error to a cmderr sentinel.
func classifyFileErr(err error, path string) error {
	switch {
	case errors.Is(err, os.ErrNotExist):
		return cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("sbom: %s", path))
	case errors.Is(err, os.ErrPermission):
		return cmderr.Wrap(cmderr.ErrPermission, fmt.Sprintf("sbom: %s", path))
	default:
		return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("sbom: %s: %v", path, err))
	}
}
