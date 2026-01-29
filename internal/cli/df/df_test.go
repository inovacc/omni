package df

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunDF(t *testing.T) {
	t.Run("default output", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunDF(&buf, []string{}, DFOptions{})
		if err != nil {
			t.Fatalf("RunDF() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Filesystem") {
			t.Errorf("RunDF() should contain 'Filesystem' header: %s", output)
		}

		if !strings.Contains(output, "1K-blocks") {
			t.Errorf("RunDF() should contain '1K-blocks' header: %s", output)
		}
	})

	t.Run("human readable", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunDF(&buf, []string{}, DFOptions{HumanReadable: true})
		if err != nil {
			t.Fatalf("RunDF() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Size") {
			t.Errorf("RunDF() -h should contain 'Size' header: %s", output)
		}
	})

	t.Run("inodes", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunDF(&buf, []string{}, DFOptions{Inodes: true})
		if err != nil {
			t.Fatalf("RunDF() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Inodes") {
			t.Errorf("RunDF() -i should contain 'Inodes' header: %s", output)
		}
	})

	t.Run("specific path", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunDF(&buf, []string{"."}, DFOptions{})
		if err != nil {
			t.Fatalf("RunDF() error = %v", err)
		}

		if buf.Len() == 0 {
			t.Error("RunDF() should produce output for current directory")
		}
	})

	t.Run("nonexistent path", func(t *testing.T) {
		var buf bytes.Buffer

		// Should not error, just print error message
		err := RunDF(&buf, []string{"/nonexistent/path/12345"}, DFOptions{})
		if err != nil {
			t.Fatalf("RunDF() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "df:") || !strings.Contains(output, "nonexistent") {
			t.Logf("RunDF() output for nonexistent: %s", output)
		}
	})

	t.Run("total option", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunDF(&buf, []string{".", "."}, DFOptions{Total: true})
		if err != nil {
			t.Fatalf("RunDF() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "total") {
			t.Errorf("RunDF() --total should show total line: %s", output)
		}
	})

	t.Run("block size", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunDF(&buf, []string{"."}, DFOptions{BlockSize: 4096})
		if err != nil {
			t.Fatalf("RunDF() error = %v", err)
		}

		if buf.Len() == 0 {
			t.Error("RunDF() with block size should produce output")
		}
	})
}

func TestGetDiskFree(t *testing.T) {
	info, err := GetDiskFree(".")
	if err != nil {
		t.Fatalf("GetDiskFree() error = %v", err)
	}

	if info.Size == 0 {
		t.Error("GetDiskFree() size should be > 0")
	}

	if info.MountedOn == "" {
		t.Error("GetDiskFree() MountedOn should not be empty")
	}
}

func TestPrintDFInfo(t *testing.T) {
	info := DFInfo{
		Filesystem:  "test",
		Size:        1024 * 1024 * 1024,
		Used:        512 * 1024 * 1024,
		Available:   512 * 1024 * 1024,
		UsePercent:  50,
		MountedOn:   "/test",
		Inodes:      1000,
		IUsed:       500,
		IFree:       500,
		IUsePercent: 50,
	}

	t.Run("default format", func(t *testing.T) {
		var buf bytes.Buffer
		printDFInfo(&buf, info, DFOptions{BlockSize: 1024})

		output := buf.String()
		if !strings.Contains(output, "test") {
			t.Errorf("printDFInfo() should contain filesystem name: %s", output)
		}

		if !strings.Contains(output, "/test") {
			t.Errorf("printDFInfo() should contain mount point: %s", output)
		}
	})

	t.Run("human readable format", func(t *testing.T) {
		var buf bytes.Buffer
		printDFInfo(&buf, info, DFOptions{HumanReadable: true})

		output := buf.String()
		// Should have human readable sizes like 1.0G or 512M
		if !strings.Contains(output, "G") && !strings.Contains(output, "M") {
			t.Errorf("printDFInfo() -h should have human readable sizes: %s", output)
		}
	})

	t.Run("inodes format", func(t *testing.T) {
		var buf bytes.Buffer
		printDFInfo(&buf, info, DFOptions{Inodes: true})

		output := buf.String()
		if !strings.Contains(output, "1000") {
			t.Errorf("printDFInfo() -i should show inode count: %s", output)
		}
	})
}
