package sleep

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// RunSleep pauses execution for specified duration
func RunSleep(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("sleep: missing operand")
	}

	var totalDuration time.Duration

	for _, arg := range args {
		d, err := parseSleepDuration(arg)
		if err != nil {
			return fmt.Errorf("sleep: invalid time interval %q", arg)
		}

		totalDuration += d
	}

	time.Sleep(totalDuration)

	return nil
}

func parseSleepDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty duration")
	}

	// Check for suffix
	suffix := s[len(s)-1]

	var multiplier = time.Second // default

	switch suffix {
	case 's':
		multiplier = time.Second
		s = s[:len(s)-1]
	case 'm':
		multiplier = time.Minute
		s = s[:len(s)-1]
	case 'h':
		multiplier = time.Hour
		s = s[:len(s)-1]
	case 'd':
		multiplier = 24 * time.Hour
		s = s[:len(s)-1]
	}

	// Parse the number (supports decimals)
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}

	return time.Duration(val * float64(multiplier)), nil
}
