package utils

import (
	"encoding/json"
	"reflect"
)

// ToByte converts any type to a byte slice.
func ToByte(i any) []byte {
	bt, _ := json.Marshal(i)
	return bt
}

// Dump to json using json marshal
func Dump(i any) string {
	return string(ToByte(i))
}

func OmitFields(u any, fields []string) interface{} {
	// Extract type information from interface u
	val := reflect.ValueOf(u)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Iterate over fields in the struct
	for i := 0; i < val.NumField(); i++ {
		// Extract field name
		fieldName := val.Type().Field(i).Name

		// Check if the field name is in the slice fields
		if Contains[string](fields, fieldName) {
			// Omit the field if it is in the slice fields
			field := val.FieldByName(fieldName)
			field.Set(reflect.Zero(field.Type()))
		}
	}

	return u
}

// MapArray mengisi array baru dengan data dari array lama menggunakan fungsi callback
func MapArray(array interface{}, callback func(i int) interface{}) interface{} {
	val := reflect.ValueOf(array)
	newVal := reflect.MakeSlice(val.Type(), val.Len(), val.Len())

	for i := 0; i < val.Len(); i++ {
		newVal.Index(i).Set(reflect.ValueOf(callback(i)))
	}

	return newVal.Interface()
}
