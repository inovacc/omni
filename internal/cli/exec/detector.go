package exec

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// CredentialStatus reports whether a tool's credentials are present.
type CredentialStatus struct {
	Tool       string
	Needed     bool
	Present    bool
	Missing    []string
	Suggestion string
}

// Detector inspects a command and returns its credential status.
type Detector func(command string, args []string) CredentialStatus

// DetectorRegistry maps command prefixes to their detectors.
type DetectorRegistry struct {
	entries []registryEntry
}

type registryEntry struct {
	prefixes []string
	detector Detector
}

// NewDetectorRegistry returns a registry with all built-in detectors.
func NewDetectorRegistry() *DetectorRegistry {
	r := &DetectorRegistry{}
	r.Register([]string{"npm", "npx", "pnpm"}, detectNpm)
	r.Register([]string{"pip", "pipx", "uv"}, detectPip)
	r.Register([]string{"docker"}, detectDocker)
	r.Register([]string{"kubectl", "helm"}, detectKubectl)
	r.Register([]string{"terraform", "tofu"}, detectTerraformCreds)
	r.Register([]string{"aws"}, detectAWS)
	r.Register([]string{"gcloud"}, detectGCloud)
	r.Register([]string{"gh"}, detectGH)
	r.Register([]string{"go"}, detectGo)
	return r
}

// Register adds a detector for the given command prefixes.
func (r *DetectorRegistry) Register(prefixes []string, d Detector) {
	r.entries = append(r.entries, registryEntry{prefixes: prefixes, detector: d})
}

// Match returns all detectors whose prefix matches the command.
func (r *DetectorRegistry) Match(command string) []Detector {
	base := filepath.Base(command)
	var matched []Detector
	for _, e := range r.entries {
		for _, p := range e.prefixes {
			if base == p {
				matched = append(matched, e.detector)
				break
			}
		}
	}
	return matched
}

// --- Detectors ---

func detectNpm(_ string, _ []string) CredentialStatus {
	status := CredentialStatus{Tool: "npm", Needed: true}

	// Check env vars
	if os.Getenv("NPM_TOKEN") != "" || os.Getenv("NODE_AUTH_TOKEN") != "" {
		status.Present = true
		return status
	}

	// Check .npmrc files for auth
	home, _ := os.UserHomeDir()
	for _, path := range []string{
		filepath.Join(home, ".npmrc"),
		".npmrc",
	} {
		if hasNpmrcAuth(path) {
			status.Present = true
			return status
		}
	}

	status.Missing = []string{"NPM_TOKEN", "NODE_AUTH_TOKEN", "~/.npmrc auth entry"}
	status.Suggestion = "Set NPM_TOKEN or add //registry.npmjs.org/:_authToken=<token> to ~/.npmrc"
	return status
}

func hasNpmrcAuth(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.Contains(line, "_authToken") || strings.Contains(line, "_auth") || strings.Contains(line, "_password") {
			return true
		}
	}
	return false
}

func detectPip(_ string, _ []string) CredentialStatus {
	status := CredentialStatus{Tool: "pip", Needed: true}

	if url := os.Getenv("PIP_INDEX_URL"); url != "" && (strings.Contains(url, "@") || strings.Contains(url, "://") && strings.Contains(url, ":")) {
		status.Present = true
		return status
	}

	home, _ := os.UserHomeDir()
	for _, path := range []string{
		filepath.Join(home, ".pip", "pip.conf"),
		filepath.Join(home, ".config", "pip", "pip.conf"),
		filepath.Join(home, ".netrc"),
	} {
		if _, err := os.Stat(path); err == nil {
			status.Present = true
			return status
		}
	}

	status.Missing = []string{"PIP_INDEX_URL (with auth)", "~/.pip/pip.conf", "~/.netrc"}
	status.Suggestion = "Set PIP_INDEX_URL with credentials or configure ~/.pip/pip.conf"
	return status
}

func detectDocker(_ string, _ []string) CredentialStatus {
	status := CredentialStatus{Tool: "docker", Needed: true}

	configDir := os.Getenv("DOCKER_CONFIG")
	if configDir == "" {
		home, _ := os.UserHomeDir()
		configDir = filepath.Join(home, ".docker")
	}

	configFile := filepath.Join(configDir, "config.json")
	data, err := os.ReadFile(configFile)
	if err == nil {
		var cfg map[string]any
		if json.Unmarshal(data, &cfg) == nil {
			if auths, ok := cfg["auths"].(map[string]any); ok && len(auths) > 0 {
				status.Present = true
				return status
			}
			if _, ok := cfg["credsStore"]; ok {
				status.Present = true
				return status
			}
		}
	}

	status.Missing = []string{"~/.docker/config.json auth entries"}
	status.Suggestion = "Run 'docker login' to authenticate"
	return status
}

