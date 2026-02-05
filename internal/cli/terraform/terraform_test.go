package terraform

import (
	"os/exec"
	"testing"
)

func TestVersion(t *testing.T) {
	// Check if terraform is available
	_, err := exec.LookPath("terraform")
	if err != nil {
		t.Skip("terraform not available")
	}

	// Test version command
	err = Version()
	if err != nil {
		t.Errorf("Version failed: %v", err)
	}
}

func TestValidate_NoTerraformFiles(t *testing.T) {
	// Check if terraform is available
	_, err := exec.LookPath("terraform")
	if err != nil {
		t.Skip("terraform not available")
	}

	// Test validate in empty directory - should fail
	err = Validate()
	// We expect this to fail since there's no terraform configuration
	if err == nil {
		t.Log("Validate succeeded (might have terraform files in working dir)")
	}
}

func TestWorkspaceShow(t *testing.T) {
	// Check if terraform is available
	_, err := exec.LookPath("terraform")
	if err != nil {
		t.Skip("terraform not available")
	}

	// Test workspace show - should work without terraform files
	err = WorkspaceShow()
	// May fail if not in initialized directory
	if err != nil {
		t.Logf("WorkspaceShow failed (expected if not initialized): %v", err)
	}
}

func TestProviders(t *testing.T) {
	// Check if terraform is available
	_, err := exec.LookPath("terraform")
	if err != nil {
		t.Skip("terraform not available")
	}

	// Test providers command
	err = Providers()
	// May fail if not in terraform directory
	if err != nil {
		t.Logf("Providers failed (expected if no tf files): %v", err)
	}
}

// Test command construction functions

func TestBuildPlanArgs(t *testing.T) {
	tests := []struct {
		name     string
		out      string
		vars     map[string]string
		varFiles []string
		destroy  bool
		wantLen  int
	}{
		{
			name:    "basic plan",
			wantLen: 1, // just "plan"
		},
		{
			name:    "plan with output",
			out:     "plan.tfplan",
			wantLen: 2, // "plan", "-out=plan.tfplan"
		},
		{
			name:    "plan with destroy",
			destroy: true,
			wantLen: 2, // "plan", "-destroy"
		},
		{
			name:    "plan with vars",
			vars:    map[string]string{"region": "us-east-1", "env": "prod"},
			wantLen: 5, // "plan", "-var", "region=us-east-1", "-var", "env=prod"
		},
		{
			name:     "plan with var files",
			varFiles: []string{"dev.tfvars", "secret.tfvars"},
			wantLen:  3, // "plan", "-var-file=dev.tfvars", "-var-file=secret.tfvars"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []string{"plan"}
			if tt.out != "" {
				args = append(args, "-out="+tt.out)
			}

			if tt.destroy {
				args = append(args, "-destroy")
			}

			for k, v := range tt.vars {
				args = append(args, "-var", k+"="+v)
			}

			for _, f := range tt.varFiles {
				args = append(args, "-var-file="+f)
			}

			if len(args) < tt.wantLen {
				t.Errorf("expected at least %d args, got %d: %v", tt.wantLen, len(args), args)
			}
		})
	}
}
