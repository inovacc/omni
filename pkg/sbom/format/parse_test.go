package format_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/inovacc/omni/pkg/sbom/format"
)

// componentByName returns the Component with the given Name, or false.
func componentByName(comps []format.Component, name string) (format.Component, bool) {
	for _, c := range comps {
		if c.Name == name {
			return c, true
		}
	}
	return format.Component{}, false
}

func TestParseRoundTripSPDX(t *testing.T) {
	doc := format.From(sample(), format.Options{OmniVersion: "v0.1.0"})
	var buf bytes.Buffer
	if err := doc.Encode(&buf, format.SPDX); err != nil {
		t.Fatalf("encode: %v", err)
	}

	parsed, err := format.Parse(&buf)
	if err != nil {
		t.Fatalf("Parse SPDX: %v", err)
	}
	comps := parsed.Components()
	cobra, ok := componentByName(comps, "github.com/spf13/cobra")
	if !ok {
		t.Fatalf("cobra component missing; got %v", comps)
	}
	if cobra.Version != "v1.10.2" {
		t.Errorf("cobra version = %q; want v1.10.2", cobra.Version)
	}
	if cobra.PURL != "pkg:golang/github.com/spf13/cobra@v1.10.2" {
		t.Errorf("cobra purl = %q", cobra.PURL)
	}
	if cobra.Ecosystem != "golang" {
		t.Errorf("cobra ecosystem = %q; want golang", cobra.Ecosystem)
	}
}

func TestParseRoundTripCycloneDX(t *testing.T) {
	doc := format.From(sample(), format.Options{OmniVersion: "v0.1.0"})
	var buf bytes.Buffer
	if err := doc.Encode(&buf, format.CycloneDX); err != nil {
		t.Fatalf("encode: %v", err)
	}

	parsed, err := format.Parse(&buf)
	if err != nil {
		t.Fatalf("Parse CycloneDX: %v", err)
	}
	comps := parsed.Components()
	mod, ok := componentByName(comps, "golang.org/x/mod")
	if !ok {
		t.Fatalf("x/mod component missing; got %v", comps)
	}
	if mod.Version != "v0.36.0" {
		t.Errorf("x/mod version = %q; want v0.36.0", mod.Version)
	}
	if mod.Ecosystem != "golang" {
		t.Errorf("x/mod ecosystem = %q; want golang", mod.Ecosystem)
	}
}

func TestParseThirdPartyCycloneDX(t *testing.T) {
	const literal = `{
  "bomFormat": "CycloneDX",
  "specVersion": "1.5",
  "metadata": {"component": {"name": "third-party-app"}},
  "components": [
    {"type": "library", "name": "cobra", "purl": "pkg:golang/github.com/spf13/cobra@v1.9.0", "unknownField": 42},
    {"type": "library", "name": "no-purl-here"}
  ],
  "anotherUnknownTopLevelField": true
}`
	parsed, err := format.Parse(strings.NewReader(literal))
	if err != nil {
		t.Fatalf("Parse third-party CycloneDX: %v", err)
	}
	comps := parsed.Components()
	if len(comps) != 1 {
		t.Fatalf("want 1 component (purl-less skipped), got %d: %v", len(comps), comps)
	}
	c, ok := componentByName(comps, "github.com/spf13/cobra")
	if !ok {
		t.Fatalf("cobra component missing; got %v", comps)
	}
	if c.Version != "v1.9.0" {
		t.Errorf("cobra version = %q; want v1.9.0", c.Version)
	}
}

func TestParseUnrecognizedFormat(t *testing.T) {
	const literal = `{"foo": "bar", "baz": 1}`
	if _, err := format.Parse(strings.NewReader(literal)); err == nil {
		t.Fatal("Parse should fail for unrecognized format")
	}
}

func TestParseInvalidJSON(t *testing.T) {
	if _, err := format.Parse(strings.NewReader("{not json")); err == nil {
		t.Fatal("Parse should fail for invalid JSON")
	}
}