func detectKubectl(_ string, _ []string) CredentialStatus {
	status := CredentialStatus{Tool: "kubectl", Needed: true}

	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		home, _ := os.UserHomeDir()
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	if _, err := os.Stat(kubeconfig); err == nil {
		status.Present = true
		return status
	}

	status.Missing = []string{"KUBECONFIG", "~/.kube/config"}
	status.Suggestion = "Set KUBECONFIG or configure ~/.kube/config"
	return status
}

func detectTerraformCreds(_ string, _ []string) CredentialStatus {
	status := CredentialStatus{Tool: "terraform", Needed: true}

	providerEnvs := [][]string{
		{"AWS_ACCESS_KEY_ID", "AWS_PROFILE", "AWS_SHARED_CREDENTIALS_FILE"},
		{"GOOGLE_CREDENTIALS", "GOOGLE_APPLICATION_CREDENTIALS", "GCLOUD_KEYFILE_JSON"},
		{"ARM_CLIENT_ID", "ARM_SUBSCRIPTION_ID"},
	}

	for _, group := range providerEnvs {
		for _, env := range group {
			if os.Getenv(env) != "" {
				status.Present = true
				return status
			}
		}
	}

	status.Missing = []string{"AWS_ACCESS_KEY_ID/AWS_PROFILE", "GOOGLE_CREDENTIALS", "ARM_CLIENT_ID"}
	status.Suggestion = "Set provider credentials or run 'omni cloud profile use <provider> <name>'"
	return status
}

func detectAWS(_ string, _ []string) CredentialStatus {
	status := CredentialStatus{Tool: "aws", Needed: true}

	if os.Getenv("AWS_ACCESS_KEY_ID") != "" || os.Getenv("AWS_PROFILE") != "" || os.Getenv("AWS_SESSION_TOKEN") != "" {
		status.Present = true
		return status
	}

	home, _ := os.UserHomeDir()
	if _, err := os.Stat(filepath.Join(home, ".aws", "credentials")); err == nil {
		status.Present = true
		return status
	}

	status.Missing = []string{"AWS_ACCESS_KEY_ID", "AWS_PROFILE", "~/.aws/credentials"}
	status.Suggestion = "Run 'omni cloud profile use aws <name>' or set AWS_ACCESS_KEY_ID + AWS_SECRET_ACCESS_KEY"
	return status
}

func detectGCloud(_ string, _ []string) CredentialStatus {
	status := CredentialStatus{Tool: "gcloud", Needed: true}

	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") != "" {
		status.Present = true
		return status
	}

	home, _ := os.UserHomeDir()
	gcloudConfig := filepath.Join(home, ".config", "gcloud", "application_default_credentials.json")
	if _, err := os.Stat(gcloudConfig); err == nil {
		status.Present = true
		return status
	}

	status.Missing = []string{"GOOGLE_APPLICATION_CREDENTIALS", "gcloud auth application-default login"}
	status.Suggestion = "Run 'gcloud auth login' or set GOOGLE_APPLICATION_CREDENTIALS"
	return status
}

func detectGH(_ string, _ []string) CredentialStatus {
	status := CredentialStatus{Tool: "gh", Needed: true}

	if os.Getenv("GH_TOKEN") != "" || os.Getenv("GITHUB_TOKEN") != "" {
		status.Present = true
		return status
	}

	home, _ := os.UserHomeDir()
	ghConfig := filepath.Join(home, ".config", "gh", "hosts.yml")
	if _, err := os.Stat(ghConfig); err == nil {
		status.Present = true
		return status
	}

	status.Missing = []string{"GH_TOKEN", "GITHUB_TOKEN", "gh auth login"}
	status.Suggestion = "Run 'gh auth login' or set GH_TOKEN"
	return status
}

func detectGo(_ string, _ []string) CredentialStatus {
	status := CredentialStatus{Tool: "go", Needed: true}

	if os.Getenv("GOPRIVATE") == "" {
		// No private modules configured, credentials not needed
		status.Needed = false
		status.Present = true
		return status
	}

	if os.Getenv("GONOSUMCHECK") != "" {
		status.Present = true
		return status
	}

	home, _ := os.UserHomeDir()
	netrc := filepath.Join(home, ".netrc")
	if _, err := os.Stat(netrc); err == nil {
		status.Present = true
		return status
	}

	status.Missing = []string{"~/.netrc", "git credential helper"}
	status.Suggestion = "Configure ~/.netrc or git credential helper for private modules"
	return status
}
