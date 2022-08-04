package format

import "time"

// GenerateISOString generates a time string equivalent to Date.now().toISOString in JavaScript
func GenerateISOString(dt time.Time) string {
	return dt.Format("2006-01-02T15:04:05.999Z07:00")
}
