package common

import (
	"crypto/rand"
	"encoding/hex"
	"reflect"
	"sort"
)

type DocumentKind string

const (
	DkString DocumentKind = "string"
	DkUuid   DocumentKind = "UUID"
	DkField  DocumentKind = "field"
)

const MetaDataID = "meta().id"

type IDocumentKey interface {
	Set(kind DocumentKind, value string)
	Get() string
}

type DocumentKey struct {
	Key          string
	NotComposite bool
	UUIDSet      bool
	HasString    bool
}

func NewDocumentKey() *DocumentKey {
	return &DocumentKey{}
}

func (d *DocumentKey) Set(kind DocumentKind, value string) {
	switch {
	case kind == "UUID":
		d.UUIDSet = true
	case kind == "string":
		d.HasString = true
	case kind == "field":
		if d.Key == "" {
			d.Key = value
		}
		d.NotComposite = true
	}
}

func (d *DocumentKey) Get() string {
	if d.NotComposite && !d.UUIDSet && !d.HasString {
		return d.Key
	}
	return ""
}

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

func GetMapKeys(val map[string]interface{}) []string {
	var keys = make([]string, 0, len(val))
	for k, _ := range val {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func GenerateShortUUIDHex() (string, error) {
	randomBytes := make([]byte, 3) // 8 bytes for a shorter UUID
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}
	shortUUIDHex := hex.EncodeToString(randomBytes)
	return shortUUIDHex, nil
}
