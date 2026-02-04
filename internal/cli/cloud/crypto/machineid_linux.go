//go:build linux

package crypto

import (
	"fmt"
	"os"
	"strings"
)

// getMachineID reads the machine ID from /etc/machine-id.
func getMachineID() (string, error) {
	// Primary location
	data, err := os.ReadFile("/etc/machine-id")
	if err == nil {
		id := strings.TrimSpace(string(data))
		if id != "" {
			return id, nil
		}
	}

	// Fallback location (used by some systems)
	data, err = os.ReadFile("/var/lib/dbus/machine-id")
	if err == nil {
		id := strings.TrimSpace(string(data))
		if id != "" {
			return id, nil
		}
	}

	return "", fmt.Errorf("could not read machine-id from /etc/machine-id or /var/lib/dbus/machine-id")
}
