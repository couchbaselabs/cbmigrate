package migrater_test

import (
	"context"
	"errors"
	"github.com/couchbaselabs/cbmigrate/internal/common"
	cOpts "github.com/couchbaselabs/cbmigrate/internal/couchbase/option"
	migrater2 "github.com/couchbaselabs/cbmigrate/internal/migrater"
	"github.com/couchbaselabs/cbmigrate/internal/mongo"
	mOpts "github.com/couchbaselabs/cbmigrate/internal/mongo/option"
	mock_test "github.com/couchbaselabs/cbmigrate/testhelper/mock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/mock/gomock"
	"reflect"
)

var index1 = mongo.Index{
	Name: "index1",
	Keys: []mongo.Key{
		{Field: "k1.n1k1", Order: 1},
		{Field: "k2.n1k1", Order: 1},
		{Field: "k3.n1k1", Order: 1},
	},
	PartialExpression: bson.D{
		{
			Key:   "k1.n1k1",
			Value: "10",
		},
		{
			Key:   "k2.n1k1",
			Value: bson.D{{Key: "$gte", Value: 100}},
		},
	},
}
var index2 = mongo.Index{
	Name: "index2",
	Keys: []mongo.Key{
		{Field: "k4.n1k1", Order: 1},
		{Field: "k5.n1k1", Order: 1},
		{Field: "k6.n1k1", Order: 1},
	},
	PartialExpression: bson.D{
		{
			Key:   "k4.n1k1",
			Value: "10",
		},
		{
			Key:   "k5.n1k1",
			Value: bson.D{{Key: "$gte", Value: 100}},
		},
	},
}
var index3 = mongo.Index{
	Name: "index3",
	Keys: []mongo.Key{
		{Field: "k7.n1k1", Order: 1},
		{Field: "k8.n1k1", Order: 1},
		{Field: "k9.n1k1", Order: 1},
	},
	PartialExpression: bson.D{
		{
			Key:   "k7.n1k1",
			Value: "10",
		},
		{
			Key:   "k8.n1k1",
			Value: bson.D{{Key: "$gte", Value: 100}},
		},
	},
}
var indexes = []mongo.Index{
	index1,
	index2,
	index3,
}

var dk = &common.DocumentKey{}

var cIndexes = []common.Index{
	{
		Name:  "index1",
		Query: "CREATE INDEX idx_airport_over1000\n  ON `travel-sample`.inventory.airport(geo.alt)\n  WHERE geo.alt > 1000",
	},
}

var _ = Describe("migrate", func() {
	dk.Set(common.DkField, "_id")
	Describe("test data migration", func() {
		var (
			ctrl        *gomock.Controller
			source      *mock_test.MockISource[mOpts.Options]
			destination *mock_test.MockIDestination
			migrater    migrater2.IMigrate[mOpts.Options]
		)
		CBOpts := &cOpts.Options{
			Cluster:   "cluster-url",
			NameSpace: &cOpts.NameSpace{Bucket: "test_bucket", Scope: "test_scope", Collection: "test_col"},
			BatchSize: 100,
		}
		MOpts := &mOpts.Options{Namespace: &mOpts.Namespace{Collection: "test_col"}}

		testData := []map[string]interface{}{{"a": 1}, {"b": 2}, {"c": 3}, {"d": 4}}
		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())
			source = mock_test.NewMockISource[mOpts.Options](ctrl)
			destination = mock_test.NewMockIDestination(ctrl)
			migrater = migrater2.NewMigrator[mOpts.Options](source, destination)
		})
		AfterEach(func() {
			ctrl.Finish()
		})
		Context("success", func() {
			It("data copied to destination", func() {
				destination.EXPECT().Init(CBOpts).Return(dk, nil)
				source.EXPECT().Init(MOpts, dk).Return(nil)
				source.EXPECT().StreamData(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, stream chan map[string]interface{}) error {
					for _, d := range testData {
						stream <- d
					}
					close(stream)
					return nil
				})
				i := 0
				destination.EXPECT().ProcessData(gomock.Any()).Times(4).DoAndReturn(func(doc map[string]interface{}) error {
					if !reflect.DeepEqual(doc, testData[i]) {
						return errors.New("process data don't match with source data")
					}
					i++
					return nil
				})
				destination.EXPECT().Complete().Return(nil)
				source.EXPECT().GetCouchbaseIndexesQuery(CBOpts.Bucket, CBOpts.Scope, CBOpts.Collection).Return(cIndexes)
				destination.EXPECT().CreateIndexes(cIndexes).Return(nil)
				err := migrater.Copy(MOpts, CBOpts, true, 10000)
				Expect(err).To(BeNil())
			})
		})
		Context("failure", func() {
			It("source connection initialization error", func() {
				destination.EXPECT().Init(CBOpts).Return(dk, nil)
				sourceError := errors.New("error occurred in source connection initialization")
				source.EXPECT().Init(MOpts, dk).Return(sourceError)
				err := migrater.Copy(MOpts, CBOpts, false, 10000)
				Expect(err).To(Equal(sourceError))
			})
			It("destination connection initialization error", func() {
				destError := errors.New("error occurred in source connection initialization")
				destination.EXPECT().Init(CBOpts).Return(nil, destError)
				err := migrater.Copy(MOpts, CBOpts, false, 10000)
				Expect(err).To(Equal(destError))
			})
			It("error while streaming the data", func() {
				streamError := errors.New("error occurred while streaming the data")
				destination.EXPECT().Init(CBOpts).Return(dk, nil)
				source.EXPECT().Init(MOpts, dk).Return(nil)
				source.EXPECT().StreamData(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, stream chan map[string]interface{}) error {
					for _, d := range testData[0:2] {
						stream <- d
					}
					close(stream)
					return streamError
				})
				i := 0
				destination.EXPECT().ProcessData(gomock.Any()).Times(2).DoAndReturn(func(doc map[string]interface{}) error {
					if !reflect.DeepEqual(doc, testData[i]) {
						return errors.New("process data don't match with source data")
					}
					i++
					return nil
				})
				destination.EXPECT().Complete().Return(nil)
				err := migrater.Copy(MOpts, CBOpts, true, 10000)
				Expect(err).To(Equal(errors.Join(streamError)))
			})

			It("error while processing the data", func() {
				dataProcessError := errors.New("error occurred while processing the data")
				contextCancelledError := errors.New("context cancelled error")
				destination.EXPECT().Init(CBOpts).Return(dk, nil)
				source.EXPECT().Init(MOpts, dk).Return(nil)
				source.EXPECT().StreamData(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, stream chan map[string]interface{}) error {
					defer close(stream)
					for _, d := range testData {
						stream <- d
					}
					<-ctx.Done()
					return contextCancelledError
				})
				i := 0
				destination.EXPECT().ProcessData(gomock.Any()).Times(2).DoAndReturn(func(doc map[string]interface{}) error {
					if !reflect.DeepEqual(doc, testData[i]) {
						return errors.New("process data don't match with source data")
					}
					if i == 1 {
						return dataProcessError
					}
					i++
					return nil
				})
				err := migrater.Copy(MOpts, CBOpts, true, 10000)
				Expect(err.Error()).To(Equal(errors.Join(dataProcessError, contextCancelledError).Error()))
			})
		})
	})
})
