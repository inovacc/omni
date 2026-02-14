package nethttp

import (
	"crypto/sha1"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// ComputeSAPISIDHash computes the SAPISIDHASH value required for authenticated
// YouTube WEB client InnerTube requests.
//
// Format: SAPISIDHASH <unix_timestamp>_<sha1(timestamp + " " + sapisid + " " + origin)>
func ComputeSAPISIDHash(sapisid, origin string) string {
	ts := time.Now().Unix()
	input := fmt.Sprintf("%d %s %s", ts, sapisid, origin)
	hash := sha1.Sum([]byte(input))

	return fmt.Sprintf("SAPISIDHASH %d_%x", ts, hash)
}

// ExtractSAPISID extracts the SAPISID (or __Secure-3PAPISID) cookie value
// from the cookie jar for the given origin URL.
// Returns empty string if neither cookie is found.
func ExtractSAPISID(jar http.CookieJar, origin string) string {
	u, err := url.Parse(origin)
	if err != nil {
		return ""
	}

	cookies := jar.Cookies(u)
	for _, c := range cookies {
		if c.Name == "SAPISID" {
			return c.Value
		}
	}

	// Fallback to __Secure-3PAPISID (newer cookie name).
	for _, c := range cookies {
		if c.Name == "__Secure-3PAPISID" {
			return c.Value
		}
	}

	return ""
}
