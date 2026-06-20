package cat

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// TestCatReaderJSON exercises the JSON line-collection path including numbering,
// show-tabs, show-ends, show-nonprint and squeeze-blank.
func TestCatReaderJSON(t *testing.T) {
	tests := []struct {
		name string
		in   string
		opts CatOptions
		want []CatLine
	}{
		{
			name: "plain",
			in:   "a\nb\n",
			opts: CatOptions{},
			want: []CatLine{{Content: "a"}, {Content: "b"}},
		},
		{
			name: "number all",
			in:   "a\n\nb\n",
			opts: CatOptions{NumberAll: true},
			want: []CatLine{{Number: 1, Content: "a"}, {Number: 2, Content: ""}, {Number: 3, Content: "b"}},
		},
		{
			name: "number non-blank",
			in:   "a\n\nb\n",
			opts: CatOptions{NumberNonBlank: true},
			want: []CatLine{{Number: 1, Content: "a"}, {Content: ""}, {Number: 2, Content: "b"}},
		},
		{
			name: "show tabs and ends",
			in:   "x\ty\n",
			opts: CatOptions{ShowTabs: true, ShowEnds: true},
			want: []CatLine{{Content: "x^Iy$"}},
		},
		{
			name: "squeeze blank",
			in:   "a\n\n\n\nb\n",
			opts: CatOptions{SqueezeBlank: true},
			want: []CatLine{{Content: "a"}, {Content: ""}, {Content: "b"}},
		},
		{
			name: "show nonprint control",
			in:   "a\x01b\n",
			opts: CatOptions{ShowNonPrint: true},
			want: []CatLine{{Content: "a^Ab"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := catReaderJSON(strings.NewReader(tt.in), tt.opts)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %d lines, want %d: %+v", len(got), len(tt.want), got)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("line %d = %+v, want %+v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// TestRunCatJSON drives RunCat in JSON mode and verifies a decodable array.
func TestRunCatJSON(t *testing.T) {
	var buf bytes.Buffer
	in := strings.NewReader("one\ntwo\n")
	if err := RunCat(&buf, in, nil, CatOptions{JSON: true, NumberAll: true}); err != nil {
		t.Fatal(err)
	}

	var lines []CatLine
	if err := json.Unmarshal(buf.Bytes(), &lines); err != nil {
		t.Fatalf("unmarshal: %v\n%s", err, buf.String())
	}
	if len(lines) != 2 || lines[0].Content != "one" || lines[1].Number != 2 {
		t.Errorf("unexpected lines: %+v", lines)
	}
}

// TestShowNonPrintable covers control, DEL, high-bit and M-^ notations.
func TestShowNonPrintable(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"tab passthrough", "\t", "\t"},
		{"control SOH", "\x01", "^A"},
		{"del", "\x7f", "^?"},
		{"meta control", string(rune(0x80)), "M-^@"},       // 0x80: r<160 -> M-^<r-128+64>
		{"meta high", string(rune(0x00E9)), "M-i"},          // é U+00E9 (233): r>=160 -> M-<233-128=105='i'>
		{"printable", "abc", "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := showNonPrintable(tt.in); got != tt.want {
				t.Errorf("showNonPrintable(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestRunCatShowAll covers the text path with -A-like flags through RunCat.
func TestRunCatShowAll(t *testing.T) {
	var buf bytes.Buffer
	in := strings.NewReader("a\tb\x01\n")
	opts := CatOptions{ShowTabs: true, ShowNonPrint: true, ShowEnds: true, NumberAll: true}
	if err := RunCat(&buf, in, nil, opts); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "^I") || !strings.Contains(out, "^A") || !strings.Contains(out, "$") {
		t.Errorf("missing expected markers: %q", out)
	}
	if !strings.Contains(out, "     1\t") {
		t.Errorf("missing line number: %q", out)
	}
}
