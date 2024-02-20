package migrater_test

import (
	"context"
	"errors"
	cOpts "github.com/couchbaselabs/cbmigrate/internal/couchbase/option"
	"github.com/couchbaselabs/cbmigrate/internal/index"
	migrater2 "github.com/couchbaselabs/cbmigrate/internal/migrater"
	mOpts "github.com/couchbaselabs/cbmigrate/internal/mongo/option"
	"github.com/couchbaselabs/cbmigrate/internal/option"
	mock_test "github.com/couchbaselabs/cbmigrate/testhelper/mock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	"reflect"
)

var index1 = index.Index{
	Name: "index1",
	Keys: []index.Key{
		{Field: "k1.n1k1", Order: 1},
		{Field: "k2.n1k1", Order: 1},
		{Field: "k3.n1k1", Order: 1},
	},
	PartialExpression: map[string]interface{}{
		"k1.n1k1": "10",
		"k2.n1k1": map[string]interface{}{
			"$gte": 100,
		},
	},
}
var index2 = index.Index{
	Name: "index2",
	Keys: []index.Key{
		{Field: "k4.n1k1", Order: 1},
		{Field: "k5.n1k1", Order: 1},
		{Field: "k6.n1k1", Order: 1},
	},
	PartialExpression: map[string]interface{}{
		"k4.n1k1": "10",
		"k5.n1k1": map[string]interface{}{
			"$gte": 100,
		},
	},
}
var index3 = index.Index{
	Name: "index3",
	Keys: []index.Key{
		{Field: "k7.n1k1", Order: 1},
		{Field: "k8.n1k1", Order: 1},
		{Field: "k9.n1k1", Order: 1},
	},
	PartialExpression: map[string]interface{}{
		"k7.n1k1": "10",
		"k8.n1k1": map[string]interface{}{
			"$gte": 100,
		},
	},
}
var indexes = []index.Index{
	index1,
	index2,
	index3,
}

var fieldPath = index.IndexFieldPath{
	"k1.n1k1": "k1[].n1k1",
}

