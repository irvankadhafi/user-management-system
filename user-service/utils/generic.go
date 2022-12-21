package utils

import "encoding/json"

// InterfaceBytesToType converts the given interface value to the specified type.
// The interface value is expected to be a []byte representation of the target type.
// If the interface value is nil, the zero value of the target type will be returned.
func InterfaceBytesToType[T any](i any) (out T) {
	if i == nil {
		return
	}
	bt := i.([]byte)

	_ = json.Unmarshal(bt, &out)
	return
}

// Contains tells whether slice A contains x.
func Contains[T comparable](a []T, x T) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}
