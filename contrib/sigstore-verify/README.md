# omni-sigstore-verify

Sigstore bundle verification, delivered as a **separate, self-contained Go
module** rather than a build tag in the main `omni` binary. `sigstore-go` drags
in roughly 50 transitive modules (Rekor, go-openapi, certificate-transparency,
timestamp-authority, go-tuf, in-toto, ...) and forces `golang.org/x/*` version
bumps. A build tag isolates compilation but **not** the `go.mod` module graph —
Go's minimal version selection still pulls those deps into the default build —
so the capability is quarantined in this module to keep the main omni `go.mod`
lean and pure-Go. The default `omni verify --bundle` returns an "unsupported"
error pointing here.

## Build / install

```sh
cd contrib/sigstore-verify && go build
# or
go install github.com/inovacc/omni/contrib/sigstore-verify@latest
```

## Usage

```sh
omni-sigstore-verify \
  --bundle artifact.sigstore.json \
  --trusted-root trusted_root.json \
  --artifact artifact.tar.gz \
  [--certificate-identity user@example.com] \
  [--certificate-oidc-issuer https://accounts.google.com]
```

Instead of `--artifact` you may pass `--artifact-digest sha256:<hex>`.
Verification is fail-closed: it succeeds only when the bundle's signature,
transparency-log inclusion, and observer timestamps all verify and the policy
(artifact + optional identity) matches. On success the `VerificationResult` is
printed as JSON; any failure exits non-zero with a clear message.
