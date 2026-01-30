package xmlutil

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunToXML(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		opts    ToXMLOptions
		want    string
		wantErr bool
	}{
		{
			name:  "simple object",
			input: `{"name":"John","age":30}`,
			opts:  ToXMLOptions{Root: "root"},
			want: `<?xml version="1.0" encoding="UTF-8"?>
<root>
  <age>30</age>
  <name>John</name>
</root>
`,
		},
		{
			name:  "nested object",
			input: `{"user":{"name":"John","address":{"city":"NYC"}}}`,
			opts:  ToXMLOptions{Root: "root"},
			want: `<?xml version="1.0" encoding="UTF-8"?>
<root>
  <user>
    <address>
      <city>NYC</city>
    </address>
    <name>John</name>
  </user>
</root>
`,
		},
		{
			name:  "array at root",
			input: `[1,2,3]`,
			opts:  ToXMLOptions{Root: "root", ItemTag: "item"},
			want: `<?xml version="1.0" encoding="UTF-8"?>
<root>
  <item>1</item>
  <item>2</item>
  <item>3</item>
</root>
`,
		},
		{
			name:  "array in object",
			input: `{"items":[1,2,3]}`,
			opts:  ToXMLOptions{Root: "root", ItemTag: "item"},
			want: `<?xml version="1.0" encoding="UTF-8"?>
<root>
  <items>
    <item>1</item>
    <item>2</item>
    <item>3</item>
  </items>
</root>
`,
		},
		{
			name:  "with attributes",
			input: `{"-id":"123","name":"John"}`,
			opts:  ToXMLOptions{Root: "root", AttrPrefix: "-"},
			want: `<?xml version="1.0" encoding="UTF-8"?>
<root id="123">
  <name>John</name>
</root>
`,
		},
		{
			name:  "null value",
			input: `{"value":null}`,
			opts:  ToXMLOptions{Root: "root"},
			want: `<?xml version="1.0" encoding="UTF-8"?>
<root>
  <value/>
</root>
`,
		},
		{
			name:  "boolean values",
			input: `{"active":true,"deleted":false}`,
			opts:  ToXMLOptions{Root: "root"},
			want: `<?xml version="1.0" encoding="UTF-8"?>
<root>
  <active>true</active>
  <deleted>false</deleted>
</root>
`,
		},
		{
			name:  "special characters",
			input: `{"text":"<hello> & \"world\""}`,
			opts:  ToXMLOptions{Root: "root"},
			want: `<?xml version="1.0" encoding="UTF-8"?>
<root>
  <text>&lt;hello&gt; &amp; &quot;world&quot;</text>
</root>
`,
		},
		{
			name:  "custom root and indent",
			input: `{"name":"John"}`,
			opts:  ToXMLOptions{Root: "person", Indent: "    "},
			want: `<?xml version="1.0" encoding="UTF-8"?>
<person>
    <name>John</name>
</person>
`,
		},
		{
			name:    "invalid JSON",
			input:   `not json`,
			opts:    ToXMLOptions{Root: "root"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			r := strings.NewReader(tt.input)

			err := RunToXML(&buf, r, nil, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunToXML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				got := buf.String()
				if got != tt.want {
					t.Errorf("RunToXML() =\n%q\nwant\n%q", got, tt.want)
				}
			}
		})
	}
}

func TestRunFromXML(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		opts    FromXMLOptions
		want    string
		wantErr bool
	}{
		{
			name:  "simple element",
			input: `<root><name>John</name><age>30</age></root>`,
			opts:  FromXMLOptions{},
			want: `{
  "root": {
    "age": "30",
    "name": "John"
  }
}
`,
		},
		{
			name:  "nested elements",
			input: `<root><user><name>John</name></user></root>`,
			opts:  FromXMLOptions{},
			want: `{
  "root": {
    "user": {
      "name": "John"
    }
  }
}
`,
		},
		{
			name:  "with attributes",
			input: `<root id="123"><name>John</name></root>`,
			opts:  FromXMLOptions{AttrPrefix: "-"},
			want: `{
  "root": {
    "-id": "123",
    "name": "John"
  }
}
`,
		},
		{
			name:  "repeated elements as array",
			input: `<root><item>1</item><item>2</item><item>3</item></root>`,
			opts:  FromXMLOptions{},
			want: `{
  "root": {
    "item": [
      "1",
      "2",
      "3"
    ]
  }
}
`,
		},
		{
			name:  "empty element",
			input: `<root><empty/></root>`,
			opts:  FromXMLOptions{},
			want: `{
  "root": {
    "empty": null
  }
}
`,
		},
		{
			name:  "text with children",
			input: `<root>text<child>value</child></root>`,
			opts:  FromXMLOptions{TextKey: "#text"},
			want: `{
  "root": {
    "#text": "text",
    "child": "value"
  }
}
`,
		},
		{
			name:  "custom attr prefix",
			input: `<root attr="value"><name>John</name></root>`,
			opts:  FromXMLOptions{AttrPrefix: "@"},
			want: `{
  "root": {
    "@attr": "value",
    "name": "John"
  }
}
`,
		},
		{
			name:    "invalid XML",
			input:   `<root><unclosed>`,
			opts:    FromXMLOptions{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			r := strings.NewReader(tt.input)

			err := RunFromXML(&buf, r, nil, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunFromXML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				got := buf.String()
				if got != tt.want {
					t.Errorf("RunFromXML() =\n%q\nwant\n%q", got, tt.want)
				}
			}
		})
	}
}

func TestEscapeXML(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "hello"},
		{"<tag>", "&lt;tag&gt;"},
		{"a & b", "a &amp; b"},
		{`"quoted"`, "&quot;quoted&quot;"},
		{"it's", "it&apos;s"},
		{"<a & b>", "&lt;a &amp; b&gt;"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := escapeXML(tt.input)
			if got != tt.want {
				t.Errorf("escapeXML(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSortedKeys(t *testing.T) {
	m := map[string]any{
		"zebra": 1,
		"apple": 2,
		"mango": 3,
	}

	got := sortedKeys(m)
	want := []string{"apple", "mango", "zebra"}

	if len(got) != len(want) {
		t.Errorf("sortedKeys() = %v, want %v", got, want)
		return
	}

	for i, k := range got {
		if k != want[i] {
			t.Errorf("sortedKeys()[%d] = %q, want %q", i, k, want[i])
		}
	}
}

func TestSeparateAttrsAndChildren(t *testing.T) {
	obj := map[string]any{
		"-id":    "123",
		"-class": "active",
		"name":   "John",
		"age":    30,
	}

	attrs, children := separateAttrsAndChildren(obj, "-")

	if len(attrs) != 2 {
		t.Errorf("attrs count = %d, want 2", len(attrs))
	}

	if attrs["id"] != "123" {
		t.Errorf("attrs[id] = %v, want 123", attrs["id"])
	}

	if attrs["class"] != "active" {
		t.Errorf("attrs[class] = %v, want active", attrs["class"])
	}

	if len(children) != 2 {
		t.Errorf("children count = %d, want 2", len(children))
	}

	if children["name"] != "John" {
		t.Errorf("children[name] = %v, want John", children["name"])
	}
}
