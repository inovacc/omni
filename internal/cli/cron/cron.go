package cron

import (
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Schedule represents a parsed cron schedule
type Schedule struct {
	Expression string   `json:"expression"`
	Minutes    []int    `json:"minutes"`
	Hours      []int    `json:"hours"`
	Days       []int    `json:"days"`
	Months     []int    `json:"months"`
	Weekdays   []int    `json:"weekdays"`
	Human      string   `json:"human"`
	NextRuns   []string `json:"next_runs,omitempty"`
}

// Options configures the cron command behavior
type Options struct {
	JSON     bool // output as JSON
	Next     int  // show next N run times
	Validate bool // just validate the expression
}

// Run parses and displays cron expression information
func Run(w io.Writer, args []string, opts Options) error {
	if len(args) == 0 {
		return fmt.Errorf("cron: missing cron expression")
	}

	expr := strings.Join(args, " ")

	schedule, err := Parse(expr)
	if err != nil {
		return fmt.Errorf("cron: %w", err)
	}

	if opts.Validate {
		if opts.JSON {
			result := map[string]any{
				"valid":      true,
				"expression": expr,
			}
			enc := json.NewEncoder(w)
			enc.SetIndent("", "  ")

			return enc.Encode(result)
		}

		_, _ = fmt.Fprintf(w, "Valid cron expression: %s\n", expr)

		return nil
	}

	if opts.Next > 0 {
		schedule.NextRuns = getNextRuns(schedule, opts.Next)
	}

	if opts.JSON {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")

		return enc.Encode(schedule)
	}

	// Human-readable output
	_, _ = fmt.Fprintf(w, "Expression: %s\n", schedule.Expression)
	_, _ = fmt.Fprintf(w, "Schedule:   %s\n", schedule.Human)
	_, _ = fmt.Fprintf(w, "\n")
	_, _ = fmt.Fprintf(w, "Minutes:    %s\n", formatList(schedule.Minutes))
	_, _ = fmt.Fprintf(w, "Hours:      %s\n", formatList(schedule.Hours))
	_, _ = fmt.Fprintf(w, "Days:       %s\n", formatList(schedule.Days))
	_, _ = fmt.Fprintf(w, "Months:     %s\n", formatList(schedule.Months))
	_, _ = fmt.Fprintf(w, "Weekdays:   %s\n", formatWeekdays(schedule.Weekdays))

	if len(schedule.NextRuns) > 0 {
		_, _ = fmt.Fprintf(w, "\nNext runs:\n")
		for _, run := range schedule.NextRuns {
			_, _ = fmt.Fprintf(w, "  %s\n", run)
		}
	}

	return nil
}

// Parse parses a cron expression into a Schedule
func Parse(expr string) (*Schedule, error) {
	expr = strings.TrimSpace(expr)

	// Handle common aliases
	expr = expandAlias(expr)

	parts := strings.Fields(expr)
	if len(parts) != 5 {
		return nil, fmt.Errorf("invalid cron expression: expected 5 fields, got %d", len(parts))
	}

	minutes, err := parseField(parts[0], 0, 59)
	if err != nil {
		return nil, fmt.Errorf("invalid minutes field: %w", err)
	}

	hours, err := parseField(parts[1], 0, 23)
	if err != nil {
		return nil, fmt.Errorf("invalid hours field: %w", err)
	}

	days, err := parseField(parts[2], 1, 31)
	if err != nil {
		return nil, fmt.Errorf("invalid days field: %w", err)
	}

	months, err := parseField(parts[3], 1, 12)
	if err != nil {
		return nil, fmt.Errorf("invalid months field: %w", err)
	}

	weekdays, err := parseField(parts[4], 0, 7) // 0 and 7 both represent Sunday
	if err != nil {
		return nil, fmt.Errorf("invalid weekdays field: %w", err)
	}

	// Normalize weekday 7 to 0 (both mean Sunday)
	weekdays = normalizeWeekdays(weekdays)

	schedule := &Schedule{
		Expression: expr,
		Minutes:    minutes,
		Hours:      hours,
		Days:       days,
		Months:     months,
		Weekdays:   weekdays,
		Human:      generateHumanDescription(minutes, hours, days, months, weekdays),
	}

	return schedule, nil
}

// expandAlias expands common cron aliases
func expandAlias(expr string) string {
	aliases := map[string]string{
		"@yearly":   "0 0 1 1 *",
		"@annually": "0 0 1 1 *",
		"@monthly":  "0 0 1 * *",
		"@weekly":   "0 0 * * 0",
		"@daily":    "0 0 * * *",
		"@midnight": "0 0 * * *",
		"@hourly":   "0 * * * *",
	}

	lower := strings.ToLower(expr)
	if expanded, ok := aliases[lower]; ok {
		return expanded
	}

	return expr
}

// parseField parses a single cron field
func parseField(field string, min, max int) ([]int, error) {
	var result []int

	// Handle * (all values)
	if field == "*" {
		for i := min; i <= max; i++ {
			result = append(result, i)
		}

		return result, nil
	}

	// Handle */step
	if strings.HasPrefix(field, "*/") {
		step, err := strconv.Atoi(field[2:])
		if err != nil || step <= 0 {
			return nil, fmt.Errorf("invalid step: %s", field)
		}

		for i := min; i <= max; i += step {
			result = append(result, i)
		}

		return result, nil
	}

	// Handle comma-separated values and ranges
	parts := strings.SplitSeq(field, ",")
	for part := range parts {
		values, err := parseRangeOrValue(part, min, max)
		if err != nil {
			return nil, err
		}

		result = append(result, values...)
	}

	// Remove duplicates and sort
	result = unique(result)

	return result, nil
}

// parseRangeOrValue parses a single value or range (e.g., "5" or "1-5" or "1-10/2")
func parseRangeOrValue(s string, min, max int) ([]int, error) {
	var result []int

	// Handle range with step (e.g., "1-10/2")
	if strings.Contains(s, "/") {
		parts := strings.SplitN(s, "/", 2)

		step, err := strconv.Atoi(parts[1])
		if err != nil || step <= 0 {
			return nil, fmt.Errorf("invalid step: %s", s)
		}

		rangeVals, err := parseRangeOrValue(parts[0], min, max)
		if err != nil {
			return nil, err
		}

		for i, v := range rangeVals {
			if i%step == 0 {
				result = append(result, v)
			}
		}

		return result, nil
	}

	// Handle range (e.g., "1-5")
	if strings.Contains(s, "-") {
		parts := strings.SplitN(s, "-", 2)

		start, err := parseValue(parts[0], min, max)
		if err != nil {
			return nil, err
		}

		end, err := parseValue(parts[1], min, max)
		if err != nil {
			return nil, err
		}

		if start > end {
			return nil, fmt.Errorf("invalid range: start > end")
		}

		for i := start; i <= end; i++ {
			result = append(result, i)
		}

		return result, nil
	}

	// Handle single value
	val, err := parseValue(s, min, max)
	if err != nil {
		return nil, err
	}

	return []int{val}, nil
}

// parseValue parses a single value, handling month and weekday names
func parseValue(s string, min, max int) (int, error) {
	// Try month names
	months := map[string]int{
		"jan": 1, "feb": 2, "mar": 3, "apr": 4,
		"may": 5, "jun": 6, "jul": 7, "aug": 8,
		"sep": 9, "oct": 10, "nov": 11, "dec": 12,
	}
	if v, ok := months[strings.ToLower(s)]; ok {
		return v, nil
	}

	// Try weekday names
	weekdays := map[string]int{
		"sun": 0, "mon": 1, "tue": 2, "wed": 3,
		"thu": 4, "fri": 5, "sat": 6,
	}
	if v, ok := weekdays[strings.ToLower(s)]; ok {
		return v, nil
	}

	// Parse as number
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid value: %s", s)
	}

	if val < min || val > max {
		return 0, fmt.Errorf("value %d out of range [%d-%d]", val, min, max)
	}

	return val, nil
}

