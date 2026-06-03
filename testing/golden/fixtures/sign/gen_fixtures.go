//go:build ignore

// Command gen_fixtures generates the committed golden-master fixtures for the
// `sign` category: a low-scrypt-cost passphrase-protected key pair, a signed
// data file, and its detached *.minisig signature. It is run by hand (never in
// CI) to (re)materialize the fixtures checked into this directory:
//
//	go run testing/golden/fixtures/sign/gen_fixtures.go
//
// CRITICAL: this uses WithScryptParams(1<<15, 8, 1) — the default SENSITIVE cost
// (~1 GiB RAM) must NEVER run inside the test path. Tests verify against the
// committed output of this tool; they do not regenerate it.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/inovacc/omni/pkg/sign"
)

const (
	passphrase   = "golden-fixture-passphrase"
	dataContent  = "omni golden-master sign fixture payload\n"
	untrustedKey = "minisign public key fixture (omni golden)"
	untrustedSig = "signature from omni golden fixture"
	trustedSig   = "omni golden fixture: deterministic signed payload"
)

func main() {
	_, thisFile, _, _ := runtime.Caller(0)
	dir := filepath.Dir(thisFile)

	// Low-cost scrypt — NEVER the default SENSITIVE cost in automation.
	lowCost := sign.WithScryptParams(1<<15, 8, 1)

	kp, err := sign.GenerateKeyPair(passphrase, lowCost)
	if err != nil {
		fatal("generate key pair: %v", err)
	}

	pubText := kp.PublicKey.MarshalText()
	skText, err := kp.SecretKey.MarshalText(passphrase, lowCost,
		sign.WithUntrustedComment(untrustedKey))
	if err != nil {
		fatal("marshal secret key: %v", err)
	}

	data := []byte(dataContent)
	sig, err := sign.Sign(data, kp.SecretKey,
		sign.WithTrustedComment(trustedSig),
		sign.WithUntrustedComment(untrustedSig))
	if err != nil {
		fatal("sign data: %v", err)
	}

	write(filepath.Join(dir, "test.pub"), pubText, 0o644)
	write(filepath.Join(dir, "test.key"), skText, 0o600)
	write(filepath.Join(dir, "data.txt"), data, 0o644)
	write(filepath.Join(dir, "data.txt.minisig"), sig, 0o644)

	// A second public key (different key id) for the wrong-key negative case is
	// generated so verify fails closed against it. Committed as wrong.pub.
	kp2, err := sign.GenerateKeyPair(passphrase, lowCost)
	if err != nil {
		fatal("generate wrong key pair: %v", err)
	}
	write(filepath.Join(dir, "wrong.pub"), kp2.PublicKey.MarshalText(), 0o644)

	// tampered.txt is a payload that does NOT match data.txt.minisig, so verify
	// must fail closed (ErrConflict / exit 1).
	write(filepath.Join(dir, "tampered.txt"), []byte("tampered payload — not what was signed\n"), 0o644)

	// bad.pub is a syntactically malformed public key (ErrInvalidInput / exit 2).
	write(filepath.Join(dir, "bad.pub"), []byte("untrusted comment: not a real key\nthis-is-not-base64!!!\n"), 0o644)

	fmt.Println("fixtures written to", dir)
}

func write(path string, b []byte, mode os.FileMode) {
	if err := os.WriteFile(path, b, mode); err != nil {
		fatal("write %s: %v", path, err)
	}
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
