package purl_test

import (
	"testing"

	"github.com/inovacc/omni/pkg/sbom/purl"
)

func TestForModule(t *testing.T) {
	cases := []struct {
		name, path, version, want string
	}{
		{"tagged", "github.com/spf13/cobra", "v1.10.2", "pkg:golang/github.com/spf13/cobra@v1.10.2"},
		{"shorthand canonicalized", "golang.org/x/mod", "v0.36", "pkg:golang/golang.org/x/mod@v0.36.0"},
		{"uppercase lowered", "github.com/BurntSushi/toml", "v1.6.0", "pkg:golang/github.com/burntsushi/toml@v1.6.0"},
		{"pseudo passthrough", "github.com/dop251/goja", "v0.0.0-20260106131823-651366fbe6e3", "pkg:golang/github.com/dop251/goja@v0.0.0-20260106131823-651366fbe6e3"},
		{"incompatible preserved", "github.com/foo/bar", "v2.0.0+incompatible", "pkg:golang/github.com/foo/bar@v2.0.0+incompatible"},
		{"empty version no suffix", "example.com/local", "", "pkg:golang/example.com/local"},
		{"std toolchain", "std", "go1.25.0", "pkg:golang/std@go1.25.0"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := purl.ForModule(c.path, c.version); got != c.want {
				t.Errorf("ForModule(%q,%q) = %q, want %q", c.path, c.version, got, c.want)
			}
		})
	}
}

func TestForModuleRejectsEncodingNeeded(t *testing.T) {
	// Go module paths never need percent-encoding once lowercased; guard the assumption.
	got := purl.ForModule("github.com/a_b/c.d-e", "v1.0.0")
	if got != "pkg:golang/github.com/a_b/c.d-e@v1.0.0" {
		t.Errorf("got %q", got)
	}
}
