package format_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/inovacc/omni/pkg/sbom/format"
	"github.com/inovacc/omni/pkg/sbom/model"
)

func sample() *model.SBOM {
	return &model.SBOM{
		Name: "omni",
		Root: model.Component{Path: "github.com/inovacc/omni", Version: "v1.0.0", Kind: model.KindRoot},
		Components: []model.Component{
			{Path: "github.com/spf13/cobra", Version: "v1.10.2", Kind: model.KindLibrary},
			{Path: "golang.org/x/mod", Version: "v0.36.0", Kind: model.KindLibrary},
		},
	}
}

func TestEmitDeterministic(t *testing.T) {
	doc := format.From(sample(), format.Options{OmniVersion: "v0.1.0", SourceDate: "1970-01-01T00:00:00Z"})
	for _, f := range []format.Kind{format.SPDX, format.CycloneDX} {
		var a, b bytes.Buffer
		if err := doc.Encode(&a, f); err != nil {
			t.Fatalf("encode %v: %v", f, err)
		}
		if err := doc.Encode(&b, f); err != nil {
			t.Fatalf("encode2 %v: %v", f, err)
		}
		if !bytes.Equal(a.Bytes(), b.Bytes()) {
			t.Errorf("%v: two encodes differ (non-deterministic)", f)
		}
		if !bytes.HasSuffix(a.Bytes(), []byte("\n")) {
			t.Errorf("%v: missing trailing newline", f)
		}
	}
}

func TestSPDXShape(t *testing.T) {
	doc := format.From(sample(), format.Options{OmniVersion: "v0.1.0"})
	var buf bytes.Buffer
	if err := doc.Encode(&buf, format.SPDX); err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if m["spdxVersion"] != "SPDX-2.3" || m["dataLicense"] != "CC0-1.0" {
		t.Errorf("bad header: %v / %v", m["spdxVersion"], m["dataLicense"])
	}
	if !strings.Contains(buf.String(), "pkg:golang/github.com/spf13/cobra@v1.10.2") {
		t.Error("missing cobra purl")
	}
}

func TestCycloneDXShape(t *testing.T) {
	doc := format.From(sample(), format.Options{OmniVersion: "v0.1.0"})
	var buf bytes.Buffer
	if err := doc.Encode(&buf, format.CycloneDX); err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	if !strings.Contains(s, `"bomFormat": "CycloneDX"`) || !strings.Contains(s, `"specVersion": "1.5"`) {
		t.Errorf("bad header: %s", s[:120])
	}
	if strings.Contains(s, "serialNumber") {
		t.Error("serialNumber must be omitted for determinism")
	}
}
