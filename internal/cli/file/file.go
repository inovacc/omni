package file

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/inovacc/omni/internal/cli/output"
)

// FileOptions configures the file command behavior
type FileOptions struct {
	Brief        bool          // -b: do not prepend filenames
	MimeType     bool          // -i: output MIME type
	NoDeref      bool          // -h: don't follow symlinks
	Separator    string        // -F: use string as separator
	OutputFormat output.Format // output format (text, json, table)
}

// FileResult represents file output for JSON
type FileResult struct {
	Path     string `json:"path"`
	Type     string `json:"type"`
	MimeType string `json:"mime_type"`
}

// RunFile determines file type
func RunFile(w io.Writer, args []string, opts FileOptions) error {
	if len(args) == 0 {
		return fmt.Errorf("file: missing file operand")
	}

	if opts.Separator == "" {
		opts.Separator = ":"
	}

	f := output.New(w, opts.OutputFormat)
	jsonMode := f.IsJSON()

	var jsonResults []FileResult

	for _, path := range args {
		fileType, mimeType := detectFileType(path, opts.NoDeref)

		if jsonMode {
			jsonResults = append(jsonResults, FileResult{Path: path, Type: fileType, MimeType: mimeType})
			continue
		}

		var out string

		if opts.MimeType {
			out = mimeType
		} else {
			out = fileType
		}

		if opts.Brief {
			_, _ = fmt.Fprintln(w, out)
		} else {
			_, _ = fmt.Fprintf(w, "%s%s %s\n", path, opts.Separator, out)
		}
	}

	if jsonMode {
		return f.Print(jsonResults)
	}

	return nil
}

func detectFileType(path string, noDeref bool) (string, string) {
	var (
		info os.FileInfo
		err  error
	)

	if noDeref {
		info, err = os.Lstat(path)
	} else {
		info, err = os.Stat(path)
	}

	if err != nil {
		if os.IsNotExist(err) {
			return "cannot open (No such file or directory)", "application/x-not-found"
		}

		return fmt.Sprintf("cannot open (%v)", err), "application/x-error"
	}

	// Check file mode
	mode := info.Mode()

	if mode&os.ModeSymlink != 0 {
		target, err := os.Readlink(path)
		if err != nil {
			return "symbolic link", "inode/symlink"
		}

		return fmt.Sprintf("symbolic link to %s", target), "inode/symlink"
	}

	if mode.IsDir() {
		return "directory", "inode/directory"
	}

	if mode&os.ModeNamedPipe != 0 {
		return "fifo (named pipe)", "inode/fifo"
	}

	if mode&os.ModeSocket != 0 {
		return "socket", "inode/socket"
	}

	if mode&os.ModeDevice != 0 {
		if mode&os.ModeCharDevice != 0 {
			return "character special", "inode/chardevice"
		}

		return "block special", "inode/blockdevice"
	}

	if !mode.IsRegular() {
		return "unknown", "application/octet-stream"
	}

	// Check if empty
	if info.Size() == 0 {
		return "empty", "inode/x-empty"
	}

	// Read file header for magic detection
	return detectByContent(path)
}

func detectByContent(path string) (string, string) {
	f, err := os.Open(path)
	if err != nil {
		return "regular file", "application/octet-stream"
	}

	defer func() { _ = f.Close() }()

	// Read first 512 bytes for magic detection
	buf := make([]byte, 512)
	n, _ := f.Read(buf)
	buf = buf[:n]

	if n == 0 {
		return "empty", "inode/x-empty"
	}

	// Check magic bytes
	if fileType, mimeType, ok := checkMagic(buf); ok {
		return fileType, mimeType
	}

	// Check if text
	if isText(buf) {
		return detectTextType(path, buf)
	}

	return "data", "application/octet-stream"
}

