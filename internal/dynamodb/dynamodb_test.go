package dynamodb_test

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	dynamodb2 "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/couchbaselabs/cbmigrate/internal/common"
	"github.com/couchbaselabs/cbmigrate/internal/dynamodb"
	dOpts "github.com/couchbaselabs/cbmigrate/internal/dynamodb/option"
	"github.com/couchbaselabs/cbmigrate/internal/dynamodb/repo"
	mocktest "github.com/couchbaselabs/cbmigrate/testhelper/mock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var repoIndexes = []repo.Index{
	{
		Name: "id",
		Keys: []string{"id"},
	},
}
var _ = Describe("DynamoDB service", func() {
	docKey := common.NewCBDocumentKey()
	Describe("test DynamoDB streaming", func() {
		var (
			ctrl            *gomock.Controller
			db              *mocktest.MockDynamoDbIRepo
			dynamodbService common.ISource[dOpts.Options]
			paginator       *mocktest.MockDynamoDbIPaginator
		)
		opts := &dOpts.Options{
			TableName:   "test1",
			EndpointUrl: "base-url",
			NoSSLVerify: true,
			Profile:     "aws-temp-profile",
			Region:      "us-east-1",
			CABundle:    "path",
			Segments:    1,
		}
		testData := []map[string]interface{}{{"a": 1.0}, {"b": 1.0}, {"c": 1.0}}
		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())
			db = mocktest.NewMockDynamoDbIRepo(ctrl)
			dynamodbService = dynamodb.NewDynamoDB(db)
			paginator = mocktest.NewMockDynamoDbIPaginator(ctrl)
		})
		AfterEach(func() {
			ctrl.Finish()
		})
		Context("success", func() {
			It("output data should match with the test data", func() {
				ctx := context.Background()
				db.EXPECT().Init(opts).Return(nil)
				db.EXPECT().GetPrimaryIndex(context.Background()).Return(repoIndexes[0], nil)
				err := dynamodbService.Init(opts, docKey)
				Expect(err).To(BeNil())
				Expect(docKey.GetKey()).To(Equal([]common.DocumentKeyPart{{Kind: common.DkField, Value: "id"}}))
				db.EXPECT().NewPaginator(int32(0), int32(1), int32(0)).Return(paginator)
				i := -1
				paginator.EXPECT().HasMorePages().Times(4).DoAndReturn(func() bool {
					i++
					return i != 3
				})
				paginator.EXPECT().NextPage(gomock.Any()).Times(3).DoAndReturn(func(ctx context.Context, optFns ...func(*dynamodb2.Options)) (*dynamodb2.ScanOutput, error) {
					var items []map[string]types.AttributeValue
					item, err := attributevalue.MarshalMap(&testData[i])
					if err != nil {
						return nil, err
					}
					items = append(items, item)
					return &dynamodb2.ScanOutput{Items: items}, nil
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
				err = dynamodbService.StreamData(ctx, stream)
				<-doneRoutine
				Expect(err).To(BeNil())
				Ω(outputData).Should(Equal(testData))
			})
		})
		Context("failure", func() {
			It("error in connection initialization", func() {
				dbConInitError := errors.New("error in initializing db connection")
				db.EXPECT().Init(opts).Return(dbConInitError)
				err := dynamodbService.Init(opts, nil)
				Expect(err).NotTo(BeNil())
				Expect(err).To(Equal(dbConInitError))
			})
			It("error in scanning the document", func() {
				dbFindError := errors.New("error in scanning the document")
				db.EXPECT().Init(opts).Return(nil)
				db.EXPECT().GetPrimaryIndex(context.Background()).Return(repoIndexes[0], nil)
				err := dynamodbService.Init(opts, docKey)
				Expect(err).To(BeNil())
				ctx := context.Background()
				db.EXPECT().NewPaginator(int32(0), int32(1), int32(0)).Return(paginator)
				i := -1
				paginator.EXPECT().HasMorePages().Times(2).DoAndReturn(func() bool {
					i++
					if i == 3 {
						return false
					}
					return true
				})
				paginator.EXPECT().NextPage(gomock.Any()).Times(2).DoAndReturn(func(ctx context.Context, optFns ...func(*dynamodb2.Options)) (*dynamodb2.ScanOutput, error) {
					var items []map[string]types.AttributeValue
					item, _ := attributevalue.MarshalMap(&testData[i])
					items = append(items, item)
					if err != nil {
						return nil, err
					}
					if i < 1 {
						return &dynamodb2.ScanOutput{Items: items}, nil
					}
					return nil, dbFindError
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
				err = dynamodbService.StreamData(ctx, stream)
				<-doneRoutine
				Expect(err).NotTo(BeNil())
				Expect(err).To(Equal(dbFindError))
			})
		})
	})
})
