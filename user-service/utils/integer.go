package utils

import "strconv"

// Int64ToString :nodoc:
func Int64ToString(i int64) string {
	s := strconv.FormatInt(i, 10)
	return s
}
