package timeutil

import "time"

func Date(layout string) string {
	if layout == "" {
		layout = time.RFC3339
	}
	return time.Now().Format(layout)
}

func Now() time.Time {
	return time.Now()
}
