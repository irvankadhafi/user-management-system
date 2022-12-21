package utils

import "time"

// FormatTimeRFC3339 formats the given time in the RFC3339Nano format.
// If the time is nil or has zero nanoseconds, the zero value of time.Time is returned.
func FormatTimeRFC3339(t *time.Time) (s string) {
	if t == nil || t.Nanosecond() == 0 {
		return time.Time{}.Format(time.RFC3339Nano)
	}
	return t.Format(time.RFC3339Nano)
}
