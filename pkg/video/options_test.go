package video

import (
	"testing"
)

func TestApplyOptions(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		o := applyOptions(nil)
		if o.Format != "best" {
			t.Errorf("Format = %q, want %q", o.Format, "best")
		}
		if o.Retries != 3 {
			t.Errorf("Retries = %d, want %d", o.Retries, 3)
		}
	})

	t.Run("single option", func(t *testing.T) {
		tests := []struct {
			name  string
			opt   Option
			check func(Options) bool
		}{
			{"WithFormat", WithFormat("worst"), func(o Options) bool { return o.Format == "worst" }},
			{"WithOutput", WithOutput("out.mp4"), func(o Options) bool { return o.Output == "out.mp4" }},
			{"WithQuiet", WithQuiet(), func(o Options) bool { return o.Quiet }},
			{"WithNoProgress", WithNoProgress(), func(o Options) bool { return o.NoProgress }},
			{"WithRateLimit", WithRateLimit(1024), func(o Options) bool { return o.RateLimit == 1024 }},
			{"WithRetries", WithRetries(5), func(o Options) bool { return o.Retries == 5 }},
			{"WithContinue", WithContinue(), func(o Options) bool { return o.Continue }},
			{"WithNoPart", WithNoPart(), func(o Options) bool { return o.NoPart }},
			{"WithCookieFile", WithCookieFile("/tmp/c.txt"), func(o Options) bool { return o.CookieFile == "/tmp/c.txt" }},
			{"WithProxy", WithProxy("socks5://localhost:1080"), func(o Options) bool { return o.Proxy == "socks5://localhost:1080" }},
			{"WithWriteInfo", WithWriteInfo(), func(o Options) bool { return o.WriteInfo }},
			{"WithWriteMarkdown", WithWriteMarkdown(), func(o Options) bool { return o.WriteMarkdown }},
			{"WithWriteSubs", WithWriteSubs(), func(o Options) bool { return o.WriteSubs }},
			{"WithNoPlaylist", WithNoPlaylist(), func(o Options) bool { return o.NoPlaylist }},
			{"WithVerbose", WithVerbose(), func(o Options) bool { return o.Verbose }},
			{"WithCacheDir", WithCacheDir("/tmp/cache"), func(o Options) bool { return o.CacheDir == "/tmp/cache" }},
			{"WithCookiesFromBrowser", WithCookiesFromBrowser(), func(o Options) bool { return o.CookiesFromBrowser }},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				o := applyOptions([]Option{tt.opt})
				if !tt.check(o) {
					t.Errorf("%s did not set expected value", tt.name)
				}
			})
		}
	})

	t.Run("multiple options compose", func(t *testing.T) {
		o := applyOptions([]Option{
			WithFormat("worst"),
			WithRetries(10),
			WithQuiet(),
		})
		if o.Format != "worst" {
			t.Errorf("Format = %q, want %q", o.Format, "worst")
		}
		if o.Retries != 10 {
			t.Errorf("Retries = %d, want %d", o.Retries, 10)
		}
		if !o.Quiet {
			t.Error("expected Quiet=true")
		}
	})
}
