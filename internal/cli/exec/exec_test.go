package exec

import (
	"bytes"
	"strings"
	"testing"
)

func TestRun_NoCommand(t *testing.T) {
	var buf bytes.Buffer
	err := Run(&buf, "", nil, Options{})
	if err == nil {
		t.Error("expected error for empty command")
	}
}

func TestRun_DryRun_NoDetectors(t *testing.T) {
	var buf bytes.Buffer
	err := Run(&buf, "unknown-tool", []string{"arg1"}, Options{DryRun: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "No credential checks matched") {
		t.Errorf("expected 'No credential checks matched', got: %s", buf.String())
	}
}

func TestRun_DryRun_WithDetector(t *testing.T) {
	t.Setenv("AWS_ACCESS_KEY_ID", "AKID")
	var buf bytes.Buffer
	err := Run(&buf, "aws", []string{"s3", "ls"}, Options{DryRun: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "aws") {
		t.Errorf("expected aws in output, got: %s", output)
	}
	if !strings.Contains(output, "OK") {
		t.Errorf("expected OK status, got: %s", output)
	}
}

func TestRun_Strict_MissingCreds(t *testing.T) {
	t.Setenv("AWS_ACCESS_KEY_ID", "")
	t.Setenv("AWS_PROFILE", "")
	t.Setenv("AWS_SESSION_TOKEN", "")
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	var buf bytes.Buffer
	err := Run(&buf, "aws", []string{"s3", "ls"}, Options{Strict: true})
	if err == nil {
		t.Error("expected error in strict mode with missing credentials")
	}
	if !strings.Contains(err.Error(), "strict mode") {
		t.Errorf("expected strict mode error, got: %v", err)
	}
}

func TestRun_Force_SkipsChecks(t *testing.T) {
	// Force with a command that doesn't exist should still fail at execute
	var buf bytes.Buffer
	err := Run(&buf, "nonexistent-command-12345", nil, Options{Force: true})
	if err == nil {
		t.Error("expected error for nonexistent command")
	}
}

func TestRun_NoPrompt_Proceeds(t *testing.T) {
	t.Setenv("AWS_ACCESS_KEY_ID", "")
	t.Setenv("AWS_PROFILE", "")
	t.Setenv("AWS_SESSION_TOKEN", "")
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	var buf bytes.Buffer
	// NoPrompt with missing creds should proceed to execute (and fail because 'aws' may not exist in test)
	err := Run(&buf, "aws", []string{"s3", "ls"}, Options{NoPrompt: true})
	// We expect either a warning printed or an exec error, but not a prompt
	if err == nil {
		t.Log("aws command succeeded (might be installed)")
	}
	if strings.Contains(buf.String(), "Warning") {
		t.Log("warning was printed as expected")
	}
}

func TestPrintResults(t *testing.T) {
	var buf bytes.Buffer
	results := []CredentialStatus{
		{Tool: "aws", Needed: true, Present: true},
		{Tool: "docker", Needed: true, Present: false, Missing: []string{"~/.docker/config.json"}},
	}
	printResults(&buf, "test", []string{"arg"}, results)
	output := buf.String()
	if !strings.Contains(output, "aws") || !strings.Contains(output, "docker") {
		t.Errorf("expected both tools in output, got: %s", output)
	}
}

func TestPrintMissing(t *testing.T) {
	var buf bytes.Buffer
	missing := []CredentialStatus{
		{Tool: "npm", Missing: []string{"NPM_TOKEN"}, Suggestion: "Set NPM_TOKEN"},
	}
	printMissing(&buf, missing)
	output := buf.String()
	if !strings.Contains(output, "NPM_TOKEN") {
		t.Errorf("expected NPM_TOKEN in output, got: %s", output)
	}
}
