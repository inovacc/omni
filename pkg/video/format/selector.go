package format

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/inovacc/omni/pkg/video/types"
)

// Selector selects formats from a list based on a format specification string.
// Supported specs:
//   - "best" / "worst" — best/worst format with video+audio
//   - "bestvideo" / "worstvideo" — best/worst video-only
//   - "bestaudio" / "worstaudio" — best/worst audio-only
//   - "FORMAT_ID" — specific format by ID
//   - "best[height<=720]" — best format matching filter
type Selector struct {
	spec string
}

// NewSelector creates a format selector from a spec string.
func NewSelector(spec string) *Selector {
	if spec == "" {
		spec = "best"
	}

	return &Selector{spec: spec}
}

// Select chooses format(s) from the given list.
func (s *Selector) Select(formats []types.Format) ([]types.Format, error) {
	if len(formats) == 0 {
		return nil, fmt.Errorf("format: no formats available")
	}

	SortFormats(formats)

	spec := strings.TrimSpace(s.spec)

	// Handle filter expressions like "best[height<=720]"
	base, filter := parseFilter(spec)

	var filtered []types.Format
	if filter != nil {
		filtered = FilterFormats(formats, filter)
	} else {
		filtered = formats
	}

	switch base {
	case "best":
		f := BestFormat(filtered)
		if f == nil {
			return nil, fmt.Errorf("format: no matching format for %q", s.spec)
		}

		return []types.Format{*f}, nil

	case "worst":
		f := WorstFormat(filtered)
		if f == nil {
			return nil, fmt.Errorf("format: no matching format for %q", s.spec)
		}

		return []types.Format{*f}, nil

	case "bestvideo":
		f := BestVideoOnly(filtered)
		if f == nil {
			// Fall back to best format with video.
			vf := FilterFormats(filtered, func(f types.Format) bool { return f.HasVideo() })
			if len(vf) > 0 {
				return []types.Format{vf[len(vf)-1]}, nil
			}

			return nil, fmt.Errorf("format: no video format found")
		}

		return []types.Format{*f}, nil

	case "worstvideo":
		vf := FilterFormats(filtered, func(f types.Format) bool { return f.HasVideo() })
		if len(vf) == 0 {
			return nil, fmt.Errorf("format: no video format found")
		}

		return []types.Format{vf[0]}, nil

	case "bestaudio":
		f := BestAudioOnly(filtered)
		if f == nil {
			af := FilterFormats(filtered, func(f types.Format) bool { return f.HasAudio() })
			if len(af) > 0 {
				return []types.Format{af[len(af)-1]}, nil
			}

			return nil, fmt.Errorf("format: no audio format found")
		}

		return []types.Format{*f}, nil

	case "worstaudio":
		af := FilterFormats(filtered, func(f types.Format) bool { return f.HasAudio() })
		if len(af) == 0 {
			return nil, fmt.Errorf("format: no audio format found")
		}

		return []types.Format{af[0]}, nil

	default:
		// Try format ID match.
		for _, f := range formats {
			if f.FormatID == spec {
				return []types.Format{f}, nil
			}
		}

		return nil, fmt.Errorf("format: no format matching %q", s.spec)
	}
}

// parseFilter splits "best[height<=720]" into ("best", filterFunc).
func parseFilter(spec string) (string, func(types.Format) bool) {
	base, filterStr, found := strings.Cut(spec, "[")
	if !found {
		return spec, nil
	}

	filterStr = strings.TrimSuffix(filterStr, "]")
	filter := parseFilterExpr(filterStr)

	return base, filter
}

func parseFilterExpr(expr string) func(types.Format) bool {
	// Support: height<=720, height>=480, ext=mp4, vcodec!=none
	operators := []struct {
		op  string
		cmp func(a, b float64) bool
	}{
		{"<=", func(a, b float64) bool { return a <= b }},
		{">=", func(a, b float64) bool { return a >= b }},
		{"!=", nil}, // Handled separately for string comparison.
		{"=", nil},
		{"<", func(a, b float64) bool { return a < b }},
		{">", func(a, b float64) bool { return a > b }},
	}

	for _, op := range operators {
		parts := strings.SplitN(expr, op.op, 2)
		if len(parts) != 2 {
			continue
		}

		field := strings.TrimSpace(parts[0])
		valStr := strings.TrimSpace(parts[1])

		if op.cmp != nil {
			// Numeric comparison.
			target, err := strconv.ParseFloat(valStr, 64)
			if err != nil {
				continue
			}

			return makeNumericFilter(field, target, op.cmp)
		}

		// String comparison (= or !=).
		negate := op.op == "!="

		return func(f types.Format) bool {
			actual := getStringField(f, field)

			match := strings.EqualFold(actual, valStr)
			if negate {
				return !match
			}

			return match
		}
	}

	return nil
}

func makeNumericFilter(field string, target float64, cmp func(float64, float64) bool) func(types.Format) bool {
	return func(f types.Format) bool {
		val := getNumericField(f, field)
		if val == nil {
			return false
		}

		return cmp(*val, target)
	}
}

func getNumericField(f types.Format, field string) *float64 {
	switch field {
	case "height":
		if f.Height != nil {
			v := float64(*f.Height)
			return &v
		}
	case "width":
		if f.Width != nil {
			v := float64(*f.Width)
			return &v
		}
	case "tbr":
		return f.TBR
	case "abr":
		return f.ABR
	case "vbr":
		return f.VBR
	case "fps":
		return f.FPS
	case "filesize":
		if f.Filesize != nil {
			v := float64(*f.Filesize)
			return &v
		}
	}

	return nil
}

func getStringField(f types.Format, field string) string {
	switch field {
	case "ext":
		return f.Ext
	case "vcodec":
		return f.VCodec
	case "acodec":
		return f.ACodec
	case "protocol":
		return f.Protocol
	case "format_id":
		return f.FormatID
	case "container":
		return f.Container
	case "language":
		return f.Language
	}

	return ""
}
