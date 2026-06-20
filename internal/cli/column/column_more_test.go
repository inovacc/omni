package column

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// TestColumnJSON exercises the JSON output path with default and custom headers,
// merge and no-merge field splitting.
func TestColumnJSON(t *testing.T) {
	t.Run("default headers", func(t *testing.T) {
		var buf bytes.Buffer
		in := strings.NewReader("a b c\nx y z\n")
		opts := ColumnOptions{JSON: true, Separator: " \t"}
		if err := RunColumn(&buf, in, nil, opts); err != nil {
			t.Fatal(err)
		}
		var rows []map[string]string
		if err := json.Unmarshal(buf.Bytes(), &rows); err != nil {
			t.Fatalf("unmarshal: %v\n%s", err, buf.String())
		}
		if len(rows) != 2 {
			t.Fatalf("got %d rows, want 2", len(rows))
		}
		if rows[0]["col1"] != "a" || rows[0]["col3"] != "c" {
			t.Errorf("row0 = %+v", rows[0])
		}
	})

	t.Run("custom headers", func(t *testing.T) {
		var buf bytes.Buffer
		in := strings.NewReader("alice 30\nbob 25\n")
		opts := ColumnOptions{JSON: true, Separator: " \t", ColumnHeaders: "name,age"}
		if err := RunColumn(&buf, in, nil, opts); err != nil {
			t.Fatal(err)
		}
		var rows []map[string]string
		if err := json.Unmarshal(buf.Bytes(), &rows); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if rows[0]["name"] != "alice" || rows[0]["age"] != "30" {
			t.Errorf("row0 = %+v", rows[0])
		}
	})

	t.Run("no merge", func(t *testing.T) {
		var buf bytes.Buffer
		in := strings.NewReader("a,,b\n")
		opts := ColumnOptions{JSON: true, Separator: ",", NoMerge: true}
		if err := RunColumn(&buf, in, nil, opts); err != nil {
			t.Fatal(err)
		}
		var rows []map[string]string
		if err := json.Unmarshal(buf.Bytes(), &rows); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		// "a,,b" split on "," with NoMerge keeps the empty middle field => 3 cols.
		if len(rows[0]) != 3 {
			t.Errorf("expected 3 columns, got %+v", rows[0])
		}
	})
}

// TestColumnJSONDirect calls columnJSON directly for the headers+merge matrix.
func TestColumnJSONDirect(t *testing.T) {
	var buf bytes.Buffer
	lines := []string{"k1 v1", "k2 v2"}
	opts := ColumnOptions{Separator: " \t", ColumnHeaders: "key,val"}
	if err := columnJSON(&buf, lines, opts); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), `"key":"k1"`) {
		t.Errorf("missing key header: %s", buf.String())
	}
}
