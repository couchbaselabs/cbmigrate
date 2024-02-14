package index

import (
	"reflect"
	"strings"
)

type Analyzer interface {
	Analyze(data map[string]interface{})
}

// isArray checks if the provided value is an array or a slice.
func isArray(value interface{}) bool {
	typeOf := reflect.TypeOf(value)
	if typeOf == nil {
		return false
	}
	kind := reflect.TypeOf(value).Kind()
	return kind == reflect.Slice || kind == reflect.Array
}

// checkPath recursively checks if the path exists in the given data and builds the path string.
func checkPath(data interface{}, elements []string, index int, path *strings.Builder) bool {
	if index >= len(elements) {
		if isArray(data) {
			path.WriteString("[]")
		}
		return true
	}

	element := elements[index]

	if isArray(data) {
		slice := reflect.ValueOf(data)
		found := false
		for i := 0; i < slice.Len(); i++ {
			item := slice.Index(i).Interface()
			if itemMap, ok := item.(map[string]interface{}); ok {
				if _, exists := itemMap[element]; exists {
					path.WriteString("[]")
					found = checkPath(itemMap, elements, index, path)
					break
				}
			}
		}
		return found
	} else if currentMap, ok := data.(map[string]interface{}); ok {
		if next, exists := currentMap[element]; exists {
			if index > 0 {
				path.WriteString(".")
			}
			path.WriteString(element)
			return checkPath(next, elements, index+1, path)
		}
	}

	return false
}

// NavigatePath initiates the recursive path check.
func NavigatePath(path string, data map[string]interface{}) (string, bool) {
	elements := strings.Split(path, ".")
	var result strings.Builder

	if len(elements) == 0 {
		return "", false
	}

	if ok := checkPath(data, elements, 0, &result); ok {
		return result.String(), true
	}
	return "", false
}
