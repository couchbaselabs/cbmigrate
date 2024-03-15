package common

//go:generate mockgen -source=analyzer.go -destination=../../testhelper/mock/index.go -package=mock_test Analyzer

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"reflect"
	"strings"
)

type Analyzer[T any] interface {
	Init(index []T, suk IDocumentKey)
	AnalyzeData(data map[string]interface{})
	GetCouchbaseQuery(bucket, scope, collection string) []Index
	//GetKeyPathWithArrayNotation(field string) string
}

func isArray(val interface{}) bool {
	switch val.(type) {
	case []interface{}, primitive.A:
		return true
	}
	return false
}

// checkPath is written based mongodb unmarshalling so casting data to []interface{} will not work, so primitive.A is used
// If this function is used for any other database other than mongo modification is needed in isArray function.
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

// ExtractKeys traverses a MongoDB filter expression and collects unique field names.
func ExtractKeys(expression bson.D) []string {
	keysMap := make(map[string]bool)
	collectKeys(expression, keysMap)

	// Convert the map to a slice for the output
	var keys []string
	for key := range keysMap {
		keys = append(keys, key)
	}

	return keys
}

// collectKeys is a recursive helper function to traverse the expression.
func collectKeys(expression bson.D, keysMap map[string]bool) {
	for i := range expression {
		key, value := expression[i].Key, expression[i].Value
		if key == "$and" || key == "$or" {
			switch exprs := value.(type) {
			case bson.A:
				for _, expr := range exprs {
					collectKeys(expr.(bson.D), keysMap)
				}
			case bson.D:
				collectKeys(exprs, keysMap)
			}
		} else {
			// Directly add the key as a field name
			keysMap[key] = true

			// Check if the value is a map indicating a comparison operation
			if _, ok := value.(bson.D); ok {
				// Previously, we had a variable here that was unused.
				// Since we only want to confirm the type, no further action is needed here.
			}
		}
	}
}
