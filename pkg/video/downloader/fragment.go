package downloader

import (
	"encoding/json"
	"fmt"
	"os"
)

// FragmentState tracks the progress of a fragmented download for resume.
type FragmentState struct {
	TotalFragments int    `json:"total_fragments"`
	LastFragment   int    `json:"last_fragment"`
	Filename       string `json:"filename"`
}

// LoadFragmentState loads download state from a state file.
func LoadFragmentState(stateFile string) (*FragmentState, error) {
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return nil, err
	}

	var state FragmentState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("fragment: invalid state file: %w", err)
	}

	return &state, nil
}

// SaveFragmentState saves download state to a state file.
func SaveFragmentState(stateFile string, state *FragmentState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("fragment: marshal state: %w", err)
	}

	return os.WriteFile(stateFile, data, 0o600)
}

// RemoveFragmentState deletes the state file.
func RemoveFragmentState(stateFile string) {
	_ = os.Remove(stateFile)
}

// AppendToFile appends data to a file, creating it if needed.
func AppendToFile(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("fragment: open: %w", err)
	}

	defer func() { _ = f.Close() }()

	_, err = f.Write(data)
	if err != nil {
		return fmt.Errorf("fragment: write: %w", err)
	}

	return nil
}
