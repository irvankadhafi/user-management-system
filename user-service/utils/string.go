package utils

import (
	uuid "github.com/satori/go.uuid"
	"strconv"
)

func GenerateID() string {
	id := uuid.NewV4()
	return id.String()
}

// StringToInt :nodoc:
func StringToInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}
