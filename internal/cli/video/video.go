package video

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/inovacc/omni/internal/cli/cmderr"
	videotypes "github.com/inovacc/omni/pkg/video/types"
)

// wrapVideoErr classifies pkg/video errors into cmderr sentinels.
// Call at the CLI boundary after Extract or Download returns an error.
func wrapVideoErr(cmd string, err error) error {
	if err == nil {
		return nil
	}

	var unsupported *videotypes.UnsupportedError
	if errors.As(err, &unsupported) {
		return cmderr.Wrap(cmderr.ErrUnsupported, fmt.Sprintf("%s: %s", cmd, err))
	}

	var extractErr *videotypes.ExtractorError
	if errors.As(err, &extractErr) {
		return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("%s: %s", cmd, err))
	}

	var geoErr *videotypes.GeoRestrictedError
	if errors.As(err, &geoErr) {
		return cmderr.Wrap(cmderr.ErrUnsupported, fmt.Sprintf("%s: %s", cmd, err))
	}

	return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("%s: %s", cmd, err))
}

// validateVideoURL returns ErrInvalidInput if the URL is clearly not a video URL.
func validateVideoURL(raw string) error {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return cmderr.Wrap(cmderr.ErrInvalidInput, "video: URL is required")
	}

	// Allow ytsearch: and other extractor prefixes unchanged.
	if strings.Contains(trimmed, ":") && !strings.HasPrefix(trimmed, "http") {
		return nil
	}

	// Must parse as a URL with a scheme, or be a bare YouTube ID (11 chars).
	u, err := url.Parse(trimmed)
	if err != nil || (u.Scheme == "" && len(trimmed) != 11) {
		return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("video: invalid URL: %s", raw))
	}

	return nil
}

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
