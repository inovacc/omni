package scaffolding

import (
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"
)

// Options configures the scaffolding command behavior.
type Options struct {
	JSON bool // --json: output as JSON
}

// WriteTemplate renders a Go text/template to a file at path.
func WriteTemplate(path string, tmpl string, data any) error {
	t, err := template.New("").Parse(tmpl)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	defer func() { _ = f.Close() }()

	return t.Execute(f, data)
}

// WriteLicense writes a LICENSE file with the given type and author.
// Supported types: MIT, Apache-2.0, BSD-3.
func WriteLicense(path, licenseType, author string) error {
	year := time.Now().Year()

	var content string

	switch strings.ToLower(licenseType) {
	case "mit":
		content = fmt.Sprintf(MITLicense, year, author)
	case "apache-2.0", "apache":
		content = fmt.Sprintf(ApacheLicense, year, author)
	case "bsd-3", "bsd":
		content = fmt.Sprintf(BSDLicense, year, author)
	default:
		return fmt.Errorf("unknown license type: %s", licenseType)
	}

	return os.WriteFile(path, []byte(content), 0o644)
}
