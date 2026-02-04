package crypto

// GetMachineID returns a unique identifier for the current machine.
// This is used to derive encryption keys that are tied to the machine.
// The implementation is platform-specific.
func GetMachineID() (string, error) {
	return getMachineID()
}
