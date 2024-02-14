package couchbase

import (
	"github.com/couchbase/gocb/v2"
	"go.uber.org/zap"
	"time"

	"github.com/couchbaselabs/cbmigrate/internal/couchbase/option"
)

type DB struct {
	*gocb.Cluster
	*gocb.Bucket
}

func (d *DB) Init(uri string, opts *option.Options) error {
	dbOpts, err := createCouchbaseOptions(opts)
	if err != nil {
		return err
	}
	cluster, err := gocb.Connect(uri, dbOpts)
	if err != nil {
		return err
	}
	err = cluster.WaitUntilReady(15*time.Second, nil)
	if err != nil {
		return err
	}
	d.Cluster = cluster
	zap.S().Info("Couchbase connection successful")
	d.Bucket = cluster.Bucket(opts.Bucket)
	return nil

}
