package video

import "github.com/inovacc/omni/pkg/video/types"

// Re-export error types from the types sub-package.
type (
	ExtractorError     = types.ExtractorError
	UnsupportedError   = types.UnsupportedError
	GeoRestrictedError = types.GeoRestrictedError
)

// DownloadError is returned when a download fails.
type DownloadError struct {
	URL     string
	Message string
	Cause   error
}

func (e *DownloadError) Error() string {
	msg := "download: " + e.Message
	if e.Cause != nil {
		msg += ": " + e.Cause.Error()
	}

	return msg
}

func (e *DownloadError) Unwrap() error { return e.Cause }

// AgeRestrictedError is returned when age verification is required.
type AgeRestrictedError struct {
	Message string
}

func (e *AgeRestrictedError) Error() string {
	if e.Message != "" {
		return "age-restricted: " + e.Message
	}

	return "age-restricted content"
}
