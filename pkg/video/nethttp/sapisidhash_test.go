package nethttp

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"testing"
)

func TestComputeSAPISIDHash(t *testing.T) {
	t.Run("format", func(t *testing.T) {
		result := ComputeSAPISIDHash("my_sapisid", "https://www.youtube.com")

		re := regexp.MustCompile(`^SAPISIDHASH \d+_[0-9a-f]{40}$`)
		if !re.MatchString(result) {
			t.Errorf("result %q does not match expected format", result)
		}
	})

	t.Run("different inputs produce different outputs", func(t *testing.T) {
		a := ComputeSAPISIDHash("sapisid_a", "https://origin-a.com")
		b := ComputeSAPISIDHash("sapisid_b", "https://origin-b.com")
		if a == b {
			t.Error("expected different outputs for different inputs")
		}
	})
}

func TestExtractSAPISID(t *testing.T) {
	origin := "https://www.youtube.com"
	u, _ := url.Parse(origin)

	t.Run("SAPISID cookie", func(t *testing.T) {
		jar, _ := cookiejar.New(nil)
		jar.SetCookies(u, []*http.Cookie{
			{Name: "SAPISID", Value: "test_value"},
		})

		got := ExtractSAPISID(jar, origin)
		if got != "test_value" {
			t.Errorf("got %q, want %q", got, "test_value")
		}
	})

	t.Run("fallback to __Secure-3PAPISID", func(t *testing.T) {
		jar, _ := cookiejar.New(nil)
		jar.SetCookies(u, []*http.Cookie{
			{Name: "__Secure-3PAPISID", Value: "fallback_value"},
		})

		got := ExtractSAPISID(jar, origin)
		if got != "fallback_value" {
			t.Errorf("got %q, want %q", got, "fallback_value")
		}
	})

	t.Run("no matching cookies", func(t *testing.T) {
		jar, _ := cookiejar.New(nil)
		jar.SetCookies(u, []*http.Cookie{
			{Name: "OTHER", Value: "irrelevant"},
		})

		got := ExtractSAPISID(jar, origin)
		if got != "" {
			t.Errorf("got %q, want empty string", got)
		}
	})
}
