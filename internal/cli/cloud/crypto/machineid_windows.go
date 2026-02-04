//go:build windows

package crypto

import (
	"fmt"

	"golang.org/x/sys/windows/registry"
)

// getMachineID retrieves the machine GUID from Windows registry.
func getMachineID() (string, error) {
	key, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Cryptography`,
		registry.QUERY_VALUE,
	)
	if err != nil {
		return "", fmt.Errorf("opening registry key: %w", err)
	}

	defer func() { _ = key.Close() }()

	machineGUID, _, err := key.GetStringValue("MachineGuid")
	if err != nil {
		return "", fmt.Errorf("reading MachineGuid: %w", err)
	}

	if machineGUID == "" {
		return "", fmt.Errorf("MachineGuid is empty")
	}

	return machineGUID, nil
}
