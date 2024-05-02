package mongo

import (
	"context"
	"fmt"
	"github.com/couchbaselabs/cbmigrate/internal/common"
	"github.com/couchbaselabs/cbmigrate/internal/errors"
	"github.com/couchbaselabs/cbmigrate/internal/mongo/option"
	"github.com/couchbaselabs/cbmigrate/internal/mongo/repo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strconv"
)

type Mongo struct {
	collection string
	analyzer   Analyzer
	db         repo.IRepo

	CopyIndexes bool
}

func NewMongo(db repo.IRepo, analyzer Analyzer) common.ISource[option.Options] {
	return &Mongo{
		db:       db,
		analyzer: analyzer,
	}
}

func (m *Mongo) Init(opts *option.Options, documentKey common.IDocumentKey) error {
	m.collection = opts.Collection
	err := m.db.Init(opts)
	if err != nil {
		return err
	}
	m.CopyIndexes = opts.CopyIndexes
	if m.CopyIndexes {
		indexes, err := m.GetIndexes(context.Background())
		if err != nil {
			return err
		}
		m.analyzer.Init(indexes, documentKey)
	}
	return nil
}

func (m *Mongo) analyseData(mChan chan map[string]interface{}) chan map[string]interface{} {
	// a new channel is used for analyzer because analyzing the data after decoding will be blocking.
	analyseChan := make(chan map[string]interface{}, cap(mChan))
	go func() {
		for data := range analyseChan {
			if m.CopyIndexes {
				m.analyzer.AnalyzeData(data)
			}
			mChan <- data
		}
		close(mChan)
	}()
	return analyseChan
}

func (m *Mongo) StreamData(ctx context.Context, mChan chan map[string]interface{}) error {
	analyseChan := m.analyseData(mChan)
	defer close(analyseChan)

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
		analyseChan <- data
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
		if ierr := record.NotSupported(); ierr != nil {
			index.Error = ierr
			indexes = append(indexes, index)
			continue
		}
		if record.IsSparse() {
			index.Sparse = true
		}
		// adaptive index wild card index need support

		for _, k := range record.GetKey() {
			v, err := strconv.Atoi(fmt.Sprintf("%v", k.Value))
			if err != nil {
				index.Error = errors.NewMongoNotSupportedError(fmt.Sprintf("error occured while getting order value %#v", err))
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

func (m *Mongo) GetCouchbaseIndexesQuery(bucket string, scope string, collection string) []common.Index {
	return m.analyzer.GetCouchbaseQuery(bucket, scope, collection)
}
