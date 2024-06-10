package mongo_test

import (
	"context"
	"errors"
	"github.com/couchbaselabs/cbmigrate/internal/mongo/repo"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/mock/gomock"
	"reflect"

	"github.com/couchbaselabs/cbmigrate/internal/common"
	"github.com/couchbaselabs/cbmigrate/internal/mongo"
	mOpts "github.com/couchbaselabs/cbmigrate/internal/mongo/option"
	mocktest "github.com/couchbaselabs/cbmigrate/testhelper/mock"
)

var repoIndexes = []repo.Indexes{
	{
		Name: "_id",
		Key:  bson.D{{Key: "_id", Value: 1}},
	},
}
var mongoIndexes = []mongo.Index{
	{
		Name: "_id",
		Keys: []mongo.Key{{Field: "_id", Order: 1}},
	},
}
var _ = Describe("mongo service", func() {
	Describe("test mongoService streaming", func() {
		var (
			ctrl         *gomock.Controller
			db           *mocktest.MockMongoIRepo
			analyzer     *mocktest.MockAnalyzer
			cursor       *mocktest.MockMongoICursor
			mongoService common.ISource[mOpts.Options]
		)
		opts := &mOpts.Options{Namespace: &mOpts.Namespace{Collection: "test_col"}, CopyIndexes: true}
		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())
			db = mocktest.NewMockMongoIRepo(ctrl)
			analyzer = mocktest.NewMockAnalyzer(ctrl)
			cursor = mocktest.NewMockMongoICursor(ctrl)
			mongoService = mongo.NewMongo(db, analyzer)
		})
		AfterEach(func() {
			ctrl.Finish()
		})
		Context("success", func() {
			It("output data should match with the test data", func() {
				testData := []map[string]interface{}{{"a": 1}, {"b": 1}}
				ctx := context.Background()
				db.EXPECT().Init(opts).Return(nil)
				db.EXPECT().GetIndexes(context.Background(), opts.Collection).Return(repoIndexes, nil)
				analyzer.EXPECT().Init(mongoIndexes, nil).Return()
				err := mongoService.Init(opts, nil)
				Expect(err).To(BeNil())

				db.EXPECT().Find(opts.Collection, ctx, bson.M{}, gomock.Any()).Return(cursor, nil)
				cursor.EXPECT().Close(ctx).Return(nil)
				dataCount := len(testData)
				n := -1
				cursor.EXPECT().Next(ctx).Times(dataCount + 1).DoAndReturn(func(ctx context.Context) bool {
					n++
					if n >= dataCount {
						return false
					}
					return true
				})
				cursor.EXPECT().Decode(gomock.Any()).Times(dataCount).DoAndReturn(func(val interface{}) error {
					reflect.ValueOf(val).Elem().Set(reflect.ValueOf(testData[n]))
					return nil
				})
				cursor.EXPECT().Err().Return(nil)

				analyzer.EXPECT().AnalyzeData(gomock.Any()).Times(dataCount).DoAndReturn(func(data map[string]interface{}) {
					reflect.DeepEqual(data, testData[n])
				})

				stream := make(chan map[string]interface{})
				var outputData []map[string]interface{}
				doneRoutine := make(chan bool)
				go func() {
					for data := range stream {
						outputData = append(outputData, data)
					}
					doneRoutine <- true
				}()
				err = mongoService.StreamData(ctx, stream)
				<-doneRoutine

				Expect(err).To(BeNil())
				Ω(outputData).Should(Equal(testData))
			})

			It("output data should match with the test data (without index copy)", func() {
				opts := &mOpts.Options{Namespace: &mOpts.Namespace{Collection: "test_col"}}
				testData := []map[string]interface{}{{"a": 1}, {"b": 1}}
				ctx := context.Background()
				db.EXPECT().Init(opts).Return(nil)
				err := mongoService.Init(opts, nil)
				Expect(err).To(BeNil())

				db.EXPECT().Find(opts.Collection, ctx, bson.M{}, gomock.Any()).Return(cursor, nil)
				cursor.EXPECT().Close(ctx).Return(nil)
				dataCount := len(testData)
				n := -1
				cursor.EXPECT().Next(ctx).Times(dataCount + 1).DoAndReturn(func(ctx context.Context) bool {
					n++
					if n >= dataCount {
						return false
					}
					return true
				})
				cursor.EXPECT().Decode(gomock.Any()).Times(dataCount).DoAndReturn(func(val interface{}) error {
					reflect.ValueOf(val).Elem().Set(reflect.ValueOf(testData[n]))
					return nil
				})
				cursor.EXPECT().Err().Return(nil)

				stream := make(chan map[string]interface{})
				var outputData []map[string]interface{}
				doneRoutine := make(chan bool)
				go func() {
					for data := range stream {
						outputData = append(outputData, data)
					}
					doneRoutine <- true
				}()
				err = mongoService.StreamData(ctx, stream)
				<-doneRoutine

				Expect(err).To(BeNil())
				Ω(outputData).Should(Equal(testData))
			})

		})
		Context("failure", func() {
			It("error in connection initialization", func() {
				dbConInitError := errors.New("error in initializing db connection")
				db.EXPECT().Init(opts).Return(dbConInitError)
				err := mongoService.Init(opts, nil)
				Expect(err).NotTo(BeNil())
				Expect(err).To(Equal(dbConInitError))
			})
			It("error in find the document", func() {
				dbFindError := errors.New("error in finding the document")
				db.EXPECT().Init(opts).Return(nil)
				db.EXPECT().GetIndexes(context.Background(), opts.Collection).Return(repoIndexes, nil)
				analyzer.EXPECT().Init(mongoIndexes, nil).Return()
				err := mongoService.Init(opts, nil)
				Expect(err).To(BeNil())
				ctx := context.Background()
				db.EXPECT().Find(opts.Collection, ctx, bson.M{}, gomock.Any()).Return(nil, dbFindError)
				stream := make(chan map[string]interface{})
				err = mongoService.StreamData(ctx, stream)
				Expect(err).NotTo(BeNil())
				Expect(err).To(Equal(dbFindError))
			})
			It("error in decoding the document", func() {
				testData := []map[string]interface{}{{"a": 1}, {"b": 1}}
				decodeError := errors.New("error in decoding the document")
				db.EXPECT().Init(opts).Return(nil)
				db.EXPECT().GetIndexes(context.Background(), opts.Collection).Return(repoIndexes, nil)
				analyzer.EXPECT().Init(mongoIndexes, nil).Return()
				err := mongoService.Init(opts, nil)
				Expect(err).To(BeNil())
				ctx := context.Background()
				db.EXPECT().Find(opts.Collection, ctx, bson.M{}, gomock.Any()).Return(cursor, nil)
				cursor.EXPECT().Next(ctx).Times(2).DoAndReturn(func(ctx context.Context) bool {
					return true
				})
				ci := 0
				cursor.EXPECT().Decode(gomock.Any()).Times(2).DoAndReturn(func(val interface{}) error {
					if ci == 1 {
						return decodeError
					}
					reflect.ValueOf(val).Elem().Set(reflect.ValueOf(testData[ci]))
					ci++
					return nil
				})

				analyzer.EXPECT().AnalyzeData(gomock.Any()).Times(1).DoAndReturn(func(data map[string]interface{}) {
					reflect.DeepEqual(data, testData[ci])
				})

				cursor.EXPECT().Close(context.Background()).Return(nil)
				stream := make(chan map[string]interface{})
				var outputData []map[string]interface{}
				doneRoutine := make(chan bool)
				go func() {
					for data := range stream {
						outputData = append(outputData, data)
					}
					doneRoutine <- true
				}()
				err = mongoService.StreamData(ctx, stream)
				<-doneRoutine
				Expect(err).NotTo(BeNil())
				Expect(err).To(Equal(decodeError))
				Ω(outputData).Should(Equal([]map[string]interface{}{testData[0]}))
			})
			It("error in cursor", func() {
				testData := []map[string]interface{}{{"a": 1}, {"b": 1}}
				cursorError := errors.New("error in cursor")
				db.EXPECT().Init(opts).Return(nil)
				db.EXPECT().GetIndexes(context.Background(), opts.Collection).Return(repoIndexes, nil)
				analyzer.EXPECT().Init(mongoIndexes, nil).Return()
				err := mongoService.Init(opts, nil)
				Expect(err).To(BeNil())
				ctx := context.Background()
				db.EXPECT().Find(opts.Collection, ctx, bson.M{}, gomock.Any()).Return(cursor, nil)
				ci := 0
				cursor.EXPECT().Next(ctx).Times(2).DoAndReturn(func(ctx context.Context) bool {
					if ci == 1 {
						return false
					}
					ci++
					return true
				})
				cj := 0
				cursor.EXPECT().Decode(gomock.Any()).Times(1).DoAndReturn(func(val interface{}) error {
					reflect.ValueOf(val).Elem().Set(reflect.ValueOf(testData[cj]))
					cj++
					return nil
				})

				analyzer.EXPECT().AnalyzeData(gomock.Any()).Times(1).DoAndReturn(func(data map[string]interface{}) {
					reflect.DeepEqual(data, testData[cj])
				})
				cursor.EXPECT().Err().Return(cursorError)
				cursor.EXPECT().Close(context.Background()).Return(nil)
				stream := make(chan map[string]interface{})
				var outputData []map[string]interface{}
				doneRoutine := make(chan bool)
				go func() {
					for data := range stream {
						outputData = append(outputData, data)
					}
					doneRoutine <- true
				}()
				err = mongoService.StreamData(ctx, stream)
				<-doneRoutine
				Expect(err).NotTo(BeNil())
				Expect(err).To(Equal(cursorError))
				Ω(outputData).Should(Equal([]map[string]interface{}{testData[0]}))
			})
		})
	})
})