func checkMagic(buf []byte) (string, string, bool) {
	magics := []struct {
		magic    []byte
		fileType string
		mimeType string
	}{
		// Images
		{[]byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}, "PNG image data", "image/png"},
		{[]byte{0xFF, 0xD8, 0xFF}, "JPEG image data", "image/jpeg"},
		{[]byte{'G', 'I', 'F', '8', '7', 'a'}, "GIF image data, version 87a", "image/gif"},
		{[]byte{'G', 'I', 'F', '8', '9', 'a'}, "GIF image data, version 89a", "image/gif"},
		{[]byte{'B', 'M'}, "BMP image data", "image/bmp"},
		{[]byte{'R', 'I', 'F', 'F'}, "RIFF data", "application/octet-stream"}, // Could be WAV, AVI, WebP

		// Archives
		{[]byte{0x1F, 0x8B}, "gzip compressed data", "application/gzip"},
		{[]byte{'P', 'K', 0x03, 0x04}, "Zip archive data", "application/zip"},
		{[]byte{'P', 'K', 0x05, 0x06}, "Zip archive data (empty)", "application/zip"},
		{[]byte{0x42, 0x5A, 0x68}, "bzip2 compressed data", "application/x-bzip2"},
		{[]byte{0xFD, '7', 'z', 'X', 'Z', 0x00}, "XZ compressed data", "application/x-xz"},
		{[]byte{0x5D, 0x00, 0x00}, "LZMA compressed data", "application/x-lzma"},
		{[]byte{'R', 'a', 'r', '!', 0x1A, 0x07}, "RAR archive data", "application/x-rar-compressed"},
		{[]byte{'7', 'z', 0xBC, 0xAF, 0x27, 0x1C}, "7-zip archive data", "application/x-7z-compressed"},

		// Documents
		{[]byte{'%', 'P', 'D', 'F'}, "PDF document", "application/pdf"},
		{[]byte{0xD0, 0xCF, 0x11, 0xE0}, "Microsoft Office document", "application/msword"},

		// Executables
		{[]byte{0x7F, 'E', 'L', 'F'}, "ELF", "application/x-executable"},
		{[]byte{'M', 'Z'}, "PE32 executable", "application/x-dosexec"},
		{[]byte{0xCA, 0xFE, 0xBA, 0xBE}, "Mach-O universal binary", "application/x-mach-binary"},
		{[]byte{0xCF, 0xFA, 0xED, 0xFE}, "Mach-O 64-bit executable", "application/x-mach-binary"},
		{[]byte{0xCE, 0xFA, 0xED, 0xFE}, "Mach-O 32-bit executable", "application/x-mach-binary"},

		// Audio/Video
		{[]byte{0x00, 0x00, 0x00}, "MPEG video", "video/mpeg"}, // Simplified
		{[]byte{'I', 'D', '3'}, "Audio file with ID3", "audio/mpeg"},
		{[]byte{0xFF, 0xFB}, "MPEG audio", "audio/mpeg"},
		{[]byte{'O', 'g', 'g', 'S'}, "Ogg data", "application/ogg"},
		{[]byte{'f', 'L', 'a', 'C'}, "FLAC audio", "audio/flac"},

		// Other
		{[]byte{0x00, 0x61, 0x73, 0x6D}, "WebAssembly binary", "application/wasm"},
		{[]byte{'S', 'Q', 'L', 'i', 't', 'e'}, "SQLite database", "application/x-sqlite3"},
	}

	for _, m := range magics {
		if bytes.HasPrefix(buf, m.magic) {
			fileType := m.fileType

			// Special handling for ELF
			if m.fileType == "ELF" && len(buf) > 16 {
				fileType = describeELF(buf)
			}

			return fileType, m.mimeType, true
		}
	}

	// Check for tar (magic at offset 257)
	if len(buf) > 262 && string(buf[257:262]) == "ustar" {
		return "POSIX tar archive", "application/x-tar", true
	}

	return "", "", false
}

