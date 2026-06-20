package exec

import (
	"os"
	"path/filepath"
	"testing"
)

// isolateHome points HOME/USERPROFILE at an empty temp dir and clears every
// credential env var the detectors consult, so the "missing" branch is
// deterministic regardless of the developer's real environment.
func isolateHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	for _, e := range []string{
		"PIP_INDEX_URL", "DOCKER_CONFIG", "KUBECONFIG",
		"AWS_ACCESS_KEY_ID", "AWS_PROFILE", "AWS_SESSION_TOKEN",
		"GOOGLE_APPLICATION_CREDENTIALS", "GOOGLE_CREDENTIALS", "GCLOUD_KEYFILE_JSON",
		"GH_TOKEN", "GITHUB_TOKEN", "GOPRIVATE", "GONOSUMCHECK",
		"NPM_TOKEN", "NODE_AUTH_TOKEN",
		"ARM_CLIENT_ID", "ARM_SUBSCRIPTION_ID",
		"AWS_SHARED_CREDENTIALS_FILE",
	} {
		t.Setenv(e, "")
	}
	return home
}

func TestDetectors_Missing(t *testing.T) {
	tests := []struct {
		name string
		fn   func(string, []string) CredentialStatus
		tool string
	}{
		{"pip", detectPip, "pip"},
		{"gcloud", detectGCloud, "gcloud"},
		{"gh", detectGH, "gh"},
		{"docker", detectDocker, "docker"},
		{"kubectl", detectKubectl, "kubectl"},
		{"aws", detectAWS, "aws"},
		{"terraform", detectTerraformCreds, "terraform"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isolateHome(t)
			st := tt.fn("cmd", nil)
			if st.Tool != tt.tool {
				t.Errorf("Tool = %q want %q", st.Tool, tt.tool)
			}
			if st.Present {
				t.Errorf("%s: expected Present=false in isolated env", tt.tool)
			}
			if len(st.Missing) == 0 {
				t.Errorf("%s: expected Missing list to be populated", tt.tool)
			}
		})
	}
}

func TestDetectors_PresentViaEnv(t *testing.T) {
	t.Run("gcloud env", func(t *testing.T) {
		isolateHome(t)
		t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/tmp/key.json")
		if st := detectGCloud("", nil); !st.Present {
			t.Error("expected gcloud present via GOOGLE_APPLICATION_CREDENTIALS")
		}
	})
	t.Run("gh env", func(t *testing.T) {
		isolateHome(t)
		t.Setenv("GH_TOKEN", "x")
		if st := detectGH("", nil); !st.Present {
			t.Error("expected gh present via GH_TOKEN")
		}
	})
	t.Run("aws env", func(t *testing.T) {
		isolateHome(t)
		t.Setenv("AWS_ACCESS_KEY_ID", "AKIA")
		if st := detectAWS("", nil); !st.Present {
			t.Error("expected aws present via AWS_ACCESS_KEY_ID")
		}
	})
	t.Run("pip via config file", func(t *testing.T) {
		home := isolateHome(t)
		pipDir := filepath.Join(home, ".pip")
		if err := os.MkdirAll(pipDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(pipDir, "pip.conf"), []byte("[global]\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if st := detectPip("", nil); !st.Present {
			t.Error("expected pip present via ~/.pip/pip.conf")
		}
	})
	t.Run("go no private modules", func(t *testing.T) {
		isolateHome(t) // GOPRIVATE empty → Needed=false, Present=true
		st := detectGo("", nil)
		if st.Needed || !st.Present {
			t.Errorf("go without GOPRIVATE: Needed=%v Present=%v want false/true", st.Needed, st.Present)
		}
	})
	t.Run("go private with nosumcheck", func(t *testing.T) {
		isolateHome(t)
		t.Setenv("GOPRIVATE", "example.com/private")
		t.Setenv("GONOSUMCHECK", "1")
		if st := detectGo("", nil); !st.Present {
			t.Error("expected go present via GONOSUMCHECK")
		}
	})
}

func TestBuildDryRunResult(t *testing.T) {
	results := []CredentialStatus{
		{Tool: "aws", Present: true},
		{Tool: "gh", Present: false, Missing: []string{"GH_TOKEN"}, Suggestion: "gh auth login"},
	}
	r := buildDryRunResult("deploy", []string{"--prod"}, results)
	if r.Command != "deploy" || len(r.Args) != 1 {
		t.Fatalf("unexpected command/args: %+v", r)
	}
	if len(r.Checks) != 2 {
		t.Fatalf("checks = %d want 2", len(r.Checks))
	}
	if r.Checks[0].Status != "ok" {
		t.Errorf("present tool status = %q want ok", r.Checks[0].Status)
	}
	if r.Checks[1].Status != "missing" || len(r.Checks[1].Missing) != 1 {
		t.Errorf("missing tool not represented: %+v", r.Checks[1])
	}
}
