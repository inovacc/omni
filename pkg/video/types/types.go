// Package types defines shared data structures for the video download engine.
// This package exists to break import cycles between pkg/video and its sub-packages.
package types

// VideoInfo holds metadata and formats for a single video or playlist.
type VideoInfo struct {
	ID            string                `json:"id"`
	Title         string                `json:"title"`
	Formats       []Format              `json:"formats,omitempty"`
	URL           string                `json:"url,omitempty"`
	Ext           string                `json:"ext,omitempty"`
	Description   string                `json:"description,omitempty"`
	Uploader      string                `json:"uploader,omitempty"`
	UploaderID    string                `json:"uploader_id,omitempty"`
	UploaderURL   string                `json:"uploader_url,omitempty"`
	Channel       string                `json:"channel,omitempty"`
	ChannelID     string                `json:"channel_id,omitempty"`
	ChannelURL    string                `json:"channel_url,omitempty"`
	Duration      float64               `json:"duration,omitempty"`
	ViewCount     *int64                `json:"view_count,omitempty"`
	LikeCount     *int64                `json:"like_count,omitempty"`
	DislikeCount  *int64                `json:"dislike_count,omitempty"`
	AverageRating *float64              `json:"average_rating,omitempty"`
	AgeLimit      int                   `json:"age_limit,omitempty"`
	WebpageURL    string                `json:"webpage_url,omitempty"`
	Categories    []string              `json:"categories,omitempty"`
	Tags          []string              `json:"tags,omitempty"`
	IsLive        *bool                 `json:"is_live,omitempty"`
	UploadDate    string                `json:"upload_date,omitempty"`
	Thumbnails    []Thumbnail           `json:"thumbnails,omitempty"`
	Subtitles     map[string][]Subtitle `json:"subtitles,omitempty"`
	Chapters      []Chapter             `json:"chapters,omitempty"`
	Series        string                `json:"series,omitempty"`
	Season        string                `json:"season,omitempty"`
	SeasonNumber  *int                  `json:"season_number,omitempty"`
	Episode       string                `json:"episode,omitempty"`
	EpisodeNumber *int                  `json:"episode_number,omitempty"`
	PlaylistTitle string                `json:"playlist_title,omitempty"`
	PlaylistIndex *int                  `json:"playlist_index,omitempty"`
	HTTPHeaders   map[string]string     `json:"http_headers,omitempty"`
	Extractor     string                `json:"extractor,omitempty"`
	ExtractorKey  string                `json:"extractor_key,omitempty"`

	// Extra metadata for extractor-specific data (e.g., channel subscriber count).
	Metadata map[string]string `json:"metadata,omitempty"`

	// Playlist fields
	Type    string      `json:"_type,omitempty"` // "video", "playlist", "url", "url_transparent"
	Entries []VideoInfo `json:"entries,omitempty"`
}

// Format describes a single downloadable media format.
type Format struct {
	URL            string            `json:"url"`
	ManifestURL    string            `json:"manifest_url,omitempty"`
	FormatID       string            `json:"format_id"`
	FormatNote     string            `json:"format_note,omitempty"`
	Ext            string            `json:"ext,omitempty"`
	Width          *int              `json:"width,omitempty"`
	Height         *int              `json:"height,omitempty"`
	Resolution     string            `json:"resolution,omitempty"`
	FPS            *float64          `json:"fps,omitempty"`
	VCodec         string            `json:"vcodec,omitempty"`
	ACodec         string            `json:"acodec,omitempty"`
	ABR            *float64          `json:"abr,omitempty"`
	VBR            *float64          `json:"vbr,omitempty"`
	TBR            *float64          `json:"tbr,omitempty"`
	ASR            *int              `json:"asr,omitempty"`
	Filesize       *int64            `json:"filesize,omitempty"`
	FilesizeApprox *int64            `json:"filesize_approx,omitempty"`
	Protocol       string            `json:"protocol,omitempty"`
	Preference     *int              `json:"preference,omitempty"`
	Quality        *float64          `json:"quality,omitempty"`
	Language       string            `json:"language,omitempty"`
	Fragments      []Fragment        `json:"fragments,omitempty"`
	HTTPHeaders    map[string]string `json:"http_headers,omitempty"`
	Container      string            `json:"container,omitempty"`
	SourcePref     *int              `json:"source_preference,omitempty"`
}

