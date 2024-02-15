package mongo

import (
	"github.com/couchbaselabs/cbmigrate/internal/common"
	"github.com/couchbaselabs/cbmigrate/internal/index"
)

type IndexFieldAnalyzer struct {
	keys map[string]*key
}

type occurrence int

type key struct {
	keys       map[string]occurrence
	occurrence int
}

func NewIndexFieldAnalyzer() common.Analyzer {
	return &IndexFieldAnalyzer{
		keys: make(map[string]*key),
	}
}

func (a *IndexFieldAnalyzer) Init(indexes []common.Index) {
	for _, index := range indexes {
		for _, key := range index.Keys {
			a.keys[key.Field] = nil
		}
		if index.PartialExpression != nil {
			for _, key := range ExtractKeys(index.PartialExpression) {
				a.keys[key] = nil
			}
		}
	}
}

func (a *IndexFieldAnalyzer) AnalyzeData(data map[string]interface{}) {
	for k := range a.keys {
		if a.keys[k] != nil && a.keys[k].occurrence > 100 {
			continue
		}
		path, found := index.NavigatePath(k, data)
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

func (a *IndexFieldAnalyzer) GetIndexFieldPath() common.IndexFieldPath {
	var indexKeyAlias = make(common.IndexFieldPath)
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
	return indexKeyAlias
}

// ExtractKeys traverses a MongoDB filter expression and collects unique field names.
func ExtractKeys(expression map[string]interface{}) []string {
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
func collectKeys(expression map[string]interface{}, keysMap map[string]bool) {
	for key, value := range expression {
		if key == "$and" || key == "$or" {
			switch exprs := value.(type) {
			case []interface{}:
				for _, expr := range exprs {
					collectKeys(expr.(map[string]interface{}), keysMap)
				}
			case map[string]interface{}:
				collectKeys(exprs, keysMap)
			}
		} else {
			// Directly add the key as a field name
			keysMap[key] = true

			// Check if the value is a map indicating a comparison operation
			if _, ok := value.(map[string]interface{}); ok {
				// Previously, we had a variable here that was unused.
				// Since we only want to confirm the type, no further action is needed here.
			}
		}
	}
}
