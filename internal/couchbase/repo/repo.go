package repo

import (
	"context"
	"github.com/couchbase/gocb/v2"
	"github.com/couchbaselabs/cbmigrate/internal/couchbase/option"
	"github.com/couchbaselabs/cbmigrate/internal/db/couchbase"
	"time"
)

//go:generate mockgen -source=repo.go -destination=../../../testhelper/mock/cb_repo.go -package=mock -mock_names=IRepo=MockCouchbaseIRepo IRepo

type IRepo interface {
	Init(uri string, opts *option.Options) error
	GetAllScopes() ([]gocb.ScopeSpec, error)
	CreateScope(name string) error
	CreateCollection(scope, name string) error
	UpsertData(scope, collection string, docs []gocb.BulkOp) error
	CreateIndex(query string) error
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

func (r *Repo) CreateIndex(query string) error {
	_, err := r.db.Query(query, &gocb.QueryOptions{})
	if err != nil {
		return err
	}
	return nil
}