// normalizeWeekdays converts weekday 7 to 0 (both mean Sunday)
func normalizeWeekdays(weekdays []int) []int {
	result := make([]int, 0, len(weekdays))
	for _, w := range weekdays {
		if w == 7 {
			w = 0
		}

		result = append(result, w)
	}

	return unique(result)
}

// unique removes duplicates and sorts the slice
func unique(vals []int) []int {
	seen := make(map[int]bool)

	var result []int

	for _, v := range vals {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}

	// Simple bubble sort (slice is typically small)
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i] > result[j] {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}

// generateHumanDescription creates a human-readable description
func generateHumanDescription(minutes, hours, days, months, weekdays []int) string {
	parts := []string{}

	// Time
	if len(minutes) == 60 && len(hours) == 24 {
		parts = append(parts, "Every minute")
	} else if len(minutes) == 60 {
		parts = append(parts, fmt.Sprintf("Every minute during hour(s) %s", formatList(hours)))
	} else if len(hours) == 24 {
		parts = append(parts, fmt.Sprintf("At minute(s) %s of every hour", formatList(minutes)))
	} else {
		parts = append(parts, fmt.Sprintf("At %s", formatTimeList(minutes, hours)))
	}

	// Days
	if len(days) < 31 {
		parts = append(parts, fmt.Sprintf("on day(s) %s of the month", formatList(days)))
	}

	// Months
	if len(months) < 12 {
		parts = append(parts, fmt.Sprintf("in %s", formatMonths(months)))
	}

	// Weekdays
	if len(weekdays) < 7 {
		parts = append(parts, fmt.Sprintf("on %s", formatWeekdays(weekdays)))
	}

	return strings.Join(parts, " ")
}

