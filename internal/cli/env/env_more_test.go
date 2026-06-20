package env

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/inovacc/omni/pkg/cobra/helper/output"
)

func TestRunEnv_JSONAll(t *testing.T) {
	t.Setenv("OMNI_ENV_JSON_ALL", "jsonval")

	var buf bytes.Buffer
	if err := RunEnv(&buf, nil, EnvOptions{OutputFormat: output.FormatJSON}); err != nil {
		t.Fatalf("RunEnv json all: %v", err)
	}

	var vars []EnvVar
	if err := json.Unmarshal(buf.Bytes(), &vars); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}

	found := false
	for _, v := range vars {
		if v.Name == "OMNI_ENV_JSON_ALL" && v.Value == "jsonval" {
			found = true
		}
	}

	if !found {
		t.Error("expected OMNI_ENV_JSON_ALL in JSON output")
	}
}

func TestRunEnv_JSONFiltered(t *testing.T) {
	t.Setenv("OMNI_ENV_JSON_FILTER", "filtered")

	var buf bytes.Buffer
	if err := RunEnv(&buf, []string{"OMNI_ENV_JSON_FILTER", "DEFINITELY_NOT_SET_12345"}, EnvOptions{OutputFormat: output.FormatJSON}); err != nil {
		t.Fatalf("RunEnv json filtered: %v", err)
	}

	var vars []EnvVar
	if err := json.Unmarshal(buf.Bytes(), &vars); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}

	if len(vars) != 1 {
		t.Fatalf("expected exactly 1 var (unset skipped), got %d: %+v", len(vars), vars)
	}

	if vars[0].Name != "OMNI_ENV_JSON_FILTER" || vars[0].Value != "filtered" {
		t.Errorf("got %+v", vars[0])
	}
}

func TestRunEnv_NullTerminated(t *testing.T) {
	t.Setenv("OMNI_ENV_NUL", "nulval")

	t.Run("filtered", func(t *testing.T) {
		var buf bytes.Buffer
		if err := RunEnv(&buf, []string{"OMNI_ENV_NUL"}, EnvOptions{NullTerminated: true}); err != nil {
			t.Fatalf("RunEnv: %v", err)
		}

		out := buf.String()
		if !strings.HasSuffix(out, "\x00") {
			t.Errorf("expected NUL terminator, got %q", out)
		}

		if strings.Contains(out, "\n") {
			t.Errorf("did not expect newline with -0, got %q", out)
		}
	})

	t.Run("all", func(t *testing.T) {
		var buf bytes.Buffer
		if err := RunEnv(&buf, nil, EnvOptions{NullTerminated: true}); err != nil {
			t.Fatalf("RunEnv: %v", err)
		}

		if !strings.Contains(buf.String(), "\x00") {
			t.Error("expected NUL bytes in full listing with -0")
		}
	})
}

func TestRunEnv_Unset(t *testing.T) {
	t.Setenv("OMNI_ENV_KEEP", "keep")
	t.Setenv("OMNI_ENV_DROP", "drop")

	var buf bytes.Buffer
	if err := RunEnv(&buf, nil, EnvOptions{Unset: "OMNI_ENV_DROP"}); err != nil {
		t.Fatalf("RunEnv: %v", err)
	}

	out := buf.String()
	if strings.Contains(out, "OMNI_ENV_DROP=") {
		t.Error("expected OMNI_ENV_DROP to be filtered out")
	}

	if !strings.Contains(out, "OMNI_ENV_KEEP=keep") {
		t.Error("expected OMNI_ENV_KEEP to remain")
	}
}

func TestRunEnv_Ignore(t *testing.T) {
	t.Setenv("OMNI_ENV_IGNORED", "should_not_show")

	var buf bytes.Buffer
	if err := RunEnv(&buf, nil, EnvOptions{Ignore: true}); err != nil {
		t.Fatalf("RunEnv: %v", err)
	}

	// With -i and no args, the environment is emptied -> empty output.
	if strings.Contains(buf.String(), "OMNI_ENV_IGNORED") {
		t.Errorf("expected empty environment with -i, got %q", buf.String())
	}
}

func TestRunEnv_JSONWriteFailure(t *testing.T) {
	t.Setenv("OMNI_ENV_JSON_FAIL", "x")

	t.Run("all", func(t *testing.T) {
		if err := RunEnv(failingWriter{}, nil, EnvOptions{OutputFormat: output.FormatJSON}); err == nil {
			t.Error("expected error writing JSON to failing writer")
		}
	})

	t.Run("filtered", func(t *testing.T) {
		if err := RunEnv(failingWriter{}, []string{"OMNI_ENV_JSON_FAIL"}, EnvOptions{OutputFormat: output.FormatJSON}); err == nil {
			t.Error("expected error writing filtered JSON to failing writer")
		}
	})
}

func TestEnvHelpers(t *testing.T) {
	t.Setenv("OMNI_ENV_HELPER", "helperval")

	if got := GetEnv("OMNI_ENV_HELPER"); got != "helperval" {
		t.Errorf("GetEnv = %q, want helperval", got)
	}

	if got := GetEnv("OMNI_ENV_DOES_NOT_EXIST_999"); got != "" {
		t.Errorf("GetEnv for unset = %q, want empty", got)
	}

	if v, ok := LookupEnv("OMNI_ENV_HELPER"); !ok || v != "helperval" {
		t.Errorf("LookupEnv = (%q,%v), want (helperval,true)", v, ok)
	}

	if _, ok := LookupEnv("OMNI_ENV_DOES_NOT_EXIST_999"); ok {
		t.Error("LookupEnv for unset should be false")
	}

	all := Environ()
	if all["OMNI_ENV_HELPER"] != "helperval" {
		t.Errorf("Environ()[OMNI_ENV_HELPER] = %q, want helperval", all["OMNI_ENV_HELPER"])
	}
}
