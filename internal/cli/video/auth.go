package video

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/inovacc/scout/pkg/scout"

	"github.com/inovacc/omni/pkg/video/nethttp"
)

// RunAuth launches Chrome with a copy of the user's profile, navigates to YouTube,
// extracts cookies via CDP, and saves them as a Netscape cookies.txt file at the
// well-known cache path for subsequent authenticated video operations.
func RunAuth(w io.Writer, _ []string, _ Options) error {
	_, _ = fmt.Fprintln(w, "[auth] Launching Chrome to extract YouTube cookies...")

	userDataDir := chromeUserDataDir()
	if userDataDir == "" {
		return fmt.Errorf("video auth: could not locate Chrome user data directory")
	}

	if _, err := os.Stat(userDataDir); os.IsNotExist(err) {
		return fmt.Errorf("video auth: Chrome user data directory not found: %s", userDataDir)
	}

	_, _ = fmt.Fprintf(w, "[auth] Using Chrome profile: %s\n", userDataDir)

	// Copy essential profile files to a temp dir so we don't conflict
	// with a running Chrome instance that holds a lock on the profile.
	tmpDir, err := copyProfileToTemp(userDataDir)
	if err != nil {
		return fmt.Errorf("video auth: copy profile: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	_, _ = fmt.Fprintf(w, "[auth] Copied profile to temp: %s\n", tmpDir)

	scoutOpts := []scout.Option{
		scout.WithHeadless(true),
		scout.WithUserDataDir(tmpDir),
		scout.WithNoSandbox(),
		// Prevent Chrome from restoring previous session tabs.
		scout.WithLaunchFlag("disable-session-crashed-bubble", ""),
		scout.WithLaunchFlag("disable-infobars", ""),
		scout.WithLaunchFlag("no-first-run", ""),
		scout.WithLaunchFlag("no-default-browser-check", ""),
		scout.WithLaunchFlag("restore-last-session", "false"),
	}

	// Explicitly set Chrome path if we can find it.
	if chromePath := findChrome(); chromePath != "" {
		scoutOpts = append(scoutOpts, scout.WithExecPath(chromePath))
	}

	browser, err := scout.New(scoutOpts...)
	if err != nil {
		return fmt.Errorf("video auth: launch browser: %w", err)
	}
	defer func() { _ = browser.Close() }()

	// Open a blank page first, then navigate.
	// This avoids conflicts with Chrome's startup/session restore.
	page, err := browser.NewPage("about:blank")
	if err != nil {
		return fmt.Errorf("video auth: create page: %w", err)
	}
	defer func() { _ = page.Close() }()

	if err := page.Navigate("https://www.youtube.com"); err != nil {
		return fmt.Errorf("video auth: navigate: %w", err)
	}

	if err := page.WaitLoad(); err != nil {
		return fmt.Errorf("video auth: wait load: %w", err)
	}

	cookies, err := page.GetCookies()
	if err != nil {
		return fmt.Errorf("video auth: get cookies: %w", err)
	}

	// Filter to YouTube and Google domain cookies.
	var filtered []*http.Cookie
	for _, c := range cookies {
		domain := c.Domain
		if isYouTubeDomain(domain) || isGoogleDomain(domain) {
			filtered = append(filtered, &http.Cookie{
				Name:     c.Name,
				Value:    c.Value,
				Domain:   c.Domain,
				Path:     c.Path,
				Expires:  c.Expires,
				Secure:   c.Secure,
				HttpOnly: c.HTTPOnly,
			})
		}
	}

	if len(filtered) == 0 {
		return fmt.Errorf("video auth: no YouTube/Google cookies found; are you logged in to YouTube in Chrome?")
	}

	// Write to well-known path.
	cookiePath := nethttp.DefaultCookiePath()
	if err := nethttp.WriteNetscapeCookies(cookiePath, filtered); err != nil {
		return fmt.Errorf("video auth: write cookies: %w", err)
	}

	_, _ = fmt.Fprintf(w, "[auth] Saved %d cookies to %s\n", len(filtered), cookiePath)

	// Check for SAPISID presence.
	hasSAPISID := false
	for _, c := range filtered {
		if c.Name == "SAPISID" || c.Name == "__Secure-3PAPISID" {
			hasSAPISID = true
			break
		}
	}

	if hasSAPISID {
		_, _ = fmt.Fprintln(w, "[auth] SAPISID cookie found - authenticated requests enabled")
	} else {
		_, _ = fmt.Fprintln(w, "[auth] Warning: SAPISID cookie not found - you may not be logged in")
	}

	_, _ = fmt.Fprintln(w, "[auth] Done. Cookies will be auto-loaded for future video commands.")

	return nil
}

// copyProfileToTemp copies the essential Chrome profile files to a temporary
// directory so we can launch a headless Chrome without conflicting with a
// running instance that holds a lock on the original profile.
func copyProfileToTemp(userDataDir string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "omni-video-auth-*")
	if err != nil {
		return "", fmt.Errorf("create temp dir: %w", err)
	}

	// Files to copy from the user data root.
	rootFiles := []string{"Local State"}

	// Files to copy from the Default profile subfolder.
	// Only copy cookie-related files; skip Preferences to avoid session restore.
	profileFiles := []string{
		"Cookies",
		"Cookies-journal",
	}

	for _, name := range rootFiles {
		src := filepath.Join(userDataDir, name)
		dst := filepath.Join(tmpDir, name)

		if err := copyFileIfExists(src, dst); err != nil {
			_ = os.RemoveAll(tmpDir)
			return "", err
		}
	}

	profileDir := filepath.Join(tmpDir, "Default")
	if err := os.MkdirAll(profileDir, 0o700); err != nil {
		_ = os.RemoveAll(tmpDir)
		return "", err
	}

	for _, name := range profileFiles {
		src := filepath.Join(userDataDir, "Default", name)
		dst := filepath.Join(profileDir, name)

		if err := copyFileIfExists(src, dst); err != nil {
			_ = os.RemoveAll(tmpDir)
			return "", err
		}
	}

	return tmpDir, nil
}

func copyFileIfExists(src, dst string) error {
	data, err := os.ReadFile(src)
	if os.IsNotExist(err) {
		return nil // Skip missing files.
	}

	if err != nil {
		return fmt.Errorf("read %s: %w", filepath.Base(src), err)
	}

	return os.WriteFile(dst, data, 0o600)
}

// findChrome returns the path to the Chrome executable, or empty string if not found.
func findChrome() string {
	switch runtime.GOOS {
	case "windows":
		candidates := []string{
			filepath.Join(os.Getenv("PROGRAMFILES"), "Google", "Chrome", "Application", "chrome.exe"),
			filepath.Join(os.Getenv("PROGRAMFILES(X86)"), "Google", "Chrome", "Application", "chrome.exe"),
			filepath.Join(os.Getenv("LOCALAPPDATA"), "Google", "Chrome", "Application", "chrome.exe"),
		}

		for _, p := range candidates {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	case "darwin":
		p := "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
		if _, err := os.Stat(p); err == nil {
			return p
		}
	default: // linux
		for _, name := range []string{"google-chrome-stable", "google-chrome", "chromium-browser", "chromium"} {
			// Check common paths.
			for _, dir := range []string{"/usr/bin", "/usr/local/bin", "/snap/bin"} {
				p := filepath.Join(dir, name)
				if _, err := os.Stat(p); err == nil {
					return p
				}
			}
		}
	}

	return ""
}

// chromeUserDataDir returns the default Chrome user data directory for the current OS.
func chromeUserDataDir() string {
	switch runtime.GOOS {
	case "windows":
		if dir := os.Getenv("LOCALAPPDATA"); dir != "" {
			return filepath.Join(dir, "Google", "Chrome", "User Data")
		}
	case "darwin":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Application Support", "Google", "Chrome")
	default: // linux
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".config", "google-chrome")
	}

	return ""
}

func isYouTubeDomain(domain string) bool {
	d := strings.TrimPrefix(domain, ".")
	return d == "youtube.com" || strings.HasSuffix(d, ".youtube.com")
}

func isGoogleDomain(domain string) bool {
	d := strings.TrimPrefix(domain, ".")
	return d == "google.com" || strings.HasSuffix(d, ".google.com")
}
