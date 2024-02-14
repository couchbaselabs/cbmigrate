package common

import "reflect"

// IsNilOrZero checks if the provided interface{} value is nil,
// a nil pointer, a nil interface, or a zero value of any type.
func IsNilOrZero(i interface{}) bool {
	if i == nil {
		return true // Directly nil
	}

	// Using reflection to further inspect the value
	v := reflect.ValueOf(i)

	// Handle nil pointers and interfaces specifically
	switch v.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		// For these types, use IsNil to check if the value is nil.
		// Note: IsNil panics if called on a value that's not one of these types,
		// hence the type switch to guard against that.
		return v.IsNil()
	case reflect.Struct:
		// Special handling for struct types to check for zero values.
		// Zero struct has zero value for all fields.
		return v.IsZero()
	default:
		// For all other types, use IsZero to determine if it's the zero value for its type.
		return v.IsZero()
	}
}
