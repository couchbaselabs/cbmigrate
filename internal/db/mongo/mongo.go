package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	"github.com/couchbaselabs/cbmigrate/internal/feature"
	"github.com/couchbaselabs/cbmigrate/internal/mongo/option"
)

type DB struct {
	*mongo.Database
}

func (d *DB) Init(opts *option.Options) error {
	var client *mongo.Client
	var err error
	switch {
	case !feature.IsFeatureEnabled(feature.CbmigrateMongoHostOptsConfig):
		client, err = configureClientWithOnlyUri(opts)
	default:
		client, err = configureClient(opts)
	}
	if err != nil {
		return err
	}
	err = client.Ping(context.Background(), nil)
	if err != nil {
		return err
	}
	zap.S().Info("Mongodb connection successful")
	d.Database = client.Database(opts.DB)
	return nil
}
