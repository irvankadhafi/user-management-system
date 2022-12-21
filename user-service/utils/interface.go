package utils

import "encoding/json"

// ToByte converts any type to a byte slice.
func ToByte(i any) []byte {
	bt, _ := json.Marshal(i)
	return bt
}

// Dump to json using json marshal
func Dump(i any) string {
	return string(ToByte(i))
}
