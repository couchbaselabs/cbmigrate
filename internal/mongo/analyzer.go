package mongo

import (
	"fmt"
	"github.com/couchbaselabs/cbmigrate/internal/common"
)

type IndexFieldAnalyzer struct {
	indexes []Index
	keys    map[string]*key
	dk      *common.DocumentKey
}

type occurrence int

type key struct {
	keys       map[string]occurrence
	occurrence int
}

func NewIndexFieldAnalyzer() common.Analyzer[Index] {
	return &IndexFieldAnalyzer{
		keys: make(map[string]*key),
	}
}

func (a *IndexFieldAnalyzer) Init(indexes []Index, dk *common.DocumentKey) {
	a.indexes = indexes
	for _, i := range indexes {
		for _, key := range i.Keys {
			a.keys[key.Field] = nil
		}
		if i.PartialExpression != nil {
			for _, key := range common.ExtractKeys(i.PartialExpression) {
				a.keys[key] = nil
			}
		}
	}
	a.dk = dk
}

func (a *IndexFieldAnalyzer) AnalyzeData(data map[string]interface{}) {
	for k := range a.keys {
		if a.keys[k] != nil && a.keys[k].occurrence > 100 {
			continue
		}
		path, found := common.NavigatePath(k, data)
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

func (a *IndexFieldAnalyzer) GetIndexFieldPath() IndexFieldPath {
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
	if k := a.dk.Get(); k != "" {
		indexKeyAlias[k] = common.MetaDataID
	}
	return indexKeyAlias
}

func (a *IndexFieldAnalyzer) GetCouchbaseQuery(bucket, scope, collection string) []common.Index {
	fieldPath := a.GetIndexFieldPath()
	var indexes []common.Index
	isPrimaryIndexPresent := false
	for _, mindex := range a.indexes {
		cindex := common.Index{
			Name: mindex.Name,
		}
		switch {
		case mindex.Error != nil:
			cindex.Error = mindex.Error
		case len(mindex.Keys) == 1 && a.dk.Get() == mindex.Keys[0].Field:
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
