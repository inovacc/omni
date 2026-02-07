package video

import "github.com/inovacc/omni/pkg/video/types"

// Re-export types from the types sub-package for convenience.
// External consumers can import either pkg/video or pkg/video/types.
type (
	VideoInfo    = types.VideoInfo
	Format       = types.Format
	Fragment     = types.Fragment
	Thumbnail    = types.Thumbnail
	Subtitle     = types.Subtitle
	Chapter      = types.Chapter
	ProgressInfo = types.ProgressInfo
	ProgressFunc = types.ProgressFunc
)
