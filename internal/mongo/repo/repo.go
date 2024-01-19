package repo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	mongodb "github.com/couchbaselabs/cbmigrate/internal/db/mongo"
	"github.com/couchbaselabs/cbmigrate/internal/mongo/option"
)

type IRepo interface {
	Init(opts *option.Options) error
	Find(collection string, ctx context.Context, filter interface{}, opts ...*options.FindOptions) (ICursor, error)
}

type ICursor interface {
	Close(ctx context.Context) error
	Next(ctx context.Context) bool
	Decode(val interface{}) error
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

type Cursor struct {
	cursor *mongo.Cursor
}

func (c *Cursor) Close(ctx context.Context) error {
	return c.cursor.Close(ctx)
}

func (c *Cursor) Next(ctx context.Context) bool {
	return c.cursor.Next(ctx)
}

func (c *Cursor) Decode(val interface{}) error {
	return c.cursor.Decode(val)
}
