package repo

//go:generate mockgen -source=repo.go -destination=../../../testhelper/mock/mongo_repo.go -package=mock -mock_names=IRepo=MockMongoIRepo,ICursor=MockMongoICursor IRepo ICursor

import (
	"context"
	mongodb "github.com/couchbaselabs/cbmigrate/internal/db/mongo"
	"github.com/couchbaselabs/cbmigrate/internal/errors"
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

type Indexes struct {
	Name                    string      `bson:"name"`
	TwoDSphereIndexVersion  interface{} `bson:"2dsphereIndexVersion"`
	Key                     bson.D      `bson:"key"`
	PartialFilterExpression bson.D      `bson:"partialFilterExpression"`
	Sparse                  bool        `bson:"sparse"`
	Weights                 interface{} `bson:"weights"`
	Collation               interface{} `bson:"collation"`
	ExpireAfterSeconds      interface{} `bson:"expireAfterSeconds"`
}

func (i *Indexes) GetName() string {
	return i.Name
}
func (i *Indexes) GetKey() bson.D {
	return i.Key
}

func (i *Indexes) GetPartialExpression() bson.D {
	return i.PartialFilterExpression
}

func (i *Indexes) IsSparse() bool {
	return i.Sparse
}

func (i *Indexes) IsText() bool {
	return i.Weights != nil
}

func (i *Indexes) IsGeoSpatial() bool {
	return i.TwoDSphereIndexVersion != nil
}

func (i *Indexes) IsTTL() bool {
	return i.ExpireAfterSeconds != nil
}

func (i *Indexes) IsCustomCollationEnabled() bool {
	return i.Collation != nil
}

func (i *Indexes) NotSupported() error {
	switch {
	case i.IsText():
		return errors.NewMongoNotSupportedError("text index not supported")
	case i.IsGeoSpatial():
		return errors.NewMongoNotSupportedError("geo spatial index not supported")
	case i.IsCustomCollationEnabled():
		return errors.NewMongoNotSupportedError("index with collation settings not supported")
	}
	return nil
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
