package pipeline

import (
	"testing"
)

func TestParse_AllStageTypes(t *testing.T) {
	tests := []struct {
		line    string
		wantErr bool
		name    string
	}{
		{"grep foo", false, "grep"},
		{"grep -i foo", false, "grep"},
		{"grep-v foo", false, "grep-v"},
		{"grep", true, ""},          // missing pattern
		{"contains bar", false, "contains"},
		{"contains -i bar", false, "contains"},
		{"contains", true, ""},      // missing substring
		{"replace a b", false, "replace"},
		{"replace a", true, ""},     // needs two args
		{"head -n 5", false, "head"},
		{"head 5", false, "head"},
		{"head -n notnum", true, ""},
		{"take 3", false, "head"},
		{"tail -n 5", false, "tail"},
		{"tail bad", false, "tail"}, // non -n bad number ignored, defaults
		{"tail -n nope", true, ""},
		{"skip 2", false, "skip"},
		{"skip", false, "skip"}, // no arg -> N=0
		{"skip bad", true, ""},
		{"sort -r", false, "sort"},
		{"sort -n", false, "sort"},
		{"sort -rn", false, "sort"},
		{"sort -nr", false, "sort"},
		{"uniq -i", false, "uniq"},
		{"cut -d , -f 1", false, "cut"},
		{"cut -d, -f1,2", false, "cut"},
		{"cut -f 1", false, "cut"},
		{"cut", true, ""},            // missing -f
		{"cut -f notnum", true, ""},  // bad field
		{"cut -fbad", true, ""},      // bad attached field
		{"tr ab xy", false, "tr"},
		{"tr a", true, ""},
		{"sed s/a/b/g", false, "sed"},
		{"sed s/a/b/", false, "sed"},
		{"sed a b", false, "sed"},    // fallback form
		{"sed", true, ""},            // missing expr
		{"sed x", true, ""},          // invalid single-arg expr
		{"rev", false, "rev"},
		{"nl -s 3", false, "nl"},
		{"nl -s bad", true, ""},
		{"nl", false, "nl"},
		{"tee /tmp/x", false, "tee"},
		{"tee", false, "tee"},
		{"tac", false, "tac"},
		{"wc -l", false, "wc"},
		{"wc -w -c", false, "wc"},
		{"wc -m", false, "wc"},
		{"boguscmd", true, ""},
		{"", true, ""}, // empty
	}
	for _, tc := range tests {
		t.Run(tc.line, func(t *testing.T) {
			s, err := Parse(tc.line)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("Parse(%q) expected error", tc.line)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse(%q): %v", tc.line, err)
			}
			if s.Name() != tc.name {
				t.Errorf("Parse(%q).Name() = %q, want %q", tc.line, s.Name(), tc.name)
			}
		})
	}
}

func TestParseFieldSpec(t *testing.T) {
	got, err := parseFieldSpec("1, 2 ,3,")
	if err != nil {
		t.Fatalf("parseFieldSpec: %v", err)
	}
	if len(got) != 3 || got[0] != 1 || got[2] != 3 {
		t.Errorf("parseFieldSpec = %v", got)
	}
	if _, err := parseFieldSpec("x"); err == nil {
		t.Error("parseFieldSpec(x) should error")
	}
}

func TestSplitSedExpr_Escapes(t *testing.T) {
	// escaped delimiter inside the expression
	parts := splitSedExpr(`a\/b/c`, '/')
	if len(parts) != 2 {
		t.Fatalf("splitSedExpr parts = %v", parts)
	}
	if parts[0] != `a\/b` {
		t.Errorf("part0 = %q", parts[0])
	}
}

func TestParseCommandLine_Quoting(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{`grep "hello world"`, []string{"grep", "hello world"}},
		{`grep 'single quote'`, []string{"grep", "single quote"}},
		{`a\ b`, []string{"a b"}},
		{"a\tb", []string{"a", "b"}},
		{`grep ""`, []string{"grep"}},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			got := parseCommandLine(tc.in)
			if len(got) != len(tc.want) {
				t.Fatalf("parseCommandLine(%q) = %v, want %v", tc.in, got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("part %d = %q, want %q", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestParseAll_Multi(t *testing.T) {
	stages, err := ParseAll([]string{"grep foo", "sort -r", "head -n 2"})
	if err != nil {
		t.Fatalf("ParseAll: %v", err)
	}
	if len(stages) != 3 {
		t.Fatalf("got %d stages", len(stages))
	}
}
