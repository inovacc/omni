package uptime

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

// UptimeOptions configures the uptime command behavior
type UptimeOptions struct {
	Pretty bool // -p: show uptime in pretty format
	Since  bool // -s: system up since
	JSON   bool // --json: output as JSON
}

// UptimeInfo contains system uptime information
type UptimeInfo struct {
	Uptime    time.Duration `json:"uptime"`
	BootTime  time.Time     `json:"bootTime"`
	Users     int           `json:"users"`
	LoadAvg1  float64       `json:"loadAvg1"`
	LoadAvg5  float64       `json:"loadAvg5"`
	LoadAvg15 float64       `json:"loadAvg15"`
}

// RunUptime shows how long the system has been running
func RunUptime(w io.Writer, opts UptimeOptions) error {
	info, err := getUptimeInfo()
	if err != nil {
		return fmt.Errorf("uptime: %w", err)
	}

	if opts.JSON {
		return json.NewEncoder(w).Encode(info)
	}

	if opts.Since {
		_, _ = fmt.Fprintln(w, info.BootTime.Format("2006-01-02 15:04:05"))
		return nil
	}

	if opts.Pretty {
		_, _ = fmt.Fprintf(w, "up %s\n", formatPrettyUptime(info.Uptime))
		return nil
	}

	// Default format: current time, up time, users, load average
	now := time.Now().Format("15:04:05")
	uptimeStr := formatUptime(info.Uptime)

	_, _ = fmt.Fprintf(w, " %s up %s, %d user(s), load average: %.2f, %.2f, %.2f\n",
		now, uptimeStr, info.Users, info.LoadAvg1, info.LoadAvg5, info.LoadAvg15)

	return nil
}

func formatUptime(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%d day(s), %d:%02d", days, hours, minutes)
	}

	return fmt.Sprintf("%d:%02d", hours, minutes)
}

func formatPrettyUptime(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	var parts []string

	if days > 0 {
		if days == 1 {
			parts = append(parts, "1 day")
		} else {
			parts = append(parts, fmt.Sprintf("%d days", days))
		}
	}

	if hours > 0 {
		if hours == 1 {
			parts = append(parts, "1 hour")
		} else {
			parts = append(parts, fmt.Sprintf("%d hours", hours))
		}
	}

	if minutes > 0 || len(parts) == 0 {
		if minutes == 1 {
			parts = append(parts, "1 minute")
		} else {
			parts = append(parts, fmt.Sprintf("%d minutes", minutes))
		}
	}

	var result strings.Builder

	for i, part := range parts {
		if i > 0 {
			if i == len(parts)-1 {
				result.WriteString(" and ")
			} else {
				result.WriteString(", ")
			}
		}

		result.WriteString(part)
	}

	return result.String()
}

// GetUptime returns the system uptime duration
func GetUptime() (time.Duration, error) {
	info, err := getUptimeInfo()
	if err != nil {
		return 0, err
	}

	return info.Uptime, nil
}
