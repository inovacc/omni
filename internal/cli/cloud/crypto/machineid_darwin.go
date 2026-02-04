//go:build darwin

package crypto

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// getMachineID retrieves the hardware UUID using ioreg on macOS.
func getMachineID() (string, error) {
	cmd := exec.Command("ioreg", "-rd1", "-c", "IOPlatformExpertDevice")

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
