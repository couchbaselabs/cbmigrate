package mongo_test

import (
	_ "embed"
	"github.com/couchbaselabs/cbmigrate/cmd/common"
	"github.com/couchbaselabs/cbmigrate/cmd/mongo"
	"github.com/couchbaselabs/cbmigrate/cmd/mongo/command"
	"github.com/couchbaselabs/cbmigrate/internal/couchbase/option"
	mOpts "github.com/couchbaselabs/cbmigrate/internal/mongo/option"
	mock_test "github.com/couchbaselabs/cbmigrate/testhelper/mock"
	"github.com/spf13/cobra"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zapcore"
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
		mongodbUriOption := "--" + command.MongoDBURI
		mongodbDbOption := "--" + command.MongoDBDatabase
		mongodbCollectionOption := "--" + command.MongoDBCollection
		cbClusterOption := "--" + common.CBCluster
		cbUserOption := "--" + common.CBUsername
		cbPasswordOption := "--" + common.CBPassword
		cbBucketOption := "--" + common.CBBucket
		cbScopeOption := "--" + common.CBScope
		cbCollectionOption := "--" + common.CBCollection
		cbBatchSizeOption := "--" + common.CBBatchSize

		bufferSizeOption := "--" + common.BufferSize

		mongodbUri := "uri"
		mongodbDb := "mongo-db"
		mongodbCollection := "mongo-collection"
		cbCluster := "localhost"
		cbUser := "admin"
		cbPassword := "password"
		cbBucket := "cb-bucket"
		cbScope := "scope"
		cbCollection := "cb-collection"
		cbBatchSize := Integer(2000)
		bufferSize := Integer(20000)
		var (
			ctrl    *gomock.Controller
			migrate *mock_test.MockIMigrate[mOpts.Options]
			cmd     *cobra.Command
			action  *mongo.Action
		)
		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())
			migrate = mock_test.NewMockIMigrate[mOpts.Options](ctrl)
			action = &mongo.Action{Migrate: migrate}
			cmd = command.NewCommand()
			cmd.RunE = action.RunE
		})
		AfterEach(func() {
			ctrl.Finish()
		})
		Context("success", func() {
			It("Input assertion case1", func() {

				var mOptsGot *mOpts.Options
				var cbOptsGot *option.Options
				var copyIndexesGot bool
				var bufferSizeGot int
				migrate.EXPECT().Copy(gomock.Any(), gomock.Any(), true, 10000).DoAndReturn(func(mOpts *mOpts.Options, cbOpts *option.Options, copyIndexes bool, bufferSize int) error {
					mOptsGot = mOpts
					cbOptsGot = cbOpts
					copyIndexesGot = copyIndexes
					bufferSizeGot = bufferSize
					return nil
				})

				_, err := common.ExecuteCommand(cmd, mongodbUriOption, mongodbUri, mongodbDbOption, mongodbDb,
					mongodbCollectionOption, mongodbCollection,
					cbClusterOption, cbCluster, cbUserOption, cbUser, cbPasswordOption, cbPassword,
					cbBucketOption, cbBucket, cbScopeOption, cbScope)
				Expect(err).To(BeNil())
				expectedMopts := &mOpts.Options{
					URI: &mOpts.URI{
						ConnectionString: mongodbUri,
					},
					Namespace: &mOpts.Namespace{
						Collection: mongodbCollection,
						DB:         mongodbDb,
					},
					Connection: &mOpts.Connection{},
					SSL:        &mOpts.SSL{UseSSL: true},
					Auth:       &mOpts.Auth{},
					Kerberos:   &mOpts.Kerberos{},
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
						Collection: mongodbCollection,
					},
					SSL:          &option.SSL{},
					GeneratedKey: "%_id%",
					BatchSize:    200,
				}

				Expect(mOptsGot).To(Equal(expectedMopts))
				Expect(cbOptsGot).To(Equal(expectedCbOpts))
				Expect(copyIndexesGot).To(Equal(true))
				Expect(bufferSizeGot).To(Equal(10000))
			})

			It("Input assertion case2", func() {

				var mOptsGot *mOpts.Options
				var cbOptsGot *option.Options
				var copyIndexesGot bool
				var bufferSizeGot int
				migrate.EXPECT().Copy(gomock.Any(), gomock.Any(), true, bufferSize.Int()).DoAndReturn(func(mOpts *mOpts.Options, cbOpts *option.Options, copyIndexes bool, bufferSize int) error {
					mOptsGot = mOpts
					cbOptsGot = cbOpts
					copyIndexesGot = copyIndexes
					bufferSizeGot = bufferSize
					return nil
				})

				_, err := common.ExecuteCommand(cmd, mongodbUriOption, mongodbUri, mongodbDbOption, mongodbDb,
					mongodbCollectionOption, mongodbCollection,
					cbClusterOption, cbCluster, cbUserOption, cbUser, cbPasswordOption, cbPassword,
					cbBucketOption, cbBucket, cbCollectionOption, cbCollection,
					cbScopeOption, cbScope, cbBatchSizeOption, cbBatchSize.String(), bufferSizeOption, bufferSize.String())
				Expect(err).To(BeNil())
				expectedMopts := &mOpts.Options{
					URI: &mOpts.URI{
						ConnectionString: mongodbUri,
					},
					Namespace: &mOpts.Namespace{
						Collection: mongodbCollection,
						DB:         mongodbDb,
					},
					Connection: &mOpts.Connection{},
					SSL:        &mOpts.SSL{UseSSL: true},
					Auth:       &mOpts.Auth{},
					Kerberos:   &mOpts.Kerberos{},
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
					GeneratedKey: "%_id%",
					BatchSize:    cbBatchSize.Int(),
				}

				Expect(mOptsGot).To(Equal(expectedMopts))
				Expect(cbOptsGot).To(Equal(expectedCbOpts))
				Expect(copyIndexesGot).To(Equal(true))
				Expect(bufferSizeGot).To(Equal(bufferSize.Int()))
			})

		})

		Context("failure", func() {
		})
	})
})

type FatalHook struct {
}

func (h FatalHook) OnWrite(*zapcore.CheckedEntry, []zapcore.Field) {
}
