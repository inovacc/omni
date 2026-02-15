package hacks

import (
	"testing"
)

func TestRunGhCommandOutput_InvalidCommand(t *testing.T) {
	// This tests that runGhCommandOutput returns an error for invalid gh subcommands
	_, err := runGhCommandOutput("__nonexistent_subcommand__")
	if err == nil {
		t.Error("expected error for invalid gh subcommand")
	}
}
