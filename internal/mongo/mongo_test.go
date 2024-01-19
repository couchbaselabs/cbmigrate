package mongo_test

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/mock/gomock"
	"reflect"

	"github.com/couchbaselabs/cbmigrate/internal/common"
	"github.com/couchbaselabs/cbmigrate/internal/mongo"
	mOpts "github.com/couchbaselabs/cbmigrate/internal/mongo/option"
	"github.com/couchbaselabs/cbmigrate/internal/option"
	mock_test "github.com/couchbaselabs/cbmigrate/testhelper/mock"
)

var _ = Describe("mongo service", func() {
	Describe("test mongoService streaming", func() {
		var (
			ctrl         *gomock.Controller
			db           *mock_test.MockMongoIRepo
			cursor       *mock_test.MockMongoICursor
			mongoService common.ISource
		)
		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())
			db = mock_test.NewMockMongoIRepo(ctrl)
			cursor = mock_test.NewMockMongoICursor(ctrl)
			mongoService = mongo.NewMongo(db)
		})
		AfterEach(func() {
			ctrl.Finish()
		})
		Context("success", func() {
			It("output data should match with the test data", func() {
				testData := []map[string]interface{}{{"a": 1}, {"b": 1}}
				ctx := context.Background()
				opts := &option.Options{
					MOpts: &mOpts.Options{Namespace: &mOpts.Namespace{Collection: "test_col"}},
				}
				db.EXPECT().Init(opts.MOpts).Return(nil)
				err := mongoService.Init(opts)
				Expect(err).To(BeNil())

				db.EXPECT().Find(opts.MOpts.Collection, ctx, bson.M{}, gomock.Any()).Return(cursor, nil)
				cursor.EXPECT().Close(ctx).Return(nil)
				dataCount := len(testData)
				n := -1
				cursor.EXPECT().Next(context.TODO()).Times(dataCount + 1).DoAndReturn(func(ctx context.Context) bool {
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
	})
})
