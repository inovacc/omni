package video

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/inovacc/omni/pkg/video/nethttp"
)

// RunAuth extracts YouTube cookies from Chrome by launching a headless
// instance with remote debugging and using CDP to get decrypted cookies.
//
// It first tries the real profile (requires Chrome to be closed).
// If that fails, it falls back to a profile copy (may not have login cookies
// due to Chrome's App-Bound Encryption in v127+).
func RunAuth(w io.Writer, _ []string, _ Options) error {
	_, _ = fmt.Fprintln(w, "[auth] Extracting YouTube cookies from Chrome...")

	chromePath := findChrome()
	if chromePath == "" {
		return fmt.Errorf("video auth: Chrome not found")
	}

	userDataDir := chromeUserDataDir()
	if userDataDir == "" {
		return fmt.Errorf("video auth: could not locate Chrome user data directory")
	}

	_, _ = fmt.Fprintf(w, "[auth] Chrome: %s\n", chromePath)
	_, _ = fmt.Fprintf(w, "[auth] Profile: %s\n", userDataDir)

	// Find a free port for debugging.
	port, err := freePort()
	if err != nil {
		return fmt.Errorf("video auth: find free port: %w", err)
	}

	// Try real profile first (requires Chrome closed).
	profileDir := userDataDir
	usingCopy := false

	_, _ = fmt.Fprintf(w, "[auth] Starting headless Chrome on port %d...\n", port)

	cmd := launchHeadlessChrome(chromePath, profileDir, port)
	if err := cmd.Start(); err != nil {
		// Real profile locked — fall back to copy.
		_, _ = fmt.Fprintln(w, "[auth] Chrome profile locked, using copy (login cookies may be unavailable)...")
		tmpDir, copyErr := copyProfileToTemp(userDataDir)
		if copyErr != nil {
			return fmt.Errorf("video auth: copy profile: %w", copyErr)
		}
		defer func() { _ = os.RemoveAll(tmpDir) }()
		profileDir = tmpDir
		usingCopy = true

		cmd = launchHeadlessChrome(chromePath, profileDir, port)
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("video auth: start Chrome: %w", err)
		}
	}

	// Ensure Chrome is killed on exit.
	defer func() {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	}()

	// Wait for debugging port to become available.
	debugURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	_, _ = fmt.Fprintln(w, "[auth] Waiting for Chrome to start...")

	if err := waitForDebugPort(debugURL, 10*time.Second); err != nil {
		// Real profile failed (likely locked) — kill and retry with copy.
		if !usingCopy {
			_ = cmd.Process.Kill()
			_, _ = cmd.Process.Wait()

			_, _ = fmt.Fprintln(w, "[auth] Real profile unavailable, using copy (close Chrome for full login cookies)...")
			tmpDir, copyErr := copyProfileToTemp(userDataDir)
			if copyErr != nil {
				return fmt.Errorf("video auth: copy profile: %w", copyErr)
			}
			defer func() { _ = os.RemoveAll(tmpDir) }()
			usingCopy = true

			port2, _ := freePort()
			cmd = launchHeadlessChrome(chromePath, tmpDir, port2)
			if startErr := cmd.Start(); startErr != nil {
				return fmt.Errorf("video auth: start Chrome with copy: %w", startErr)
			}
			debugURL = fmt.Sprintf("http://127.0.0.1:%d", port2)

			if err := waitForDebugPort(debugURL, 15*time.Second); err != nil {
				return fmt.Errorf("video auth: Chrome did not start: %w", err)
			}
			// Give YouTube page time to load and set cookies.
			time.Sleep(3 * time.Second)
		} else {
			return fmt.Errorf("video auth: Chrome did not start: %w", err)
		}
	} else {
		// Give YouTube page time to load and set cookies.
		time.Sleep(3 * time.Second)
	}

	// Get cookies via CDP.
	_, _ = fmt.Fprintln(w, "[auth] Extracting cookies via Chrome DevTools Protocol...")

	cookies, err := getCDPCookies(debugURL)
	if err != nil {
		return fmt.Errorf("video auth: get cookies: %w", err)
	}

	// Filter to YouTube and Google domains.
	var filtered []*http.Cookie
	for _, c := range cookies {
		if isYouTubeDomain(c.Domain) || isGoogleDomain(c.Domain) {
			filtered = append(filtered, c)
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
		_, _ = fmt.Fprintln(w, "[auth] Warning: SAPISID cookie not found - login cookies could not be extracted")
		if usingCopy {
			_, _ = fmt.Fprintln(w, "[auth] Tip: close Chrome completely, then run 'omni video auth' again for full login cookies")
		}
	}

	_, _ = fmt.Fprintln(w, "[auth] Done. Use --cookies-from-browser to auto-load these cookies.")

	return nil
}