func describeELF(buf []byte) string {
	if len(buf) < 20 {
		return "ELF"
	}

	var result strings.Builder

	result.WriteString("ELF ")

	// 32 or 64 bit
	switch buf[4] {
	case 1:
		result.WriteString("32-bit ")
	case 2:
		result.WriteString("64-bit ")
	}

	// Endianness
	switch buf[5] {
	case 1:
		result.WriteString("LSB ")
	case 2:
		result.WriteString("MSB ")
	}

	// Type
	var elfType uint16

	if buf[5] == 1 { // Little endian
		elfType = binary.LittleEndian.Uint16(buf[16:18])
	} else {
		elfType = binary.BigEndian.Uint16(buf[16:18])
	}

	switch elfType {
	case 1:
		result.WriteString("relocatable")
	case 2:
		result.WriteString("executable")
	case 3:
		result.WriteString("shared object")
	case 4:
		result.WriteString("core file")
	}

	return result.String()
}

func isText(buf []byte) bool {
	// Check if mostly printable ASCII or valid UTF-8
	textChars := 0

	for _, b := range buf {
		if b == '\t' || b == '\n' || b == '\r' || (b >= 32 && b < 127) {
			textChars++
		} else if b >= 128 {
			// Could be UTF-8, count as text
			textChars++
		}
	}

	return float64(textChars)/float64(len(buf)) > 0.85
}

func detectTextType(path string, buf []byte) (string, string) {
	content := string(buf)
	ext := strings.ToLower(filepath.Ext(path))

	// Check shebang
	if strings.HasPrefix(content, "#!") {
		line := strings.SplitN(content, "\n", 2)[0]

		if strings.Contains(line, "python") {
			return "Python script, ASCII text executable", "text/x-python"
		}

		if strings.Contains(line, "bash") || strings.Contains(line, "/sh") {
			return "Bourne-Again shell script, ASCII text executable", "text/x-shellscript"
		}

		if strings.Contains(line, "perl") {
			return "Perl script, ASCII text executable", "text/x-perl"
		}

		if strings.Contains(line, "ruby") {
			return "Ruby script, ASCII text executable", "text/x-ruby"
		}

		if strings.Contains(line, "node") {
			return "Node.js script, ASCII text executable", "text/javascript"
		}

		return "script, ASCII text executable", "text/plain"
	}

	// Check by extension and content
	switch ext {
	case ".go":
		return "Go source, ASCII text", "text/x-go"
	case ".py":
		return "Python script, ASCII text", "text/x-python"
	case ".js":
		return "JavaScript source, ASCII text", "text/javascript"
	case ".ts":
		return "TypeScript source, ASCII text", "text/typescript"
	case ".json":
		return "JSON data", "application/json"
	case ".xml":
		return "XML document", "application/xml"
	case ".html", ".htm":
		return "HTML document, ASCII text", "text/html"
	case ".css":
		return "CSS stylesheet, ASCII text", "text/css"
	case ".md":
		return "Markdown document, ASCII text", "text/markdown"
	case ".yaml", ".yml":
		return "YAML document, ASCII text", "text/yaml"
	case ".sh":
		return "Bourne-Again shell script, ASCII text", "text/x-shellscript"
	case ".c":
		return "C source, ASCII text", "text/x-c"
	case ".h":
		return "C header, ASCII text", "text/x-c"
	case ".cpp", ".cc", ".cxx":
		return "C++ source, ASCII text", "text/x-c++"
	case ".java":
		return "Java source, ASCII text", "text/x-java"
	case ".rs":
		return "Rust source, ASCII text", "text/x-rust"
	}

	// Check content patterns
	if strings.HasPrefix(content, "<?xml") {
		return "XML document", "application/xml"
	}

	if strings.HasPrefix(content, "<!DOCTYPE html") || strings.HasPrefix(content, "<html") {
		return "HTML document, ASCII text", "text/html"
	}

	if strings.HasPrefix(content, "{") || strings.HasPrefix(content, "[") {
		// Might be JSON - check first non-empty line
		f, err := os.Open(path)
		if err == nil {
			scanner := bufio.NewScanner(f)
			if scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line != "" && (strings.HasPrefix(line, "{") || strings.HasPrefix(line, "[")) {
					_ = f.Close()

					return "JSON data", "application/json"
				}
			}

			_ = f.Close()
		}
	}

	return "ASCII text", "text/plain"
}
