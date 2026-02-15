package exec

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectorRegistry_Match(t *testing.T) {
	r := NewDetectorRegistry()

	tests := []struct {
		command string
		want    int // expected number of matched detectors
	}{
		{"npm", 1},
		{"npx", 1},
		{"pnpm", 1},
		{"docker", 1},
		{"kubectl", 1},
		{"helm", 1},
		{"aws", 1},
		{"gh", 1},
		{"go", 1},
		{"terraform", 1},
		{"tofu", 1},
		{"unknown-cmd", 0},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			got := r.Match(tt.command)
			if len(got) != tt.want {
				t.Errorf("Match(%q) returned %d detectors, want %d", tt.command, len(got), tt.want)
			}
		})
	}
}

func TestDetectNpm_EnvVar(t *testing.T) {
	t.Setenv("NPM_TOKEN", "test-token")
	status := detectNpm("npm", []string{"install"})
	if !status.Present {
		t.Error("expected credentials present when NPM_TOKEN is set")
	}
}

func TestDetectNpm_NodeAuthToken(t *testing.T) {
	t.Setenv("NPM_TOKEN", "")
	t.Setenv("NODE_AUTH_TOKEN", "test-token")
	status := detectNpm("npm", []string{"install"})
	if !status.Present {
		t.Error("expected credentials present when NODE_AUTH_TOKEN is set")
	}
}

func TestDetectNpm_Missing(t *testing.T) {
	t.Setenv("NPM_TOKEN", "")
	t.Setenv("NODE_AUTH_TOKEN", "")
	// Point HOME to temp dir with no .npmrc
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	status := detectNpm("npm", []string{"install"})
	if status.Present {
		t.Error("expected credentials missing when no NPM_TOKEN and no .npmrc")
	}
	if len(status.Missing) == 0 {
		t.Error("expected Missing to be populated")
	}
}

func TestDetectNpm_Npmrc(t *testing.T) {
	t.Setenv("NPM_TOKEN", "")
	t.Setenv("NODE_AUTH_TOKEN", "")
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	npmrc := filepath.Join(tmp, ".npmrc")
	if err := os.WriteFile(npmrc, []byte("//registry.npmjs.org/:_authToken=secret\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	status := detectNpm("npm", []string{"install"})
	if !status.Present {
		t.Error("expected credentials present when .npmrc has _authToken")
	}
}

func TestDetectAWS_EnvVar(t *testing.T) {
	t.Setenv("AWS_ACCESS_KEY_ID", "AKID")
	t.Setenv("AWS_PROFILE", "")
	t.Setenv("AWS_SESSION_TOKEN", "")
	status := detectAWS("aws", nil)
	if !status.Present {
		t.Error("expected credentials present when AWS_ACCESS_KEY_ID is set")
	}
}

func TestDetectAWS_Profile(t *testing.T) {
	t.Setenv("AWS_ACCESS_KEY_ID", "")
	t.Setenv("AWS_PROFILE", "myprofile")
	t.Setenv("AWS_SESSION_TOKEN", "")
	status := detectAWS("aws", nil)
	if !status.Present {
		t.Error("expected credentials present when AWS_PROFILE is set")
	}
}

func TestDetectAWS_Missing(t *testing.T) {
	t.Setenv("AWS_ACCESS_KEY_ID", "")
	t.Setenv("AWS_PROFILE", "")
	t.Setenv("AWS_SESSION_TOKEN", "")
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	status := detectAWS("aws", nil)
	if status.Present {
		t.Error("expected credentials missing")
	}
}

func TestDetectDocker_ConfigAuth(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("DOCKER_CONFIG", tmp)
	configFile := filepath.Join(tmp, "config.json")
	if err := os.WriteFile(configFile, []byte(`{"auths":{"https://index.docker.io/v1/":{"auth":"dGVzdDp0ZXN0"}}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	status := detectDocker("docker", nil)
	if !status.Present {
		t.Error("expected credentials present when docker config has auths")
	}
}

func TestDetectDocker_Missing(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("DOCKER_CONFIG", tmp)
	status := detectDocker("docker", nil)
	if status.Present {
		t.Error("expected credentials missing when no docker config")
	}
}

func TestDetectKubectl_Present(t *testing.T) {
	tmp := t.TempDir()
	kubeDir := filepath.Join(tmp, ".kube")
	if err := os.MkdirAll(kubeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(kubeDir, "config"), []byte("apiVersion: v1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("KUBECONFIG", "")
	t.Setenv("HOME", tmp)
	status := detectKubectl("kubectl", nil)
	if !status.Present {
		t.Error("expected credentials present when kubeconfig exists")
	}
}

func TestDetectGo_NoPrivate(t *testing.T) {
	t.Setenv("GOPRIVATE", "")
	status := detectGo("go", nil)
	if status.Needed {
		t.Error("expected Needed=false when GOPRIVATE is empty")
	}
}

func TestDetectGo_PrivateWithNetrc(t *testing.T) {
	t.Setenv("GOPRIVATE", "github.com/private/*")
	t.Setenv("GONOSUMCHECK", "")
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	if err := os.WriteFile(filepath.Join(tmp, ".netrc"), []byte("machine github.com\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	status := detectGo("go", nil)
	if !status.Present {
		t.Error("expected credentials present when .netrc exists")
	}
}

func TestDetectGH_Token(t *testing.T) {
	t.Setenv("GH_TOKEN", "ghp_test")
	t.Setenv("GITHUB_TOKEN", "")
	status := detectGH("gh", nil)
	if !status.Present {
		t.Error("expected credentials present when GH_TOKEN is set")
	}
}

func TestDetectTerraformCreds_AWS(t *testing.T) {
	t.Setenv("AWS_ACCESS_KEY_ID", "AKID")
	status := detectTerraformCreds("terraform", nil)
	if !status.Present {
		t.Error("expected credentials present when AWS_ACCESS_KEY_ID is set")
	}
}
