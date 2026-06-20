package file

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDetectTextType exercises the text-type sniffer over shebangs, extensions
// and content patterns.
func TestDetectTextType(t *testing.T) {
	dir := t.TempDir()

	// Create a real JSON file so the JSON content branch can re-open it.
	jsonPath := filepath.Join(dir, "data.txt")
	if err := os.WriteFile(jsonPath, []byte("{\n  \"a\": 1\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		path     string
		buf      string
		wantType string
		wantMime string
	}{
		{"python shebang", "script", "#!/usr/bin/python\nprint(1)\n", "Python script, ASCII text executable", "text/x-python"},
		{"bash shebang", "s", "#!/bin/bash\necho hi\n", "Bourne-Again shell script, ASCII text executable", "text/x-shellscript"},
		{"sh shebang", "s", "#!/bin/sh\necho hi\n", "Bourne-Again shell script, ASCII text executable", "text/x-shellscript"},
		{"perl shebang", "s", "#!/usr/bin/perl\n", "Perl script, ASCII text executable", "text/x-perl"},
		{"ruby shebang", "s", "#!/usr/bin/ruby\n", "Ruby script, ASCII text executable", "text/x-ruby"},
		{"node shebang", "s", "#!/usr/bin/env node\n", "Node.js script, ASCII text executable", "text/javascript"},
		{"generic shebang", "s", "#!/unknown/thing\n", "script, ASCII text executable", "text/plain"},
		{"go ext", "main.go", "package main\n", "Go source, ASCII text", "text/x-go"},
		{"py ext", "a.py", "x=1\n", "Python script, ASCII text", "text/x-python"},
		{"js ext", "a.js", "var x;\n", "JavaScript source, ASCII text", "text/javascript"},
		{"ts ext", "a.ts", "let x;\n", "TypeScript source, ASCII text", "text/typescript"},
		{"json ext", "a.json", "{}", "JSON data", "application/json"},
		{"xml ext", "a.xml", "<a/>", "XML document", "application/xml"},
		{"html ext", "a.html", "<p>", "HTML document, ASCII text", "text/html"},
		{"css ext", "a.css", "a{}", "CSS stylesheet, ASCII text", "text/css"},
		{"md ext", "a.md", "# x", "Markdown document, ASCII text", "text/markdown"},
		{"yaml ext", "a.yaml", "a: 1", "YAML document, ASCII text", "text/yaml"},
		{"sh ext", "a.sh", "echo", "Bourne-Again shell script, ASCII text", "text/x-shellscript"},
		{"c ext", "a.c", "int", "C source, ASCII text", "text/x-c"},
		{"h ext", "a.h", "int", "C header, ASCII text", "text/x-c"},
		{"cpp ext", "a.cpp", "int", "C++ source, ASCII text", "text/x-c++"},
		{"java ext", "a.java", "class", "Java source, ASCII text", "text/x-java"},
		{"rs ext", "a.rs", "fn", "Rust source, ASCII text", "text/x-rust"},
		{"xml content", "noext", "<?xml version=\"1.0\"?>", "XML document", "application/xml"},
		{"html content", "noext", "<!DOCTYPE html>", "HTML document, ASCII text", "text/html"},
		{"plain", "noext", "just words\n", "ASCII text", "text/plain"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotMime := detectTextType(tt.path, []byte(tt.buf))
			if gotType != tt.wantType {
				t.Errorf("type = %q, want %q", gotType, tt.wantType)
			}
			if gotMime != tt.wantMime {
				t.Errorf("mime = %q, want %q", gotMime, tt.wantMime)
			}
		})
	}

	// JSON content branch that re-opens the file from disk.
	gotType, gotMime := detectTextType(jsonPath, []byte("{\n  \"a\": 1\n}\n"))
	if gotType != "JSON data" || gotMime != "application/json" {
		t.Errorf("json file branch = %q/%q", gotType, gotMime)
	}
}

