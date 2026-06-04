package purl

import "testing"

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		in          string
		wantModule  string
		wantVersion string
		wantOK      bool
	}{
		{
			name:        "module with version",
			in:          "pkg:golang/github.com/spf13/cobra@v1.10.2",
			wantModule:  "github.com/spf13/cobra",
			wantVersion: "v1.10.2",
			wantOK:      true,
		},
		{
			name:        "module without version",
			in:          "pkg:golang/golang.org/x/mod",
			wantModule:  "golang.org/x/mod",
			wantVersion: "",
			wantOK:      true,
		},
		{
			name:        "non-golang type",
			in:          "pkg:npm/left-pad@1.0.0",
			wantModule:  "",
			wantVersion: "",
			wantOK:      false,
		},
		{
			name:        "empty string",
			in:          "",
			wantModule:  "",
			wantVersion: "",
			wantOK:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotModule, gotVersion, gotOK := Parse(tt.in)
			if gotModule != tt.wantModule || gotVersion != tt.wantVersion || gotOK != tt.wantOK {
				t.Errorf("Parse(%q) = (%q, %q, %v); want (%q, %q, %v)",
					tt.in, gotModule, gotVersion, gotOK,
					tt.wantModule, tt.wantVersion, tt.wantOK)
			}
		})
	}
}

// TestParseRoundTrip confirms Parse inverts ForModule for the canonical case.
func TestParseRoundTrip(t *testing.T) {
	const mod, ver = "github.com/spf13/cobra", "v1.10.2"
	p := ForModule(mod, ver)
	gotMod, gotVer, ok := Parse(p)
	if !ok || gotMod != mod || gotVer != ver {
		t.Errorf("round-trip ForModule->Parse(%q) = (%q, %q, %v); want (%q, %q, true)",
			p, gotMod, gotVer, ok, mod, ver)
	}
}
