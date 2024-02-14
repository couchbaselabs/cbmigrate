package repo

import (
	"context"
	"errors"
	"fmt"
	"github.com/couchbase/gocb/v2"
	"github.com/couchbaselabs/cbmigrate/internal/common"
	"github.com/couchbaselabs/cbmigrate/internal/couchbase/option"
	"github.com/couchbaselabs/cbmigrate/internal/db/couchbase"
	"strings"
	"time"
)

//go:generate mockgen -source=repo.go -destination=../../../testhelper/mock/cb_repo.go -package=mock_test -mock_names=IRepo=MockCouchbaseIRepo IRepo

type IRepo interface {
	Init(uri string, opts *option.Options) error
	GetAllScopes() ([]gocb.ScopeSpec, error)
	CreateScope(name string) error
	CreateCollection(scope, name string) error
	UpsertData(scope, collection string, docs []gocb.BulkOp) error
	CreateIndex(scope, collection string, index common.Index, fieldPath common.IndexFieldPath) error
}

type Repo struct {
	db *couchbase.DB
}

func NewRepo() IRepo {
	return &Repo{
		db: new(couchbase.DB),
	}
}

func (r *Repo) Init(uri string, opts *option.Options) error {
	return r.db.Init(uri, opts)
}

func (r *Repo) GetAllScopes() ([]gocb.ScopeSpec, error) {
	return r.db.Collections().GetAllScopes(&gocb.GetAllScopesOptions{RetryStrategy: gocb.NewBestEffortRetryStrategy(nil)})
}

func (r *Repo) CreateScope(name string) error {
	return r.db.Collections().CreateScope(
		name,
		&gocb.CreateScopeOptions{RetryStrategy: gocb.NewBestEffortRetryStrategy(nil)})
}

func (r *Repo) CreateCollection(scope, name string) error {
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	return r.db.Collections().CreateCollection(
		gocb.CollectionSpec{
			Name:      name,
			ScopeName: scope,
		},
		&gocb.CreateCollectionOptions{
			RetryStrategy: gocb.NewBestEffortRetryStrategy(nil),
			Context:       ctx,
		})
}

func (r *Repo) UpsertData(scope, collection string, docs []gocb.BulkOp) error {
	col := r.db.Scope(scope).Collection(collection)
	return col.Do(docs, nil)
}

func (r *Repo) CreateIndex(scope, collection string, index common.Index, fieldPath common.IndexFieldPath) error {

	query, err := CreateIndexQuery(r.db.Bucket.Name(), scope, collection, index, fieldPath)
	if err != nil {
		return err
	}
	_, err = r.db.Query(query, &gocb.QueryOptions{})
	if err != nil {
		return err
	}
	return nil
}

func CreateIndexQuery(bucket, scope, collection string, index common.Index, fieldPath common.IndexFieldPath) (string, error) {
	var arrFields []common.Key
	for _, key := range index.Keys {
		key.Field = fieldPath.Get(key.Field)
		if strings.Index(key.Field, "[]") > 0 {
			arrFields = append(arrFields, key)
		}
	}
	arrayIndexExp, err := GroupAndCombine(arrFields, index.Sparse)
	if err != nil {
		return "", err
	}
	arrIndex := true
	var fields []string
	for _, key := range index.Keys {
		switch {
		// I am grouping all the array notation fields into single flatten couchbase array index expression
		case strings.Index(key.Field, "[]") > 0:
			if arrIndex {
				fields = append(fields, GenerateCouchbaseArrayIndex(arrayIndexExp))
				arrIndex = false
			}
		default:
			fields = append(fields, getField(key.Field, index.Sparse, key.Order))
		}
	}
	query := fmt.Sprintf(
		"create index %s on `%s`.`%s`.`%s` (%s)",
		index.Name, bucket, scope, collection, strings.Join(fields, ","))
	return query, nil
}

func getField(field string, sparse bool, order int) string {
	includeMissing := "INCLUDE MISSING"
	indexOrder := "ASC"
	if sparse == true {
		includeMissing = ""
	}
	if order == -1 {
		indexOrder = "DESC"
	}
	return fmt.Sprintf("%s %s %s", formatFieldReference(field), indexOrder, includeMissing)
}

// GroupAndCombine array fields are combined because only one array field can be indexed in a compound index
func GroupAndCombine(keys []common.Key, sparse bool) (string, error) {
	// Assuming all keys have the same prefix for simplicity, as demonstrated in the combineStrings function
	var prefix string
	var combined []string

	for _, key := range keys {
		lastIndex := strings.LastIndex(key.Field, "[]")
		if lastIndex != -1 {
			tempPrefix := key.Field[:lastIndex+2]
			if prefix == "" {
				prefix = tempPrefix
			}
			if tempPrefix != prefix {
				return "", errors.New("multiple array reference")
			}

			suffix := key.Field[lastIndex+2:]

			orderSuffix := "ASC"
			if key.Order == -1 {
				orderSuffix = "DESC"
			}

			missingSuffix := ""
			if !sparse {
				missingSuffix = " INCLUDE MISSING"
			}

			combinedPart := fmt.Sprintf("%s %s%s", suffix, orderSuffix, missingSuffix)
			combined = append(combined, combinedPart)
		}
	}

	// Join the combined parts with commas and prepend the prefix
	result := prefix + strings.Join(combined, ",")
	return result, nil
}
