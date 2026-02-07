package video

// Options configures the video download client.
type Options struct {
	Format        string            // Format selector (e.g., "best", "worst", "bestvideo")
	Output        string            // Output filename template
	Quiet         bool              // Suppress progress output
	NoProgress    bool              // Disable progress bar
	RateLimit     int64             // Rate limit in bytes/sec
	Retries       int               // Number of download retries
	Continue      bool              // Resume partial downloads
	NoPart        bool              // Don't use .part files
	CookieFile    string            // Path to cookies.txt
	Proxy         string            // HTTP/SOCKS proxy URL
	UserAgent     string            // Custom User-Agent
	Headers       map[string]string // Additional HTTP headers
	WriteInfo     bool              // Write .info.json alongside video
	WriteSubs     bool              // Write subtitles
	NoPlaylist    bool              // Don't download playlists
	PlaylistStart int               // Playlist start index (1-based)
	PlaylistEnd   int               // Playlist end index
	Verbose       bool              // Verbose logging
	CacheDir      string            // Cache directory
	Progress      ProgressFunc      // Progress callback
}

// Option is a functional option for configuring the client.
type Option func(*Options)

// WithFormat sets the format selector string.
func WithFormat(f string) Option {
	return func(o *Options) { o.Format = f }
}

// WithOutput sets the output filename template.
func WithOutput(tmpl string) Option {
	return func(o *Options) { o.Output = tmpl }
}

// WithQuiet suppresses progress output.
func WithQuiet() Option {
	return func(o *Options) { o.Quiet = true }
}

// WithNoProgress disables the progress bar.
func WithNoProgress() Option {
	return func(o *Options) { o.NoProgress = true }
}

// WithRateLimit sets the download rate limit in bytes per second.
func WithRateLimit(bytesPerSec int64) Option {
	return func(o *Options) { o.RateLimit = bytesPerSec }
}

// WithRetries sets the number of download retries.
func WithRetries(n int) Option {
	return func(o *Options) { o.Retries = n }
}

// WithContinue enables resuming partial downloads.
func WithContinue() Option {
	return func(o *Options) { o.Continue = true }
}

// WithNoPart disables .part file usage.
func WithNoPart() Option {
	return func(o *Options) { o.NoPart = true }
}

// WithCookieFile sets the cookie file path.
func WithCookieFile(path string) Option {
	return func(o *Options) { o.CookieFile = path }
}

// WithProxy sets the proxy URL.
func WithProxy(proxy string) Option {
	return func(o *Options) { o.Proxy = proxy }
}

// WithHeaders adds custom HTTP headers.
func WithHeaders(headers map[string]string) Option {
	return func(o *Options) { o.Headers = headers }
}

// WithWriteInfo enables writing .info.json files.
func WithWriteInfo() Option {
	return func(o *Options) { o.WriteInfo = true }
}

// WithWriteSubs enables writing subtitle files.
func WithWriteSubs() Option {
	return func(o *Options) { o.WriteSubs = true }
}

// WithNoPlaylist disables playlist downloading.
func WithNoPlaylist() Option {
	return func(o *Options) { o.NoPlaylist = true }
}

// WithProgress sets the progress callback.
func WithProgress(fn ProgressFunc) Option {
	return func(o *Options) { o.Progress = fn }
}

// WithVerbose enables verbose logging.
func WithVerbose() Option {
	return func(o *Options) { o.Verbose = true }
}

// WithCacheDir sets the cache directory.
func WithCacheDir(dir string) Option {
	return func(o *Options) { o.CacheDir = dir }
}

func applyOptions(opts []Option) Options {
	o := Options{
		Format:  "best",
		Retries: 3,
	}
	for _, opt := range opts {
		opt(&o)
	}

	return o
}