// launchHeadlessChrome creates an exec.Cmd for headless Chrome with remote debugging.
func launchHeadlessChrome(chromePath, profileDir string, port int) *exec.Cmd {
	cmd := exec.Command(chromePath,
		fmt.Sprintf("--remote-debugging-port=%d", port),
		"--headless=new",
		fmt.Sprintf("--user-data-dir=%s", profileDir),
		"--no-first-run",
		"--no-default-browser-check",
		"--disable-extensions",
		"--disable-background-networking",
		"--disable-sync",
		"--disable-translate",
		"--disable-features=TranslateUI,MediaRouter",
		"https://www.youtube.com",
	)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd
}

// freePort finds an available TCP port.
func freePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	port := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()
	return port, nil
}

// waitForDebugPort polls the Chrome debug endpoint until it responds.
func waitForDebugPort(debugURL string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 2 * time.Second}

	for time.Now().Before(deadline) {
		resp, err := client.Get(debugURL + "/json/version")
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == 200 {
				return nil
			}
		}
		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for %s", debugURL)
}

// cdpCookie is the JSON shape returned by Chrome's Network.getAllCookies.
type cdpCookie struct {
	Name     string  `json:"name"`
	Value    string  `json:"value"`
	Domain   string  `json:"domain"`
	Path     string  `json:"path"`
	Expires  float64 `json:"expires"`
	Secure   bool    `json:"secure"`
	HTTPOnly bool    `json:"httpOnly"`
}

// getCDPCookies retrieves all cookies from Chrome via the CDP HTTP endpoint.
func getCDPCookies(debugURL string) ([]*http.Cookie, error) {
	// First get the WebSocket debugger URL for a page.
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(debugURL + "/json/list")
	if err != nil {
		return nil, fmt.Errorf("list pages: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var pages []struct {
		ID                   string `json:"id"`
		WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&pages); err != nil {
		return nil, fmt.Errorf("decode pages: %w", err)
	}

	if len(pages) == 0 {
		return nil, fmt.Errorf("no pages found")
	}

	// Use the CDP HTTP endpoint to send a command.
	// Chrome's /json/protocol doesn't support direct command execution,
	// but we can use the WebSocket. For simplicity, use the /json endpoint
	// to get cookies via a different approach.

	// Alternative: use Chrome's /json/new to create a page and execute via ws.
	// Simpler: use the send_command approach via ws.

	// Actually, the simplest approach is to use the WebSocket.
	wsURL := pages[0].WebSocketDebuggerURL
	if wsURL == "" {
		return nil, fmt.Errorf("no WebSocket URL for page")
	}

	return getCookiesViaWebSocket(wsURL)
}

// getCookiesViaWebSocket connects to Chrome's WebSocket and calls Network.getAllCookies.
func getCookiesViaWebSocket(wsURL string) ([]*http.Cookie, error) {
	// Use a raw WebSocket connection via net.Dial + HTTP upgrade.
	// Parse the wsURL to get host and path.
	// wsURL format: ws://127.0.0.1:PORT/devtools/page/ID
	wsURL = strings.Replace(wsURL, "ws://", "", 1)
	parts := strings.SplitN(wsURL, "/", 2)
	host := parts[0]
	path := "/"
	if len(parts) > 1 {
		path = "/" + parts[1]
	}

	conn, err := net.DialTimeout("tcp", host, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}
	defer func() { _ = conn.Close() }()

	// Send WebSocket upgrade request.
	upgrade := fmt.Sprintf(
		"GET %s HTTP/1.1\r\nHost: %s\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\nSec-WebSocket-Version: 13\r\n\r\n",
		path, host,
	)
	if _, err := conn.Write([]byte(upgrade)); err != nil {
		return nil, fmt.Errorf("upgrade: %w", err)
	}

	// Read upgrade response.
	buf := make([]byte, 4096)
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	n, err := conn.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("read upgrade: %w", err)
	}
	if !strings.Contains(string(buf[:n]), "101") {
		return nil, fmt.Errorf("WebSocket upgrade failed: %s", string(buf[:n]))
	}

	// Send Network.getAllCookies command.
	cmdJSON := `{"id":1,"method":"Network.getAllCookies"}`
	frame := wsEncodeTextFrame([]byte(cmdJSON))
	if _, err := conn.Write(frame); err != nil {
		return nil, fmt.Errorf("send command: %w", err)
	}

	// Read response (may come in multiple reads).
	_ = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	var responseData []byte
	for {
		readBuf := make([]byte, 65536)
		rn, readErr := conn.Read(readBuf)
		if rn > 0 {
			responseData = append(responseData, readBuf[:rn]...)
			// Try to extract the JSON response from WebSocket frames.
			if payload := wsExtractPayload(responseData); payload != nil {
				var result struct {
					ID     int `json:"id"`
					Result struct {
						Cookies []cdpCookie `json:"cookies"`
					} `json:"result"`
				}
				if err := json.Unmarshal(payload, &result); err == nil && result.ID == 1 {
					// Convert to http.Cookie.
					var cookies []*http.Cookie
					for _, c := range result.Result.Cookies {
						hc := &http.Cookie{
							Name:     c.Name,
							Value:    c.Value,
							Domain:   c.Domain,
							Path:     c.Path,
							Secure:   c.Secure,
							HttpOnly: c.HTTPOnly,
						}
						if c.Expires > 0 {
							hc.Expires = time.Unix(int64(c.Expires), 0)
						}
						cookies = append(cookies, hc)
					}
					return cookies, nil
				}
			}
		}
		if readErr != nil {
			return nil, fmt.Errorf("read response: %w (got %d bytes)", readErr, len(responseData))
		}
	}
}

