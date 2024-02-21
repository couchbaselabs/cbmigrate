package mongo

import (
	"context"
	"github.com/couchbaselabs/cbmigrate/internal/common"
	"github.com/couchbaselabs/cbmigrate/internal/mongo/option"
	"github.com/couchbaselabs/cbmigrate/internal/mongo/repo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Mongo struct {
	collection string
	db         repo.IRepo
}

func NewMongo(db repo.IRepo) common.ISource[Index, option.Options] {
	return &Mongo{
		db: db,
	}
}

func (m *Mongo) Init(opts *option.Options) error {
	m.collection = opts.Collection
	return m.db.Init(opts)
}

func (m *Mongo) StreamData(ctx context.Context, mChan chan map[string]interface{}) error {
	opts := options.Find().SetSort(bson.D{{"_id", 1}})
	cursor, err := m.db.Find(m.collection, ctx, bson.M{}, opts)
	if err != nil {
		return err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(ctx) {
		var data map[string]interface{}
		err = cursor.Decode(&data)
		if err != nil {
			return err
		}
		mChan <- data
	}
	err = cursor.Err()
	return err
}

func (m *Mongo) GetIndexes(ctx context.Context) ([]Index, error) {
	var indexes []Index
	records, err := m.db.GetIndexes(ctx, m.collection)
	if err != nil {
		return nil, err
	}
	for _, record := range records {

		index := Index{
			Name: record.GetName(),
		}
		if record.NotSupported() {
			continue
		}
		if record.IsSparse() {
			index.Sparse = true
		}
		// adaptive index wild card index need support

		for _, k := range record.GetKey() {
			v, ok := k.Value.(int32)
			if !ok {
				index.NotSupported = true
				break
			}
			index.Keys = append(index.Keys, Key{
				Field: k.Key,
				Order: int(v),
			})
		}
		index.PartialExpression = record.GetPartialExpression()
		indexes = append(indexes, index)
	}
	return indexes, nil
}
