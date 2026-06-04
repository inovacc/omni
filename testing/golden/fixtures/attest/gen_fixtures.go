//go:build ignore

// Command gen_fixtures generates the committed golden-master fixtures for the
// `attest` category: a fixed artifact, a deterministically-signed DSSE/SLSA
// provenance envelope (LOCAL builder.id, no --from-env, no timestamps -> byte
// deterministic), a tampered copy whose payload no longer matches the signature,
// and copies of the reused `sign` keypair (test.key/test.pub/wrong.pub).
//
// Run by hand (never in CI) to (re)materialize the fixtures:
//
//	go run testing/golden/fixtures/attest/gen_fixtures.go
//
// It reuses the sign category's low-cost key (passphrase golden-fixture-passphrase);
// the LOCAL builder.id path emits no wall-clock fields, so the envelope is stable.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	cliattest "github.com/inovacc/omni/internal/cli/attest"
)

const passphrase = "golden-fixture-passphrase"
const artifactBody = "omni attest fixture v1\n"

func main() {
	_, thisFile, _, _ := runtime.Caller(0)
	dir := filepath.Dir(thisFile)
	signDir := filepath.Join(dir, "..", "sign")

	// Reuse the sign keypair: copy test.key/test.pub/wrong.pub into the attest
	// fixture dir so the {fixtures} placeholder (-> fixtures/attest/) resolves them.
	for _, name := range []string{"test.key", "test.pub", "wrong.pub"} {
		b, err := os.ReadFile(filepath.Join(signDir, name))
		if err != nil {
			fatal("read sign/%s: %v", name, err)
		}
		write(filepath.Join(dir, name), b, 0o644)
	}

	artifact := filepath.Join(dir, "artifact.bin")
	write(artifact, []byte(artifactBody), 0o644)

	// Generate the signed envelope via the real CLI glue (local builder.id).
	_ = os.Setenv("OMNI_SIGN_PASSPHRASE", passphrase)
	out := filepath.Join(dir, "app.intoto.jsonl")
	if err := cliattest.RunAttest(os.Stdout, cliattest.GenOptions{
		KeyPath:       filepath.Join(dir, "test.key"),
		ArtifactPath:  artifact,
		PredicateType: "slsa-provenance",
		OutPath:       out,
	}); err != nil {
		fatal("RunAttest: %v", err)
	}

	// Tampered copy: flip one character of the base64 payload so the (unchanged)
	// signature no longer verifies -> fail-closed. Keep it valid JSON.
	envBytes, err := os.ReadFile(out)
	if err != nil {
		fatal("read envelope: %v", err)
	}
	var env map[string]any
	if err := json.Unmarshal(envBytes, &env); err != nil {
		fatal("unmarshal envelope: %v", err)
	}
	payload, _ := env["payload"].(string)
	if len(payload) < 8 {
		fatal("payload too short to tamper")
	}
	b := []byte(payload)
	i := len(b) / 2
	if b[i] == 'A' {
		b[i] = 'B'
	} else {
		b[i] = 'A'
	}
	env["payload"] = string(b)
	tampered, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		fatal("marshal tampered: %v", err)
	}
	tampered = append(tampered, '\n')
	write(filepath.Join(dir, "tampered.jsonl"), tampered, 0o644)

	// Determinism self-check: regenerate to a buffer and compare bytes.
	var buf bytes.Buffer
	if err := cliattest.RunAttest(&buf, cliattest.GenOptions{
		KeyPath: filepath.Join(dir, "test.key"), ArtifactPath: artifact, PredicateType: "slsa-provenance",
	}); err != nil {
		fatal("RunAttest (determinism check): %v", err)
	}
	if !bytes.Equal(bytes.TrimRight(envBytes, "\n"), bytes.TrimRight(buf.Bytes(), "\n")) {
		fatal("NON-DETERMINISTIC: local envelope differs across runs")
	}

	fmt.Println("attest fixtures written to", dir)
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