// TestDescribeELF exercises the ELF header describer for bitness/endianness/type.
func TestDescribeELF(t *testing.T) {
	mk := func(class, data byte, etype uint16, little bool) []byte {
		b := make([]byte, 24)
		b[0], b[1], b[2], b[3] = 0x7f, 'E', 'L', 'F'
		b[4] = class
		b[5] = data
		if little {
			b[16] = byte(etype)
			b[17] = byte(etype >> 8)
		} else {
			b[16] = byte(etype >> 8)
			b[17] = byte(etype)
		}
		return b
	}

	tests := []struct {
		name string
		buf  []byte
		want string
	}{
		{"64-bit LSB executable", mk(2, 1, 2, true), "ELF 64-bit LSB executable"},
		{"32-bit LSB relocatable", mk(1, 1, 1, true), "ELF 32-bit LSB relocatable"},
		{"64-bit MSB shared object", mk(2, 2, 3, false), "ELF 64-bit MSB shared object"},
		{"64-bit LSB core file", mk(2, 1, 4, true), "ELF 64-bit LSB core file"},
		{"too short", []byte{0x7f, 'E', 'L', 'F'}, "ELF"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := describeELF(tt.buf); got != tt.want {
				t.Errorf("describeELF = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestCheckMagicViaDetect drives detectFileType over real fixture files holding
// representative magic-byte headers, including the ELF special-case.
func TestCheckMagicViaDetect(t *testing.T) {
	dir := t.TempDir()

	// A full 64-bit LSB executable ELF header so describeELF is reached.
	elf := make([]byte, 64)
	elf[0], elf[1], elf[2], elf[3] = 0x7f, 'E', 'L', 'F'
	elf[4] = 2 // 64-bit
	elf[5] = 1 // LSB
	elf[16] = 2 // executable

	fixtures := []struct {
		name string
		data []byte
		want string
	}{
		{"png.bin", []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A, 0, 0, 0, 0}, "PNG image data"},
		{"data.gz", []byte{0x1F, 0x8B, 0x08, 0x00, 0, 0, 0, 0}, "gzip compressed data"},
		{"prog.elf", elf, "ELF 64-bit LSB executable"},
	}

	for _, fx := range fixtures {
		t.Run(fx.name, func(t *testing.T) {
			p := filepath.Join(dir, fx.name)
			if err := os.WriteFile(p, fx.data, 0o644); err != nil {
				t.Fatal(err)
			}
			gotType, _ := detectFileType(p, false)
			if !strings.HasPrefix(gotType, fx.want) {
				t.Errorf("detectFileType(%s) = %q, want prefix %q", fx.name, gotType, fx.want)
			}
		})
	}
}

// TestDetectFileTypeSpecial covers non-content branches: missing file, dir, empty.
func TestDetectFileTypeSpecial(t *testing.T) {
	dir := t.TempDir()

	t.Run("missing", func(t *testing.T) {
		ft, mt := detectFileType(filepath.Join(dir, "nope"), false)
		if !strings.Contains(ft, "cannot open") || mt != "application/x-not-found" {
			t.Errorf("missing = %q/%q", ft, mt)
		}
	})

	t.Run("directory", func(t *testing.T) {
		ft, mt := detectFileType(dir, false)
		if ft != "directory" || mt != "inode/directory" {
			t.Errorf("dir = %q/%q", ft, mt)
		}
	})

	t.Run("empty", func(t *testing.T) {
		p := filepath.Join(dir, "empty")
		if err := os.WriteFile(p, nil, 0o644); err != nil {
			t.Fatal(err)
		}
		ft, mt := detectFileType(p, false)
		if ft != "empty" || mt != "inode/x-empty" {
			t.Errorf("empty = %q/%q", ft, mt)
		}
	})

	t.Run("plain text", func(t *testing.T) {
		p := filepath.Join(dir, "note.txt")
		if err := os.WriteFile(p, []byte("hello world\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		ft, mt := detectFileType(p, false)
		if ft != "ASCII text" || mt != "text/plain" {
			t.Errorf("text = %q/%q", ft, mt)
		}
	})
}