// Fragment describes a single fragment/segment of a fragmented download.
type Fragment struct {
	URL      string  `json:"url,omitempty"`
	Path     string  `json:"path,omitempty"`
	Duration float64 `json:"duration,omitempty"`
}

// Thumbnail describes a video thumbnail.
type Thumbnail struct {
	URL        string `json:"url"`
	ID         string `json:"id,omitempty"`
	Width      *int   `json:"width,omitempty"`
	Height     *int   `json:"height,omitempty"`
	Resolution string `json:"resolution,omitempty"`
	Preference int    `json:"preference,omitempty"`
}

// Subtitle describes a subtitle file.
type Subtitle struct {
	URL  string `json:"url"`
	Ext  string `json:"ext,omitempty"`
	Data string `json:"data,omitempty"`
}

// Chapter describes a video chapter.
type Chapter struct {
	StartTime float64 `json:"start_time"`
	EndTime   float64 `json:"end_time"`
	Title     string  `json:"title"`
}

// ProgressInfo reports download progress.
type ProgressInfo struct {
	Status          string   `json:"status"` // "downloading", "finished", "error"
	Filename        string   `json:"filename"`
	DownloadedBytes int64    `json:"downloaded_bytes"`
	TotalBytes      *int64   `json:"total_bytes,omitempty"`
	Elapsed         float64  `json:"elapsed"`
	ETA             *float64 `json:"eta,omitempty"`
	Speed           *float64 `json:"speed,omitempty"`
	FragmentIndex   *int     `json:"fragment_index,omitempty"`
	FragmentCount   *int     `json:"fragment_count,omitempty"`
}

// ProgressFunc is called during download to report progress.
type ProgressFunc func(ProgressInfo)

// HasVideo returns true if the format has a video stream.
func (f Format) HasVideo() bool {
	return f.VCodec != "" && f.VCodec != "none"
}

// HasAudio returns true if the format has an audio stream.
func (f Format) HasAudio() bool {
	return f.ACodec != "" && f.ACodec != "none"
}

// GetFilesize returns the best available filesize estimate.
func (f Format) GetFilesize() int64 {
	if f.Filesize != nil {
		return *f.Filesize
	}

	if f.FilesizeApprox != nil {
		return *f.FilesizeApprox
	}

	return 0
}

// FormatResolution returns a human-readable resolution string.
func (f Format) FormatResolution() string {
	if f.Resolution != "" {
		return f.Resolution
	}

	if f.Height != nil {
		h := *f.Height
		if f.Width != nil {
			return itoa(*f.Width) + "x" + itoa(h)
		}

		return itoa(h) + "p"
	}

	return "audio only"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}

	var digits []byte

	neg := n < 0
	if neg {
		n = -n
	}

	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}

	if neg {
		digits = append([]byte{'-'}, digits...)
	}

	return string(digits)
}

// ExtractorError is returned when an extractor fails to extract video info.
type ExtractorError struct {
	Extractor string
	VideoID   string
	Message   string
	Cause     error
}

func (e *ExtractorError) Error() string {
	msg := e.Extractor + ": " + e.Message
	if e.VideoID != "" {
		msg = e.Extractor + " [" + e.VideoID + "]: " + e.Message
	}

	if e.Cause != nil {
		msg += ": " + e.Cause.Error()
	}

	return msg
}

func (e *ExtractorError) Unwrap() error { return e.Cause }

// UnsupportedError is returned when the URL is not supported by any extractor.
type UnsupportedError struct {
	URL string
}

func (e *UnsupportedError) Error() string {
	return "unsupported URL: " + e.URL
}

// GeoRestrictedError is returned when content is geo-restricted.
type GeoRestrictedError struct {
	Countries []string
	Message   string
}

func (e *GeoRestrictedError) Error() string {
	if e.Message != "" {
		return "geo-restricted: " + e.Message
	}

	return "geo-restricted content"
}
