package mongo

//go:generate mockgen -source=analyzer.go -destination=../../testhelper/mock/mongo_analyzer.go -package=mock_test Analyzer

import (
	"fmt"
	"github.com/couchbaselabs/cbmigrate/internal/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"reflect"
	"strings"
)

type Analyzer interface {
	Init(index []Index, documentKey common.ICBDocumentKey)
	AnalyzeData(data map[string]interface{})
	GetCouchbaseQuery(bucket, scope, collection string) []common.Index
	//GetKeyPathWithArrayNotation(field string) string
}

type IndexFieldAnalyzer struct {
	indexes []Index
	keys    map[string]*key
	dk      common.ICBDocumentKey
}

type occurrence int

type key struct {
	keys       map[string]occurrence
	occurrence int
}

func NewIndexFieldAnalyzer() Analyzer {
	return &IndexFieldAnalyzer{
		keys: make(map[string]*key),
	}
}

func (a *IndexFieldAnalyzer) Init(indexes []Index, documentKey common.ICBDocumentKey) {
	a.indexes = indexes
	for _, i := range indexes {
		for _, key := range i.Keys {
			a.keys[key.Field] = nil
		}
		if i.PartialExpression != nil {
			for _, key := range extractKeys(i.PartialExpression) {
				a.keys[key] = nil
			}
		}
	}
	a.dk = documentKey
}

func (a *IndexFieldAnalyzer) AnalyzeData(data map[string]interface{}) {
	for k := range a.keys {
		if a.keys[k] != nil && a.keys[k].occurrence > 100 {
			continue
		}
		path, found := NavigatePath(k, data)
		if found {
			if a.keys[k] == nil {
				a.keys[k] = &key{
					keys: make(map[string]occurrence),
				}
			}
			if _, ok := a.keys[k].keys[path]; !ok {
				a.keys[k].keys[path] = 1
			} else {
				a.keys[k].keys[path]++
			}
			a.keys[k].occurrence++
		}
	}
}

func (a *IndexFieldAnalyzer) getIndexFieldPath() IndexFieldPath {
	var indexKeyAlias = make(IndexFieldPath)
	for field, v := range a.keys {
		f := field
		maxOccurrence := occurrence(0)
		if v == nil {
			continue
		}
		for path, v := range v.keys {
			if v > maxOccurrence {
				f = path
				maxOccurrence = v
			}
		}
		indexKeyAlias[field] = f
	}
	if k := a.dk.GetNonCompoundPrimaryKeyOnly(); k != "" {
		indexKeyAlias[k] = common.MetaDataID
	}
	return indexKeyAlias
}

func (a *IndexFieldAnalyzer) GetCouchbaseQuery(bucket, scope, collection string) []common.Index {
	fieldPath := a.getIndexFieldPath()
	var indexes []common.Index
	isPrimaryIndexPresent := false
	for _, mindex := range a.indexes {
		cindex := common.Index{
			Name: mindex.Name,
		}
		switch {
		case mindex.Error != nil:
			cindex.Error = mindex.Error
		case len(mindex.Keys) == 1 && a.dk.GetNonCompoundPrimaryKeyOnly() == mindex.Keys[0].Field:
			cindex.Query = fmt.Sprintf(
				"CREATE PRIMARY INDEX `%s` on `%s`.`%s`.`%s` USING GSI WITH {\"defer_build\":true}",
				mindex.Name, bucket, scope, collection)
			isPrimaryIndexPresent = true
		default:
			query, err := CreateIndexQuery(bucket, scope, collection, mindex, fieldPath)
			cindex.Query = query
			cindex.Error = err
		}
		indexes = append(indexes, cindex)
	}
	if !isPrimaryIndexPresent {
		uuid, _ := common.GenerateShortUUIDHex()
		key := "primary-" + uuid
		index := common.Index{
			Name: key,
			Query: fmt.Sprintf(
				"CREATE PRIMARY INDEX `%s` on `%s`.`%s`.`%s` USING GSI WITH {\"defer_build\":true}",
				key, bucket, scope, collection),
		}
		indexes = append(indexes, index)
	}
	return indexes
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
func extractKeys(expression bson.D) []string {
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