func formatList(vals []int) string {
	if len(vals) == 0 {
		return "*"
	}

	// Check if it's a continuous range
	if len(vals) > 2 && vals[len(vals)-1]-vals[0] == len(vals)-1 {
		return fmt.Sprintf("%d-%d", vals[0], vals[len(vals)-1])
	}

	strs := make([]string, len(vals))
	for i, v := range vals {
		strs[i] = strconv.Itoa(v)
	}

	return strings.Join(strs, ",")
}

func formatTimeList(minutes, hours []int) string {
	if len(minutes) == 1 && len(hours) == 1 {
		return fmt.Sprintf("%02d:%02d", hours[0], minutes[0])
	}

	times := []string{}

	for _, h := range hours {
		for _, m := range minutes {
			times = append(times, fmt.Sprintf("%02d:%02d", h, m))
		}
	}

	if len(times) > 5 {
		return fmt.Sprintf("%s (and %d more)", strings.Join(times[:5], ", "), len(times)-5)
	}

	return strings.Join(times, ", ")
}

func formatMonths(months []int) string {
	names := []string{
		"", "January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December",
	}

	strs := make([]string, len(months))
	for i, m := range months {
		if m >= 1 && m <= 12 {
			strs[i] = names[m]
		} else {
			strs[i] = strconv.Itoa(m)
		}
	}

	return strings.Join(strs, ", ")
}

func formatWeekdays(weekdays []int) string {
	names := []string{
		"Sunday", "Monday", "Tuesday", "Wednesday",
		"Thursday", "Friday", "Saturday",
	}

	strs := make([]string, len(weekdays))
	for i, w := range weekdays {
		if w >= 0 && w <= 6 {
			strs[i] = names[w]
		} else {
			strs[i] = strconv.Itoa(w)
		}
	}

	return strings.Join(strs, ", ")
}

// getNextRuns calculates the next N run times
func getNextRuns(schedule *Schedule, count int) []string {
	runs := []string{}
	now := time.Now()

	// Simple brute force: check each minute
	t := now.Truncate(time.Minute).Add(time.Minute)

	for len(runs) < count && t.Before(now.AddDate(2, 0, 0)) {
		if matches(schedule, t) {
			runs = append(runs, t.Format("2006-01-02 15:04 (Mon)"))
		}

		t = t.Add(time.Minute)
	}

	return runs
}

// matches checks if a time matches the schedule
func matches(schedule *Schedule, t time.Time) bool {
	minute := t.Minute()
	hour := t.Hour()
	day := t.Day()
	month := int(t.Month())
	weekday := int(t.Weekday())

	return contains(schedule.Minutes, minute) &&
		contains(schedule.Hours, hour) &&
		contains(schedule.Days, day) &&
		contains(schedule.Months, month) &&
		contains(schedule.Weekdays, weekday)
}

func contains(vals []int, v int) bool {
	return slices.Contains(vals, v)
}
