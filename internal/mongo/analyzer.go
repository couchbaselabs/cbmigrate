package mongo

import (
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

func NewIndexFieldAnalyzer() index.Analyzer {
	return &IndexFieldAnalyzer{
		keys: make(map[string]*key),
	}
}

func (a *IndexFieldAnalyzer) Init(indexes []index.Index) {
	for _, i := range indexes {
		for _, key := range i.Keys {
			a.keys[key.Field] = nil
		}
		if i.PartialExpression != nil {
			for _, key := range index.ExtractKeys(i.PartialExpression) {
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

func (a *IndexFieldAnalyzer) GetIndexFieldPath() index.IndexFieldPath {
	var indexKeyAlias = make(index.IndexFieldPath)
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
