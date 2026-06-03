//go:build darwin

package crypto

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// SANCTIONED EXEC EXCEPTION (machine identity, darwin only).
//
// omni's design forbids spawning external processes. This file is the single
// deliberate exception, and it exists for a hard platform constraint, not for
// convenience:
//
//   - macOS exposes the stable hardware UUID (IOPlatformUUID) only through the
//     IOKit framework, which requires cgo, or through the `ioreg` CLI. There is
//     no pure-Go / no-cgo API that returns this same value. omni must build with
//     CGO_ENABLED=0, so IOKit is off the table, leaving `ioreg` as the only way
//     to read the identifier.
//
//   - The machine ID is fed directly into the master-key KDF (see master.go,
//     deriveMasterEncKey -> pbkdf2.Key). Changing the *value* returned here would
//     change every derived key and make any existing machine-bound master.key on
//     a macOS install permanently undecryptable. A different pure-Go source
//     (persisted random UUID, hashed primary MAC, etc.) would change the value
//     and is therefore NOT read-through-compatible with installs that already
//     hold ciphertext derived from IOPlatformUUID. Per the project's
//     backward-compat rule, that silent value change is forbidden.
//
// Both facts together mean the no-exec migration cannot be done here without
// breaking existing encrypted data. The exec call is kept as the sanctioned
// machine-identity exec site and hardened against PATH hijacking and injection:
// it resolves the absolute system path to ioreg, never invokes a shell, and
// passes a fixed argument vector (no user-controlled input).
//
// If macOS ever ships a pure-Go-reachable source for the same IOPlatformUUID,
// or omni adopts a migration that re-encrypts master.key on identifier change,
// this exception should be removed.

// ioregPath is the absolute path to the macOS ioreg utility. Using the fixed
// system path (rather than relying on $PATH) prevents a hijacked PATH from
// substituting a malicious binary.
const ioregPath = "/usr/sbin/ioreg"

// getMachineID retrieves the hardware UUID (IOPlatformUUID) using ioreg on
// macOS. See the SANCTIONED EXEC EXCEPTION note above for why exec is used here.
func getMachineID() (string, error) {
	// Fixed binary path and fixed argument vector: no shell, no user input, so
	// this call is not injection-prone.
	cmd := exec.Command(ioregPath, "-rd1", "-c", "IOPlatformExpertDevice")

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("running ioreg: %w", err)
	}

	// Parse output to find IOPlatformUUID
	output := out.String()
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "IOPlatformUUID") {
			// Line format: "IOPlatformUUID" = "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"
			parts := strings.Split(line, "=")
			if len(parts) >= 2 {
				uuid := strings.TrimSpace(parts[1])
				uuid = strings.Trim(uuid, `"`)
				if uuid != "" {
					return uuid, nil
				}
			}
		}
	}

	return "", fmt.Errorf("IOPlatformUUID not found in ioreg output")
}
