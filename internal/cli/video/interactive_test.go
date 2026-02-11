package video

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunInteractivePromptURLThenQuit(t *testing.T) {
	in := strings.NewReader("https://www.youtube.com/watch?v=dQw4w9WgXcQ\n8\n")
	var out bytes.Buffer
	var prompt bytes.Buffer

	if err := RunInteractive(&out, &prompt, in, nil, Options{}); err != nil {
		t.Fatalf("RunInteractive() error = %v", err)
	}

	promptText := prompt.String()
	if !strings.Contains(promptText, "Video URL: ") {
		t.Fatalf("expected URL prompt, got %q", promptText)
	}

	if !strings.Contains(promptText, "Select an option: ") {
		t.Fatalf("expected menu selection prompt, got %q", promptText)
	}
}

func TestRunInteractiveQuitWithArgs(t *testing.T) {
	in := strings.NewReader("8\n")
	var out bytes.Buffer
	var prompt bytes.Buffer

	err := RunInteractive(&out, &prompt, in, []string{"https://example.com/video.mp4"}, Options{})
	if err != nil {
		t.Fatalf("RunInteractive() error = %v", err)
	}

	if !strings.Contains(prompt.String(), "Interactive video menu for: https://example.com/video.mp4") {
		t.Fatalf("expected menu header with URL, got %q", prompt.String())
	}
}
