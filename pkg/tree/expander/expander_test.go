package expander

import (
	"reflect"
	"sort"
	"testing"
)

func TestExpand(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		want    []string
		wantErr bool
	}{
		{
			name:    "simple brace expansion",
			pattern: "docs/{guides,apis,gov,arch}",
			want:    []string{"docs/guides", "docs/apis", "docs/gov", "docs/arch"},
			wantErr: false,
		},
		{
			name:    "no braces",
			pattern: "simple/path",
			want:    []string{"simple/path"},
			wantErr: false,
		},
		{
			name:    "nested braces",
			pattern: "a/{b,c/{d,e}}",
			want:    []string{"a/b", "a/c/d", "a/c/e"},
			wantErr: false,
		},
		{
			name:    "multiple brace groups",
			pattern: "{a,b}/{c,d}",
			want:    []string{"a/c", "a/d", "b/c", "b/d"},
			wantErr: false,
		},
		{
			name:    "single alternative",
			pattern: "path/{single}",
			want:    []string{"path/single"},
			wantErr: false,
		},
		{
			name:    "complex nested pattern",
			pattern: "src/{cmd/{main,sub},pkg/{util,core}}",
			want:    []string{"src/cmd/main", "src/cmd/sub", "src/pkg/util", "src/pkg/core"},
			wantErr: false,
		},
		{
			name:    "triple expansion",
			pattern: "{a,b}/{c,d}/{e,f}",
			want: []string{
				"a/c/e", "a/c/f", "a/d/e", "a/d/f",
				"b/c/e", "b/c/f", "b/d/e", "b/d/f",
			},
			wantErr: false,
		},
		{
			name:    "braces at start",
			pattern: "{docs,src}/content",
			want:    []string{"docs/content", "src/content"},
			wantErr: false,
		},
		{
			name:    "braces at end",
			pattern: "base/{end1,end2}",
			want:    []string{"base/end1", "base/end2"},
			wantErr: false,
		},
		{
			name:    "empty pattern",
			pattern: "",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "unmatched opening brace",
			pattern: "path/{incomplete",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "unmatched closing brace",
			pattern: "path/incomplete}",
			want:    []string{"path/incomplete}"},
			wantErr: false,
		},
		{
			name:    "empty brace content",
			pattern: "path/{}",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Expand(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Expand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Sort both slices for comparison
				sort.Strings(got)
				sort.Strings(tt.want)

				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("Expand() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestFindMatchingBrace(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		openIdx int
		want    int
	}{
		{
			name:    "simple matching brace",
			s:       "{abc}",
			openIdx: 0,
			want:    4,
		},
		{
			name:    "nested braces",
			s:       "{a{b}c}",
			openIdx: 0,
			want:    6,
		},
		{
			name:    "inner brace",
			s:       "{a{b}c}",
			openIdx: 2,
			want:    4,
		},
		{
			name:    "no matching brace",
			s:       "{abc",
			openIdx: 0,
			want:    -1,
		},
		{
			name:    "invalid index",
			s:       "abc",
			openIdx: 0,
			want:    -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := findMatchingBrace(tt.s, tt.openIdx); got != tt.want {
				t.Errorf("findMatchingBrace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSplitAlternatives(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
		wantErr bool
	}{
		{
			name:    "simple split",
			content: "a,b,c",
			want:    []string{"a", "b", "c"},
			wantErr: false,
		},
		{
			name:    "single item",
			content: "single",
			want:    []string{"single"},
			wantErr: false,
		},
		{
			name:    "nested braces",
			content: "a,{b,c},d",
			want:    []string{"a", "{b,c}", "d"},
			wantErr: false,
		},
		{
			name:    "empty content",
			content: "",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "unmatched opening brace",
			content: "a,{b,c",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "unmatched closing brace",
			content: "a,b},c",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := splitAlternatives(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("splitAlternatives() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitAlternatives() = %v, want %v", got, tt.want)
			}
		})
	}
}
