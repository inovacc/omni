package video

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/inovacc/scout/pkg/scout"
)

func TestScoutBasicLaunch(t *testing.T) {
	if os.Getenv("TEST_SCOUT") == "" {
		t.Skip("set TEST_SCOUT=1 to run browser tests")
	}

	chromePath := findChrome()
	t.Logf("Chrome path: %s", chromePath)

	opts := []scout.Option{
		scout.WithHeadless(true),
		scout.WithNoSandbox(),
	}
	if chromePath != "" {
		opts = append(opts, scout.WithExecPath(chromePath))
	}

	b, err := scout.New(opts...)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}
	defer func() { _ = b.Close() }()

	page, err := b.NewPage("about:blank")
	if err != nil {
		t.Fatalf("new page: %v", err)
	}
	defer func() { _ = page.Close() }()

	t.Log("Basic launch succeeded")

	if err := page.Navigate("https://www.youtube.com"); err != nil {
		t.Fatalf("navigate: %v", err)
	}

	if err := page.WaitLoad(); err != nil {
		t.Fatalf("wait load: %v", err)
	}

	cookies, err := page.GetCookies()
	if err != nil {
		t.Fatalf("get cookies: %v", err)
	}

	t.Logf("Got %d cookies (no profile)", len(cookies))
	for _, c := range cookies {
		val := c.Value
		if len(val) > 30 {
			val = val[:30] + "..."
		}
		t.Logf("  %s = %s (domain: %s)", c.Name, val, c.Domain)
	}
}

func TestScoutWithProfile(t *testing.T) {
	if os.Getenv("TEST_SCOUT") == "" {
		t.Skip("set TEST_SCOUT=1 to run browser tests")
	}

	userDataDir := chromeUserDataDir()
	if userDataDir == "" {
		t.Skip("Chrome user data dir not found")
	}

	tmpDir, err := copyProfileToTemp(userDataDir)
	if err != nil {
		t.Fatalf("copy profile: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Logf("Temp dir: %s", tmpDir)

	// List what we copied
	entries, _ := os.ReadDir(tmpDir)
	for _, e := range entries {
		t.Logf("  root: %s", e.Name())
	}
	defEntries, _ := os.ReadDir(fmt.Sprintf("%s/Default", tmpDir))
	for _, e := range defEntries {
		t.Logf("  Default/: %s", e.Name())
	}

	chromePath := findChrome()
	opts := []scout.Option{
		scout.WithHeadless(true),
		scout.WithUserDataDir(tmpDir),
		scout.WithNoSandbox(),
		scout.WithLaunchFlag("disable-session-crashed-bubble", ""),
		scout.WithLaunchFlag("no-first-run", ""),
	}
	if chromePath != "" {
		opts = append(opts, scout.WithExecPath(chromePath))
	}

	b, err := scout.New(opts...)
	if err != nil {
		t.Fatalf("launch with profile: %v", err)
	}
	defer func() { _ = b.Close() }()

	// List all pages after startup
	pages, _ := b.Pages()
	t.Logf("Pages after startup: %d", len(pages))

	page, err := b.NewPage("about:blank")
	if err != nil {
		t.Fatalf("new page: %v", err)
	}
	defer func() { _ = page.Close() }()

	t.Log("Page created, navigating...")

	if err := page.Navigate("https://www.youtube.com"); err != nil {
		t.Fatalf("navigate: %v", err)
	}

	t.Log("Navigated, waiting for load...")

	if err := page.WaitLoad(); err != nil {
		t.Fatalf("wait load: %v", err)
	}

	cookies, err := page.GetCookies()
	if err != nil {
		t.Fatalf("get cookies: %v", err)
	}

	t.Logf("Got %d cookies (with profile)", len(cookies))
	for _, c := range cookies {
		val := c.Value
		if len(val) > 30 {
			val = val[:30] + "..."
		}
		t.Logf("  %s = %s (domain: %s)", c.Name, val, c.Domain)
	}
}

func TestRunAuthIntegration(t *testing.T) {
	if os.Getenv("TEST_SCOUT") == "" {
		t.Skip("set TEST_SCOUT=1 to run browser tests")
	}

	var buf bytes.Buffer
	err := RunAuth(&buf, nil, Options{})
	t.Log(buf.String())
	if err != nil {
		t.Fatalf("RunAuth: %v", err)
	}
}
