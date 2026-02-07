package m3u8

import (
	"regexp"
	"strconv"
	"strings"
)

// PlaylistType indicates whether this is a master or media playlist.
type PlaylistType int

const (
	PlaylistMedia PlaylistType = iota
	PlaylistMaster
)

// Playlist represents a parsed M3U8 playlist.
type Playlist struct {
	Type           PlaylistType
	TargetDuration float64
	MediaSequence  int
	Version        int
	Segments       []Segment
	Variants       []Variant
	Keys           []Key
}

// Variant represents a variant stream in a master playlist.
type Variant struct {
	URL        string
	Bandwidth  int
	Resolution string
	Width      int
	Height     int
	Codecs     string
	Audio      string
	Video      string
	FrameRate  float64
	Name       string
}

// Segment represents a media segment.
type Segment struct {
	URL       string
	Duration  float64
	Title     string
	Key       *Key
	ByteRange *ByteRange
}

// Key represents an encryption key.
type Key struct {
	Method string // "NONE", "AES-128", "SAMPLE-AES"
	URI    string
	IV     string
}

// ByteRange represents a byte range for partial segment requests.
type ByteRange struct {
	Length int64
	Offset int64
}

var (
	attributeRe = regexp.MustCompile(`([A-Z0-9-]+)=(?:"([^"]*)"|([^,]*))`)
)

// Parse parses an M3U8 playlist string.
func Parse(content string) (*Playlist, error) {
	lines := strings.Split(strings.TrimSpace(content), "\n")
	p := &Playlist{}

	// Determine type by looking for stream-inf tags.
	for _, line := range lines {
		if strings.HasPrefix(line, "#EXT-X-STREAM-INF:") {
			p.Type = PlaylistMaster
			break
		}
	}

	if p.Type == PlaylistMaster {
		parseMaster(p, lines)
	} else {
		parseMedia(p, lines)
	}

	return p, nil
}

func parseMaster(p *Playlist, lines []string) {
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		if versionStr, ok := strings.CutPrefix(line, "#EXT-X-VERSION:"); ok {
			p.Version, _ = strconv.Atoi(versionStr)
		}

		if attrStr, ok := strings.CutPrefix(line, "#EXT-X-STREAM-INF:"); ok {
			attrs := parseAttributes(attrStr)
			v := Variant{
				Codecs: attrs["CODECS"],
				Audio:  attrs["AUDIO"],
				Video:  attrs["VIDEO"],
				Name:   attrs["NAME"],
			}

			if bw, ok := attrs["BANDWIDTH"]; ok {
				v.Bandwidth, _ = strconv.Atoi(bw)
			}

			if res, ok := attrs["RESOLUTION"]; ok {
				v.Resolution = res

				parts := strings.SplitN(res, "x", 2)
				if len(parts) == 2 {
					v.Width, _ = strconv.Atoi(parts[0])
					v.Height, _ = strconv.Atoi(parts[1])
				}
			}

			if fr, ok := attrs["FRAME-RATE"]; ok {
				v.FrameRate, _ = strconv.ParseFloat(fr, 64)
			}

			// Next non-comment line is the URL.
			for i++; i < len(lines); i++ {
				next := strings.TrimSpace(lines[i])
				if next != "" && !strings.HasPrefix(next, "#") {
					v.URL = next
					break
				}
			}

			p.Variants = append(p.Variants, v)
		}
	}
}

func parseMedia(p *Playlist, lines []string) {
	var currentKey *Key

	var segDuration float64

	var segTitle string

	for i := range len(lines) {
		line := strings.TrimSpace(lines[i])

		switch {
		case strings.HasPrefix(line, "#EXT-X-VERSION:"):
			p.Version, _ = strconv.Atoi(strings.TrimPrefix(line, "#EXT-X-VERSION:"))

		case strings.HasPrefix(line, "#EXT-X-TARGETDURATION:"):
			p.TargetDuration, _ = strconv.ParseFloat(strings.TrimPrefix(line, "#EXT-X-TARGETDURATION:"), 64)

		case strings.HasPrefix(line, "#EXT-X-MEDIA-SEQUENCE:"):
			p.MediaSequence, _ = strconv.Atoi(strings.TrimPrefix(line, "#EXT-X-MEDIA-SEQUENCE:"))

		case strings.HasPrefix(line, "#EXT-X-KEY:"):
			attrs := parseAttributes(strings.TrimPrefix(line, "#EXT-X-KEY:"))
			currentKey = &Key{
				Method: attrs["METHOD"],
				URI:    attrs["URI"],
				IV:     attrs["IV"],
			}
			p.Keys = append(p.Keys, *currentKey)

		case strings.HasPrefix(line, "#EXTINF:"):
			info := strings.TrimPrefix(line, "#EXTINF:")
			parts := strings.SplitN(info, ",", 2)

			segDuration, _ = strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
			if len(parts) > 1 {
				segTitle = strings.TrimSpace(parts[1])
			}

		case line != "" && !strings.HasPrefix(line, "#"):
			seg := Segment{
				URL:      line,
				Duration: segDuration,
				Title:    segTitle,
				Key:      currentKey,
			}
			p.Segments = append(p.Segments, seg)
			segDuration = 0
			segTitle = ""
		}
	}
}

func parseAttributes(s string) map[string]string {
	attrs := make(map[string]string)

	for _, m := range attributeRe.FindAllStringSubmatch(s, -1) {
		key := m[1]

		value := m[2]
		if value == "" {
			value = m[3]
		}

		attrs[key] = value
	}

	return attrs
}
