package mongo

import (
	"context"
	"github.com/couchbaselabs/cbmigrate/internal/common"
	"github.com/couchbaselabs/cbmigrate/internal/mongo/repo"
	"github.com/couchbaselabs/cbmigrate/internal/option"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Mongo struct {
	collection string
	db         repo.IRepo
}

func NewMongo(db repo.IRepo) common.ISource {
	return &Mongo{
		db: db,
	}
}

func (m *Mongo) Init(opts *option.Options) error {
	m.collection = opts.MOpts.Collection
	return m.db.Init(opts.MOpts)
}

func (m *Mongo) StreamData(ctx context.Context, mChan chan map[string]interface{}) error {
	defer close(mChan)
	opts := options.Find().SetSort(bson.D{{"_id", 1}})
	cursor, err := m.db.Find(m.collection, ctx, bson.M{}, opts)
	if err != nil {
		panic(err)
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.TODO()) {
		var data bson.M
		err := cursor.Decode(&data)
		if err != nil {
			return err
		}
		mChan <- data
	}
	return nil
}
