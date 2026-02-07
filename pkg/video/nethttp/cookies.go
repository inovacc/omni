package nethttp

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
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
