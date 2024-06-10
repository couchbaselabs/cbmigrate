package dynamodb_test

import (
	_ "embed"
	"github.com/couchbaselabs/cbmigrate/cmd/common"
	"github.com/couchbaselabs/cbmigrate/cmd/dynamodb"
	"github.com/couchbaselabs/cbmigrate/cmd/dynamodb/command"
	"github.com/couchbaselabs/cbmigrate/internal/couchbase/option"
	dOpts "github.com/couchbaselabs/cbmigrate/internal/dynamodb/option"
	mocktest "github.com/couchbaselabs/cbmigrate/testhelper/mock"
	"github.com/spf13/cobra"
	"go.uber.org/mock/gomock"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type Integer int

func (i *Integer) String() string {
	return strconv.Itoa(int(*i))
}
func (i *Integer) Int() int {
	return int(*i)
}

var _ = Describe("mongo", func() {

	Describe("mongo command", func() {
		dynamoDBTableNameOption := "--" + command.DynamoDBTableName
		dynamoDBEndpointURLOption := "--" + command.DynamoDBEndpointURL
		dynamoDBProfileOption := "--" + command.DynamoDBProfile
		dynamoDBAccessKeyOption := "--" + command.DynamoDBAccessKey
		dynamoDBSecretKeyOption := "--" + command.DynamoDBSecretKey
		dynamoDBRegionOption := "--" + command.DynamoDBRegion
		dynamoDBCaBundleOption := "--" + command.DynamoDBCaBundle
		dynamoDBNoVerifySSLOption := "--" + command.DynamoDBNoVerifySSL

		cbClusterOption := "--" + common.CBCluster
		cbUserOption := "--" + common.CBUsername
		cbPasswordOption := "--" + common.CBPassword
		cbBucketOption := "--" + common.CBBucket
		cbScopeOption := "--" + common.CBScope
		cbCollectionOption := "--" + common.CBCollection
		cbBatchSizeOption := "--" + common.CBBatchSize
		cbGeneratorKeyOption := "--" + common.CBGenerateKey

		bufferSizeOption := "--" + common.BufferSize

		dynamoDBTableName := "test-table"
		dynamoDBEndpointURL := "aws-endpoint"
		dynamoDBProfile := "aws-profile"
		dynamoDBAccessKey := "access-key"
		dynamoDBSecretKey := "secret-key"
		dynamoDBRegion := "us-east-1"
		dynamoDBCaBundle := "path"

		cbCluster := "localhost"
		cbUser := "admin"
		cbPassword := "password"
		cbBucket := "cb-bucket"
		cbScope := "scope"
		cbCollection := "cb-collection"
		cbGeneratorKey := "%id%"

		cbBatchSize := Integer(2000)
		bufferSize := Integer(20000)

		var (
			ctrl    *gomock.Controller
			migrate *mocktest.MockIMigrate[dOpts.Options]
			cmd     *cobra.Command
			action  *dynamodb.Action
		)
		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())
			migrate = mocktest.NewMockIMigrate[dOpts.Options](ctrl)
			action = &dynamodb.Action{Migrate: migrate}
			cmd = command.NewCommand()
			cmd.RunE = action.RunE
		})
		AfterEach(func() {
			ctrl.Finish()
		})
		Context("success", func() {
			It("Input assertion case1", func() {

				var dOptsGot *dOpts.Options
				var cbOptsGot *option.Options
				var copyIndexesGot bool
				var bufferSizeGot int
				migrate.EXPECT().Copy(gomock.Any(), gomock.Any(), true, 10000).DoAndReturn(func(dOpts *dOpts.Options, cbOpts *option.Options, copyIndexes bool, bufferSize int) error {
					dOptsGot = dOpts
					cbOptsGot = cbOpts
					copyIndexesGot = copyIndexes
					bufferSizeGot = bufferSize
					return nil
				})

				_, err := common.ExecuteCommand(cmd, dynamoDBTableNameOption, dynamoDBTableName,
					dynamoDBEndpointURLOption, dynamoDBEndpointURL, dynamoDBProfileOption, dynamoDBProfile,
					dynamoDBRegionOption, dynamoDBRegion, dynamoDBCaBundleOption, dynamoDBCaBundle,
					dynamoDBNoVerifySSLOption, cbClusterOption, cbCluster, cbUserOption, cbUser, cbPasswordOption,
					cbPassword, cbBucketOption, cbBucket, cbScopeOption, cbScope)
				Expect(err).To(BeNil())
				expectedDopts := &dOpts.Options{
					TableName:   dynamoDBTableName,
					EndpointUrl: dynamoDBEndpointURL,
					NoSSLVerify: true,
					Profile:     dynamoDBProfile,
					Region:      dynamoDBRegion,
					CABundle:    dynamoDBCaBundle,
				}
				expectedCbOpts := &option.Options{
					Cluster: cbCluster,
					Auth: &option.Auth{
						Username: cbUser,
						Password: cbPassword,
					},
					NameSpace: &option.NameSpace{
						Bucket:     cbBucket,
						Scope:      cbScope,
						Collection: dynamoDBTableName,
					},
					SSL:       &option.SSL{},
					BatchSize: 200,
				}

				Expect(dOptsGot).To(Equal(expectedDopts))
				Expect(cbOptsGot).To(Equal(expectedCbOpts))
				Expect(copyIndexesGot).To(Equal(true))
				Expect(bufferSizeGot).To(Equal(10000))
			})

			It("Input assertion case2", func() {

				var dOptsGot *dOpts.Options
				var cbOptsGot *option.Options
				var copyIndexesGot bool
				var bufferSizeGot int
				migrate.EXPECT().Copy(gomock.Any(), gomock.Any(), true, bufferSize.Int()).DoAndReturn(func(dOpts *dOpts.Options, cbOpts *option.Options, copyIndexes bool, bufferSize int) error {
					dOptsGot = dOpts
					cbOptsGot = cbOpts
					copyIndexesGot = copyIndexes
					bufferSizeGot = bufferSize
					return nil
				})

				_, err := common.ExecuteCommand(cmd, dynamoDBTableNameOption, dynamoDBTableName,
					cbClusterOption, cbCluster, cbUserOption, cbUser, cbPasswordOption, cbPassword,
					cbBucketOption, cbBucket, cbCollectionOption, cbCollection, cbGeneratorKeyOption, cbGeneratorKey,
					cbScopeOption, cbScope, cbBatchSizeOption, cbBatchSize.String(), bufferSizeOption, bufferSize.String())
				Expect(err).To(BeNil())
				expectedDopts := &dOpts.Options{
					TableName: dynamoDBTableName,
				}
				expectedCbOpts := &option.Options{
					Cluster: cbCluster,
					Auth: &option.Auth{
						Username: cbUser,
						Password: cbPassword,
					},
					NameSpace: &option.NameSpace{
						Bucket:     cbBucket,
						Scope:      cbScope,
						Collection: cbCollection,
					},
					SSL:          &option.SSL{},
					GeneratedKey: "%id%",
					BatchSize:    cbBatchSize.Int(),
				}

				Expect(dOptsGot).To(Equal(expectedDopts))
				Expect(cbOptsGot).To(Equal(expectedCbOpts))
				Expect(copyIndexesGot).To(Equal(true))
				Expect(bufferSizeGot).To(Equal(bufferSize.Int()))
			})

			It("Input assertion case3", func() {

				var dOptsGot *dOpts.Options
				var cbOptsGot *option.Options
				var copyIndexesGot bool
				var bufferSizeGot int
				migrate.EXPECT().Copy(gomock.Any(), gomock.Any(), true, bufferSize.Int()).DoAndReturn(func(dOpts *dOpts.Options, cbOpts *option.Options, copyIndexes bool, bufferSize int) error {
					dOptsGot = dOpts
					cbOptsGot = cbOpts
					copyIndexesGot = copyIndexes
					bufferSizeGot = bufferSize
					return nil
				})

				_, err := common.ExecuteCommand(cmd, dynamoDBTableNameOption, dynamoDBTableName, dynamoDBAccessKeyOption,
					dynamoDBAccessKey, dynamoDBSecretKeyOption, dynamoDBSecretKey,
					cbClusterOption, cbCluster, cbUserOption, cbUser, cbPasswordOption, cbPassword,
					cbBucketOption, cbBucket, cbCollectionOption, cbCollection, cbGeneratorKeyOption, cbGeneratorKey,
					cbScopeOption, cbScope, cbBatchSizeOption, cbBatchSize.String(), bufferSizeOption, bufferSize.String())
				Expect(err).To(BeNil())
				expectedDopts := &dOpts.Options{
					TableName: dynamoDBTableName,
					AccessKey: dynamoDBAccessKey,
					SecretKey: dynamoDBSecretKey,
				}
				expectedCbOpts := &option.Options{
					Cluster: cbCluster,
					Auth: &option.Auth{
						Username: cbUser,
						Password: cbPassword,
					},
					NameSpace: &option.NameSpace{
						Bucket:     cbBucket,
						Scope:      cbScope,
						Collection: cbCollection,
					},
					SSL:          &option.SSL{},
					GeneratedKey: "%id%",
					BatchSize:    cbBatchSize.Int(),
				}

				Expect(dOptsGot).To(Equal(expectedDopts))
				Expect(cbOptsGot).To(Equal(expectedCbOpts))
				Expect(copyIndexesGot).To(Equal(true))
				Expect(bufferSizeGot).To(Equal(bufferSize.Int()))
			})
		})
		Context("failure", func() {
			It("missing required flags", func() {
				_, err := common.ExecuteCommand(cmd,
					dynamoDBEndpointURLOption, dynamoDBEndpointURL, dynamoDBProfileOption, dynamoDBProfile,
					dynamoDBRegionOption, dynamoDBRegion, dynamoDBCaBundleOption, dynamoDBCaBundle,
					dynamoDBNoVerifySSLOption, cbClusterOption, cbCluster, cbUserOption, cbUser, cbPasswordOption,
					cbPassword, cbBucketOption, cbBucket, cbScopeOption, cbScope)
				Expect(err.Error()).To(Equal("required flag(s) \"dynamodb-table-name\" not set"))
			})
			It("exclusive error case 2", func() {
				_, err := common.ExecuteCommand(cmd, dynamoDBTableNameOption, dynamoDBTableName,
					dynamoDBProfileOption, dynamoDBProfile, dynamoDBSecretKeyOption, dynamoDBSecretKey,
					dynamoDBEndpointURLOption, dynamoDBEndpointURL,
					dynamoDBRegionOption, dynamoDBRegion, dynamoDBCaBundleOption, dynamoDBCaBundle,
					dynamoDBNoVerifySSLOption, cbClusterOption, cbCluster, cbUserOption, cbUser, cbPasswordOption,
					cbPassword, cbBucketOption, cbBucket, cbScopeOption, cbScope)
				Expect(err.Error()).To(Equal("error: \"--aws-profile\" is mutually exclusive with \"--aws-access-key\",\"aws-secret-key\" flags"))
			})
			It("exclusive error case 2", func() {
				_, err := common.ExecuteCommand(cmd, dynamoDBTableNameOption, dynamoDBTableName,
					dynamoDBProfileOption, dynamoDBProfile, dynamoDBAccessKeyOption, dynamoDBAccessKey,
					dynamoDBEndpointURLOption, dynamoDBEndpointURL,
					dynamoDBRegionOption, dynamoDBRegion, dynamoDBCaBundleOption, dynamoDBCaBundle,
					dynamoDBNoVerifySSLOption, cbClusterOption, cbCluster, cbUserOption, cbUser, cbPasswordOption,
					cbPassword, cbBucketOption, cbBucket, cbScopeOption, cbScope)
				Expect(err.Error()).To(Equal("error: \"--aws-profile\" is mutually exclusive with \"--aws-access-key\",\"aws-secret-key\" flags"))
			})
			It("must all error case 1", func() {
				_, err := common.ExecuteCommand(cmd, dynamoDBTableNameOption, dynamoDBTableName,
					dynamoDBAccessKeyOption, dynamoDBAccessKey,
					dynamoDBEndpointURLOption, dynamoDBEndpointURL,
					dynamoDBRegionOption, dynamoDBRegion, dynamoDBCaBundleOption, dynamoDBCaBundle,
					dynamoDBNoVerifySSLOption, cbClusterOption, cbCluster, cbUserOption, cbUser, cbPasswordOption,
					cbPassword, cbBucketOption, cbBucket, cbScopeOption, cbScope)
				Expect(err.Error()).To(Equal("inconsistent flag usage. Flags aws-access-key, aws-secret-key must all be provided together or not at all. Missing: aws-secret-key"))
			})
			It("must all error case 2", func() {
				_, err := common.ExecuteCommand(cmd, dynamoDBTableNameOption, dynamoDBTableName,
					dynamoDBSecretKeyOption, dynamoDBSecretKey,
					dynamoDBEndpointURLOption, dynamoDBEndpointURL,
					dynamoDBRegionOption, dynamoDBRegion, dynamoDBCaBundleOption, dynamoDBCaBundle,
					dynamoDBNoVerifySSLOption, cbClusterOption, cbCluster, cbUserOption, cbUser, cbPasswordOption,
					cbPassword, cbBucketOption, cbBucket, cbScopeOption, cbScope)
				Expect(err.Error()).To(Equal("inconsistent flag usage. Flags aws-access-key, aws-secret-key must all be provided together or not at all. Missing: aws-access-key"))
			})
		})
	})
})
