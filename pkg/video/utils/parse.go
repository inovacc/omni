package utils

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	// Duration patterns: "1h30m", "PT1H30M", "01:30:00", "5400", etc.
	durationISORe  = regexp.MustCompile(`^P(?:(\d+)D)?T?(?:(\d+)H)?(?:(\d+)M)?(?:(\d+(?:\.\d+)?)S)?$`)
	durationHMSRe  = regexp.MustCompile(`^(?:(\d+):)?(\d{1,2}):(\d{2})(?:\.(\d+))?$`)
	durationSecRe  = regexp.MustCompile(`^(\d+(?:\.\d+)?)$`)
	durationWordRe = regexp.MustCompile(`(?i)(?:(\d+)\s*(?:hours?|h))\s*(?:(\d+)\s*(?:min(?:utes?)?|m))?\s*(?:(\d+)\s*(?:sec(?:onds?)?|s))?`)

	// Filesize patterns: "1.5 MB", "100kb", etc.
	filesizeRe = regexp.MustCompile(`(?i)^([\d.]+)\s*([KMGTP]?)(I?)B?$`)

	// ISO 8601 date: "2024-01-15T10:30:00Z" or "20240115"
	isoDateRe  = regexp.MustCompile(`^(\d{4})-?(\d{2})-?(\d{2})(?:T(\d{2}):?(\d{2}):?(\d{2})(?:\.(\d+))?(?:Z|([+-]\d{2}:?\d{2}))?)?$`)
	yyyymmddRe = regexp.MustCompile(`^(\d{4})(\d{2})(\d{2})$`)
)

// ParseDuration parses various duration string formats to seconds.
// Supports: ISO 8601 (PT1H30M), HH:MM:SS, seconds, natural language ("1h 30m").
func ParseDuration(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}

	// Try pure seconds.
	if m := durationSecRe.FindStringSubmatch(s); m != nil {
		v, err := strconv.ParseFloat(m[1], 64)
		if err == nil {
			return v, true
		}
	}

	// Try HH:MM:SS or MM:SS.
	if m := durationHMSRe.FindStringSubmatch(s); m != nil {
		var hours, minutes, seconds float64
		if m[1] != "" {
			hours, _ = strconv.ParseFloat(m[1], 64)
		}

		minutes, _ = strconv.ParseFloat(m[2], 64)
		seconds, _ = strconv.ParseFloat(m[3], 64)

		if m[4] != "" {
			frac, _ := strconv.ParseFloat("0."+m[4], 64)
			seconds += frac
		}

		return hours*3600 + minutes*60 + seconds, true
	}

	// Try ISO 8601 duration.
	upper := strings.ToUpper(s)
	if m := durationISORe.FindStringSubmatch(upper); m != nil {
		var total float64

		if m[1] != "" {
			d, _ := strconv.ParseFloat(m[1], 64)
			total += d * 86400
		}

		if m[2] != "" {
			h, _ := strconv.ParseFloat(m[2], 64)
			total += h * 3600
		}

		if m[3] != "" {
			min, _ := strconv.ParseFloat(m[3], 64)
			total += min * 60
		}

		if m[4] != "" {
			sec, _ := strconv.ParseFloat(m[4], 64)
			total += sec
		}

		if total > 0 {
			return total, true
		}
	}

	// Try natural language ("1 hour 30 minutes").
	if m := durationWordRe.FindStringSubmatch(s); m != nil {
		var total float64

		if m[1] != "" {
			h, _ := strconv.ParseFloat(m[1], 64)
			total += h * 3600
		}

		if m[2] != "" {
			min, _ := strconv.ParseFloat(m[2], 64)
			total += min * 60
		}

		if m[3] != "" {
			sec, _ := strconv.ParseFloat(m[3], 64)
			total += sec
		}

		if total > 0 {
			return total, true
		}
	}

	return 0, false
}

// ParseFilesize parses a filesize string into bytes.
func ParseFilesize(s string) (int64, bool) {
	s = strings.TrimSpace(s)

	m := filesizeRe.FindStringSubmatch(s)
	if m == nil {
		return 0, false
	}

	val, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return 0, false
	}

	// Determine multiplier.
	var base float64 = 1000
	if m[3] == "i" || m[3] == "I" {
		base = 1024
	}

	prefix := strings.ToUpper(m[2])
	switch prefix {
	case "K":
		val *= base
	case "M":
		val *= base * base
	case "G":
		val *= base * base * base
	case "T":
		val *= base * base * base * base
	case "P":
		val *= base * base * base * base * base
	}

	return int64(math.Round(val)), true
}

// ParseISO8601 parses an ISO 8601 date string to time.Time.
func ParseISO8601(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)

	// Try YYYYMMDD compact.
	if m := yyyymmddRe.FindStringSubmatch(s); m != nil {
		year, _ := strconv.Atoi(m[1])
		month, _ := strconv.Atoi(m[2])
		day, _ := strconv.Atoi(m[3])

		return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC), true
	}

	// Try full ISO 8601.
	if m := isoDateRe.FindStringSubmatch(s); m != nil {
		year, _ := strconv.Atoi(m[1])
		month, _ := strconv.Atoi(m[2])
		day, _ := strconv.Atoi(m[3])

		var hour, min, sec int
		if m[4] != "" {
			hour, _ = strconv.Atoi(m[4])
		}

		if m[5] != "" {
			min, _ = strconv.Atoi(m[5])
		}

		if m[6] != "" {
			sec, _ = strconv.Atoi(m[6])
		}

		loc := time.UTC

		if m[8] != "" {
			offset := strings.ReplaceAll(m[8], ":", "")

			sign := 1
			if offset[0] == '-' {
				sign = -1
			}

			offsetH, _ := strconv.Atoi(offset[1:3])

			offsetM := 0
			if len(offset) >= 5 {
				offsetM, _ = strconv.Atoi(offset[3:5])
			}

			totalOffset := sign * (offsetH*3600 + offsetM*60)
			loc = time.FixedZone("", totalOffset)
		}

		return time.Date(year, time.Month(month), day, hour, min, sec, 0, loc), true
	}

	return time.Time{}, false
}

// IntOrNone converts a string to *int64, returning nil on failure.
func IntOrNone(s string) *int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	// Remove commas.
	s = strings.ReplaceAll(s, ",", "")

	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil
	}

	return &n
}

// FloatOrNone converts a string to *float64, returning nil on failure.
func FloatOrNone(s string) *float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	s = strings.ReplaceAll(s, ",", "")

	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}

	return &f
}

// StrOrNone returns nil if the string is empty, or a pointer to it otherwise.
func StrOrNone(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	return &s
}

// FormatDate converts a time.Time to YYYYMMDD string.
func FormatDate(t time.Time) string {
	return t.Format("20060102")
}
