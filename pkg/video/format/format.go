package format

import (
	"sort"
	"strings"

	"github.com/inovacc/omni/pkg/video/types"
)

// Extension preference order (higher index = more preferred).
var extPreference = map[string]int{
	"3gp":  0,
	"flv":  1,
	"mp4":  2,
	"webm": 3,
	"mkv":  4,
}

// SortFormats sorts formats by quality, best last.
// Order: resolution (height) → tbr → ext preference → format_id.
func SortFormats(formats []types.Format) {
	sort.SliceStable(formats, func(i, j int) bool {
		a, b := formats[i], formats[j]

		// Preference override.
		ap, bp := preferenceVal(a.Preference), preferenceVal(b.Preference)
		if ap != bp {
			return ap < bp
		}

		// Has video vs audio-only.
		av, bv := a.HasVideo(), b.HasVideo()
		if av != bv {
			return !av // audio-only comes first (lower quality).
		}

		// Height.
		ah, bh := heightVal(a.Height), heightVal(b.Height)
		if ah != bh {
			return ah < bh
		}

		// Total bitrate.
		at, bt := tbrVal(a.TBR), tbrVal(b.TBR)
		if at != bt {
			return at < bt
		}

		// Extension preference.
		ae := extPreference[strings.ToLower(a.Ext)]

		be := extPreference[strings.ToLower(b.Ext)]
		if ae != be {
			return ae < be
		}

		// Filesize.
		af, bf := a.GetFilesize(), b.GetFilesize()
		if af != bf {
			return af < bf
		}

		return a.FormatID < b.FormatID
	})
}

// BestFormat returns the best format from a sorted list.
// Prefers formats with both video and audio.
func BestFormat(formats []types.Format) *types.Format {
	if len(formats) == 0 {
		return nil
	}

	// Look for best format with both video and audio first.
	for i := len(formats) - 1; i >= 0; i-- {
		f := &formats[i]
		if f.HasVideo() && f.HasAudio() {
			return f
		}
	}

	// Fall back to last (best quality).
	return &formats[len(formats)-1]
}

// WorstFormat returns the worst format from a sorted list.
func WorstFormat(formats []types.Format) *types.Format {
	if len(formats) == 0 {
		return nil
	}

	// Look for worst format with both video and audio first.
	for i := range formats {
		f := &formats[i]
		if f.HasVideo() && f.HasAudio() {
			return f
		}
	}

	return &formats[0]
}

// FilterFormats returns formats matching a predicate.
func FilterFormats(formats []types.Format, pred func(types.Format) bool) []types.Format {
	var result []types.Format

	for _, f := range formats {
		if pred(f) {
			result = append(result, f)
		}
	}

	return result
}

// BestVideoOnly returns the best video-only format.
func BestVideoOnly(formats []types.Format) *types.Format {
	filtered := FilterFormats(formats, func(f types.Format) bool {
		return f.HasVideo() && !f.HasAudio()
	})
	if len(filtered) == 0 {
		return nil
	}

	return &filtered[len(filtered)-1]
}

// BestAudioOnly returns the best audio-only format.
func BestAudioOnly(formats []types.Format) *types.Format {
	filtered := FilterFormats(formats, func(f types.Format) bool {
		return f.HasAudio() && !f.HasVideo()
	})
	if len(filtered) == 0 {
		return nil
	}

	return &filtered[len(filtered)-1]
}

func preferenceVal(p *int) int {
	if p != nil {
		return *p
	}

	return 0
}

func heightVal(h *int) int {
	if h != nil {
		return *h
	}

	return 0
}

func tbrVal(t *float64) float64 {
	if t != nil {
		return *t
	}

	return 0
}