// wsEncodeTextFrame creates a masked WebSocket text frame.
func wsEncodeTextFrame(payload []byte) []byte {
	length := len(payload)
	var frame []byte

	// Opcode 0x81 = final frame + text.
	frame = append(frame, 0x81)

	// Length + mask bit (client must mask).
	if length < 126 {
		frame = append(frame, byte(length)|0x80)
	} else if length < 65536 {
		frame = append(frame, 126|0x80, byte(length>>8), byte(length))
	} else {
		frame = append(frame, 127|0x80)
		for i := 7; i >= 0; i-- {
			frame = append(frame, byte(length>>(i*8)))
		}
	}

	// Mask key (all zeros for simplicity — fine for localhost).
	mask := []byte{0, 0, 0, 0}
	frame = append(frame, mask...)

	// Masked payload (XOR with zeros = plaintext).
	frame = append(frame, payload...)

	return frame
}

// maxWSFrame bounds the payload length we will accept from a WebSocket frame.
// A 64-bit length with the high bit set decodes to a negative Go int, and even
// a large positive length could trigger an unbounded slice/allocation. We only
// ever read CDP responses from a localhost debugger, so 16 MiB is generous.
const maxWSFrame = 16 << 20

// wsExtractPayload extracts the payload from a WebSocket frame.
func wsExtractPayload(data []byte) []byte {
	if len(data) < 2 {
		return nil
	}

	payloadLen := int(data[1] & 0x7F)
	masked := data[1]&0x80 != 0
	offset := 2

	if payloadLen == 126 {
		if len(data) < 4 {
			return nil
		}
		payloadLen = int(data[2])<<8 | int(data[3])
		offset = 4
	} else if payloadLen == 127 {
		if len(data) < 10 {
			return nil
		}
		payloadLen = 0
		for i := 0; i < 8; i++ {
			payloadLen = payloadLen<<8 | int(data[2+i])
		}
		offset = 10
	}

	// Reject lengths that are negative (high bit set in a 64-bit length) or
	// larger than we are willing to buffer. This must happen before any
	// slicing or allocation below to avoid a slice-bounds panic.
	if payloadLen < 0 || payloadLen > maxWSFrame {
		return nil
	}

	if masked {
		offset += 4
	}

	// Guard against offset+payloadLen overflowing past the buffer.
	if offset+payloadLen < offset || len(data) < offset+payloadLen {
		return nil
	}

	payload := data[offset : offset+payloadLen]
	if masked {
		maskKey := data[offset-4 : offset]
		unmasked := make([]byte, payloadLen)
		for i := range payload {
			unmasked[i] = payload[i] ^ maskKey[i%4]
		}
		return unmasked
	}

	return payload
}

// copyProfileToTemp copies essential Chrome profile files to a temp directory
// so we can launch a headless Chrome without conflicting with a running instance.
func copyProfileToTemp(userDataDir string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "omni-video-auth-*")
	if err != nil {
		return "", fmt.Errorf("create temp dir: %w", err)
	}

	// Copy Local State (contains encryption keys).
	if err := copyFileIfExists(
		filepath.Join(userDataDir, "Local State"),
		filepath.Join(tmpDir, "Local State"),
	); err != nil {
		_ = os.RemoveAll(tmpDir)
		return "", err
	}

	// Create Default profile dir.
	profileDir := filepath.Join(tmpDir, "Default")
	if err := os.MkdirAll(profileDir, 0o700); err != nil {
		_ = os.RemoveAll(tmpDir)
		return "", err
	}

	// Copy cookie files (old and new locations).
	networkDir := filepath.Join(profileDir, "Network")
	_ = os.MkdirAll(networkDir, 0o700)

	for _, name := range []string{"Cookies", "Cookies-journal"} {
		// New location: Default/Network/
		_ = copyFileIfExists(
			filepath.Join(userDataDir, "Default", "Network", name),
			filepath.Join(networkDir, name),
		)
		// Old location: Default/
		_ = copyFileIfExists(
			filepath.Join(userDataDir, "Default", name),
			filepath.Join(profileDir, name),
		)
	}

	// Write minimal Preferences to prevent crash recovery.
	prefs := `{"profile":{"exit_type":"Normal","exited_cleanly":true},"session":{"restore_on_startup":5}}`
	if err := os.WriteFile(filepath.Join(profileDir, "Preferences"), []byte(prefs), 0o600); err != nil {
		_ = os.RemoveAll(tmpDir)
		return "", fmt.Errorf("write preferences: %w", err)
	}

	return tmpDir, nil
}

func copyFileIfExists(src, dst string) error {
	data, err := os.ReadFile(src)
	if os.IsNotExist(err) {
		return nil
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
