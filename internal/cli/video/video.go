package video

// Options holds all CLI options for the video command.
type Options struct {
	Format             string
	Output             string
	Quiet              bool
	NoProgress         bool
	RateLimit          string // e.g., "1M", "500K"
	Retries            int
	Continue           bool
	NoPart             bool
	CookieFile         string
	Proxy              string
	WriteInfoJSON      bool
	WriteSubs          bool
	NoPlaylist         bool
	PlaylistStart      int
	PlaylistEnd        int
	Verbose            bool
	JSON               bool
	Complete           bool
	Limit              int  // Max videos for channel command (-1 = all)
	CookiesFromBrowser bool // Auto-load cookies from well-known path
}
