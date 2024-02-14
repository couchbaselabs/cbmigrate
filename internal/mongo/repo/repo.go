package repo

//go:generate mockgen -source=repo.go -destination=../../../testhelper/mock/mongo_repo.go -package=mock_test -mock_names=IRepo=MockMongoIRepo,ICursor=MockMongoICursor IRepo ICursor

import (
	"context"
	mongodb "github.com/couchbaselabs/cbmigrate/internal/db/mongo"
	"github.com/couchbaselabs/cbmigrate/internal/mongo/option"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IRepo interface {
	Init(opts *option.Options) error
	Find(collection string, ctx context.Context, filter interface{}, opts ...*options.FindOptions) (ICursor, error)
	GetIndexes(ctx context.Context, collection string) ([]Indexes, error)
}

type ICursor interface {
	Close(ctx context.Context) error
	Next(ctx context.Context) bool
	Decode(val interface{}) error
	Err() error
}

type Indexes bson.M

func (i Indexes) GetName() string {
	name, _ := i["name"].(string)
	return name
}
func (i Indexes) GetKey() map[string]interface{} {
	key, _ := i["key"].(map[string]interface{})
	return key
}

func (i Indexes) IsSparse() bool {
	sparse, _ := i["sparse"].(bool)
	return sparse
}

func (i Indexes) IsText() bool {
	sparse, _ := i["weights"].(bool)
	return sparse
}

func (i Indexes) IsGeoSpatial() bool {
	sparse, _ := i["sparse"].(bool)
	return sparse
}

func (i Indexes) IsTTL() bool {
	_, ok := i["expireAfterSeconds"]
	return ok
}

func (i Indexes) IsCustomCollationEnabled() bool {
	_, ok := i["collation"]
	return ok
}

func (i Indexes) NotSupported() bool {
	if i.IsText() || i.IsGeoSpatial() || i.IsTTL() || i.IsCustomCollationEnabled() {
		return true
	}
	return false
}

type Repo struct {
	db *mongodb.DB
}

func NewRepo() IRepo {
	return &Repo{
		db: new(mongodb.DB),
	}
}

func (r *Repo) Init(opts *option.Options) error {
	return r.db.Init(opts)
}

func (r *Repo) Find(collection string, ctx context.Context, filter interface{}, opts ...*options.FindOptions) (ICursor, error) {
	col := r.db.Collection(collection)
	c, err := col.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	return &Cursor{cursor: c}, nil
}

func (r *Repo) GetIndexes(ctx context.Context, collection string) ([]Indexes, error) {
	col := r.db.Collection(collection)
	cursor, err := col.Indexes().List(ctx)
	if err != nil {
		return nil, err
	}
	var results []Indexes
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}
	return results, nil
}

type Cursor struct {
	cursor *mongo.Cursor
}

func (c *Cursor) Close(ctx context.Context) error {
	return c.cursor.Close(ctx)
}

func (c *Cursor) Next(ctx context.Context) bool {
	return c.cursor.Next(ctx)
}

func (c *Cursor) Err() error {
	return c.cursor.Err()
}

func (c *Cursor) Decode(val interface{}) error {
	return c.cursor.Decode(val)
}
