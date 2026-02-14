package nethttp

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// LoadNetscapeCookies reads a Netscape/Mozilla format cookies.txt file.
// Format: domain  flag  path  secure  expiration  name  value
func LoadNetscapeCookies(path string) ([]*http.Cookie, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cookies: %w", err)
	}

	defer func() { _ = f.Close() }()

	var cookies []*http.Cookie

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines.
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) < 7 {
			continue
		}

		domain := fields[0]
		path := fields[2]
		secure := strings.EqualFold(fields[3], "TRUE")
		name := fields[5]
		value := fields[6]

		var expires time.Time
		if exp, err := strconv.ParseInt(fields[4], 10, 64); err == nil && exp > 0 {
			expires = time.Unix(exp, 0)
		}

		c := &http.Cookie{
			Name:     name,
			Value:    value,
			Domain:   domain,
			Path:     path,
			Secure:   secure,
			HttpOnly: false,
		}
		if !expires.IsZero() {
			c.Expires = expires
		}

		cookies = append(cookies, c)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("cookies: reading file: %w", err)
	}

	return cookies, nil
}

// DefaultCookiePath returns the well-known path for extracted YouTube cookies.
//   - Windows: %LOCALAPPDATA%\omni-video\cookies\youtube.txt
//   - Linux/macOS: ~/.cache/omni-video/cookies/youtube.txt
func DefaultCookiePath() string {
	return filepath.Join(defaultCookieDir(), "youtube.txt")
}

// AutoLoadCookies loads cookies from the default path if the file exists.
// Returns nil, nil if the file does not exist.
func AutoLoadCookies() ([]*http.Cookie, error) {
	path := DefaultCookiePath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	}

	return LoadNetscapeCookies(path)
}

func defaultCookieDir() string {
	if dir := os.Getenv("XDG_CACHE_HOME"); dir != "" {
		return filepath.Join(dir, "omni-video", "cookies")
	}

	switch runtime.GOOS {
	case "windows":
		if dir := os.Getenv("LOCALAPPDATA"); dir != "" {
			return filepath.Join(dir, "omni-video", "cookies")
		}

		home, _ := os.UserHomeDir()

		return filepath.Join(home, "AppData", "Local", "omni-video", "cookies")
	default:
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".cache", "omni-video", "cookies")
	}
}
