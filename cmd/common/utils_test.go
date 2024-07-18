package common_test

import (
	"github.com/couchbaselabs/cbmigrate/internal/couchbase/option"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
	"strconv"

	"github.com/couchbaselabs/cbmigrate/cmd/common"
	"github.com/couchbaselabs/cbmigrate/cmd/flag"
)

func newCommand() (*cobra.Command, *option.Options) {
	var opts = &option.Options{}
	var flags []flag.Flag
	flags = append(flags, common.GetCBFlags()...)
	flags = append(flags, common.GetCBGenerateKeyOption("%_id%"))
	flags = append(flags, common.GetCommonFlags()...)
	cmd := common.NewCommand("test", nil, nil, "test couchbase options", "test couchbase options", flags)
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		copts, err := common.ParesCouchbaseOptions(cmd, "test")
		if err != nil {
			return err
		}
		*opts = *copts
		return nil
	}
	return cmd, opts
}

type Integer int

func (i *Integer) String() string {
	return strconv.Itoa(int(*i))
}
func (i *Integer) Int() int {
	return int(*i)
}

var _ = Describe("utils", func() {
	cbClusterOption := "--" + common.CBCluster
	cbUserOption := "--" + common.CBUsername
	cbPasswordOption := "--" + common.CBPassword
	cbBucketOption := "--" + common.CBBucket
	cbScopeOption := "--" + common.CBScope
	cbCollectionOption := "--" + common.CBCollection
	cbBatchSizeOption := "--" + common.CBBatchSize
	cbKeepPrimaryKeyOption := "--" + common.KeepPrimaryKey
	cbHashDocumentKeyOption := "--" + common.HashDocumentKey

	bufferSizeOption := "--" + common.BufferSize

	cbCluster := "localhost"
	cbUser := "admin"
	cbPassword := "password"
	cbBucket := "cb-bucket"
	cbScope := "scope"
	cbCollection := "cb-collection"
	cbBatchSize := Integer(2000)
	cbHashDocumentKey := "sha256"
	bufferSize := Integer(20000)
	Describe("test couchbase common utilities", func() {
		Context("ParesCouchbaseOptions hash document key success", func() {
			It("keep primary key", func() {
				cmd, opts := newCommand()
				_, err := common.ExecuteCommand(cmd, cbClusterOption, cbCluster, cbUserOption, cbUser, cbPasswordOption, cbPassword,
					cbBucketOption, cbBucket, cbCollectionOption, cbCollection, cbKeepPrimaryKeyOption,
					cbScopeOption, cbScope, cbBatchSizeOption, cbBatchSize.String(), bufferSizeOption, bufferSize.String())
				Expect(err).To(BeNil())
				expectedOpts := &option.Options{
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
					SSL:            new(option.SSL),
					KeepPrimaryKey: true,
					GeneratedKey:   "%_id%",
					BatchSize:      int(cbBatchSize),
				}
				Expect(opts).To(Equal(expectedOpts))
			})
			It("hash document key sha256", func() {
				cmd, opts := newCommand()
				_, err := common.ExecuteCommand(cmd, cbClusterOption, cbCluster, cbUserOption, cbUser, cbPasswordOption, cbPassword,
					cbBucketOption, cbBucket, cbCollectionOption, cbCollection, cbHashDocumentKeyOption, cbHashDocumentKey,
					cbScopeOption, cbScope, cbBatchSizeOption, cbBatchSize.String(), bufferSizeOption, bufferSize.String())
				Expect(err).To(BeNil())
				expectedOpts := &option.Options{
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
					SSL:             new(option.SSL),
					GeneratedKey:    "%_id%",
					HashDocumentKey: cbHashDocumentKey,
					BatchSize:       int(cbBatchSize),
				}
				Expect(opts).To(Equal(expectedOpts))
			})
			It("hash document key sha512", func() {
				cmd, opts := newCommand()
				_, err := common.ExecuteCommand(cmd, cbClusterOption, cbCluster, cbUserOption, cbUser, cbPasswordOption, cbPassword,
					cbBucketOption, cbBucket, cbCollectionOption, cbCollection, cbHashDocumentKeyOption, "sha512",
					cbScopeOption, cbScope, cbBatchSizeOption, cbBatchSize.String(), bufferSizeOption, bufferSize.String())
				Expect(err).To(BeNil())
				expectedOpts := &option.Options{
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
					SSL:             new(option.SSL),
					GeneratedKey:    "%_id%",
					HashDocumentKey: "sha512",
					BatchSize:       int(cbBatchSize),
				}
				Expect(opts).To(Equal(expectedOpts))
			})
			It("hash document key sha512", func() {
				cmd, opts := newCommand()
				_, err := common.ExecuteCommand(cmd, cbClusterOption, cbCluster, cbUserOption, cbUser, cbPasswordOption, cbPassword,
					cbBucketOption, cbBucket, cbCollectionOption, cbCollection, cbHashDocumentKeyOption, "sha512",
					cbScopeOption, cbScope, cbBatchSizeOption, cbBatchSize.String(), bufferSizeOption, bufferSize.String())
				Expect(err).To(BeNil())
				expectedOpts := &option.Options{
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
					SSL:             new(option.SSL),
					GeneratedKey:    "%_id%",
					HashDocumentKey: "sha512",
					BatchSize:       int(cbBatchSize),
				}
				Expect(opts).To(Equal(expectedOpts))
			})
		})
		Context("ParesCouchbaseOptions hash document key failure", func() {
			It("hash document key random", func() {
				cmd, _ := newCommand()
				_, err := common.ExecuteCommand(cmd, cbClusterOption, cbCluster, cbUserOption, cbUser, cbPasswordOption, cbPassword,
					cbBucketOption, cbBucket, cbCollectionOption, cbCollection, cbHashDocumentKeyOption, "random",
					cbScopeOption, cbScope, cbBatchSizeOption, cbBatchSize.String(), bufferSizeOption, bufferSize.String())
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(Equal("value random must be one of [sha256 sha512]"))
			})
		})
	})
})