var _ = Describe("migrate", func() {
	Describe("test data migration", func() {
		var (
			ctrl        *gomock.Controller
			source      *mock_test.MockISource
			destination *mock_test.MockIDestination
			analyzer    *mock_test.MockAnalyzer
			migrater    migrater2.IMigrate
		)
		opts := &option.Options{
			CBOpts: &cOpts.Options{
				Cluster:   "cluster-url",
				NameSpace: &cOpts.NameSpace{Bucket: "test_bucket", Scope: "test_scope", Collection: "test_col"},
				BatchSize: 100,
			},
			MOpts: &mOpts.Options{Namespace: &mOpts.Namespace{Collection: "test_col"}},
		}
		testData := []map[string]interface{}{{"a": 1}, {"b": 2}, {"c": 3}, {"d": 4}}
		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())
			source = mock_test.NewMockISource(ctrl)
			destination = mock_test.NewMockIDestination(ctrl)
			analyzer = mock_test.NewMockAnalyzer(ctrl)
			migrater = migrater2.NewMigrator(source, destination, analyzer)
		})
		AfterEach(func() {
			ctrl.Finish()
		})
		Context("success", func() {
			It("data copied to destination", func() {
				source.EXPECT().Init(opts).Return(nil)
				destination.EXPECT().Init(opts).Return(nil)
				source.EXPECT().GetIndexes(gomock.Any()).Return(indexes, nil)
				analyzer.EXPECT().Init(indexes).Return()
				source.EXPECT().StreamData(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, stream chan map[string]interface{}) error {
					for _, d := range testData {
						stream <- d
					}
					return nil
				})
				i := 0
				analyzer.EXPECT().AnalyzeData(gomock.Any()).Times(4).DoAndReturn(func(doc map[string]interface{}) {
					if !reflect.DeepEqual(doc, testData[i]) {
						panic(errors.New("process data don't match with source data"))
					}
				})
				destination.EXPECT().ProcessData(gomock.Any()).Times(4).DoAndReturn(func(doc map[string]interface{}) error {
					if !reflect.DeepEqual(doc, testData[i]) {
						return errors.New("process data don't match with source data")
					}
					i++
					return nil
				})
				destination.EXPECT().Complete().Return(nil)
				analyzer.EXPECT().GetIndexFieldPath().Return(fieldPath)
				destination.EXPECT().CreateIndexes(indexes, fieldPath).Return(nil)
				err := migrater.Copy(opts)
				Expect(err).To(BeNil())
			})
		})
		Context("failure", func() {
			It("source connection initialization error", func() {
				sourceError := errors.New("error occurred in source connection initialization")
				source.EXPECT().Init(opts).Return(sourceError)
				err := migrater.Copy(opts)
				Expect(err).To(Equal(sourceError))
			})
			It("destination connection initialization error", func() {
				destError := errors.New("error occurred in source connection initialization")
				source.EXPECT().Init(opts).Return(nil)
				destination.EXPECT().Init(opts).Return(destError)
				err := migrater.Copy(opts)
				Expect(err).To(Equal(destError))
			})
			It("error occurred while getting the indexes", func() {
				indexError := errors.New("error occurred while getting the indexes")
				source.EXPECT().Init(opts).Return(nil)
				destination.EXPECT().Init(opts).Return(nil)
				source.EXPECT().GetIndexes(gomock.Any()).Return(nil, indexError)
				err := migrater.Copy(opts)
				Expect(err).To(Equal(indexError))
			})
			It("error while streaming the data", func() {
				streamError := errors.New("error occurred while streaming the data")
				source.EXPECT().Init(opts).Return(nil)
				destination.EXPECT().Init(opts).Return(nil)
				source.EXPECT().GetIndexes(gomock.Any()).Return(indexes, nil)
				analyzer.EXPECT().Init(indexes).Return()
				source.EXPECT().StreamData(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, stream chan map[string]interface{}) error {
					for _, d := range testData[0:2] {
						stream <- d
					}
					return streamError
				})
				i := 0
				analyzer.EXPECT().AnalyzeData(gomock.Any()).Times(2).DoAndReturn(func(doc map[string]interface{}) {
					if !reflect.DeepEqual(doc, testData[i]) {
						panic(errors.New("process data don't match with source data"))
					}
				})
				destination.EXPECT().ProcessData(gomock.Any()).Times(2).DoAndReturn(func(doc map[string]interface{}) error {
					if !reflect.DeepEqual(doc, testData[i]) {
						return errors.New("process data don't match with source data")
					}
					i++
					return nil
				})
				destination.EXPECT().Complete().Return(nil)
				err := migrater.Copy(opts)
				Expect(err).To(Equal(errors.Join(streamError)))
			})

			It("error while processing the data", func() {
				dataProcessError := errors.New("error occurred while processing the data")
				contextCancelledError := errors.New("context cancelled error")
				source.EXPECT().Init(opts).Return(nil)
				destination.EXPECT().Init(opts).Return(nil)
				source.EXPECT().GetIndexes(gomock.Any()).Return(indexes, nil)
				analyzer.EXPECT().Init(indexes).Return()
				source.EXPECT().StreamData(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, stream chan map[string]interface{}) error {
					defer GinkgoRecover()
					for _, d := range testData {
						stream <- d
					}
					<-ctx.Done()
					return contextCancelledError
				})
				i := 0
				analyzer.EXPECT().AnalyzeData(gomock.Any()).Times(2).DoAndReturn(func(doc map[string]interface{}) {
					if !reflect.DeepEqual(doc, testData[i]) {
						panic(errors.New("process data don't match with source data"))
					}
				})
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
				err := migrater.Copy(opts)
				Expect(err.Error()).To(Equal(errors.Join(dataProcessError, contextCancelledError).Error()))
			})

		})
	})
})
