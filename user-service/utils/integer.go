package utils

import (
	"math/rand"
	"strconv"
	"time"
)

// Int64ToString :nodoc:
func Int64ToString(i int64) string {
	s := strconv.FormatInt(i, 10)
	return s
}

// GenerateID based on current time
func GenerateID() int64 {
	return time.Now().UnixNano() + int64(rand.Intn(10000))
}

// Offset to get offset from page and limit, min value for page = 1
func Offset(page, limit int64) int64 {
	offset := (page - 1) * limit
	if offset < 0 {
		return 0
	}
	return offset
}
