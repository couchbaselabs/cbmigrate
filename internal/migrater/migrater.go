package migrater

import (
	"context"
	"errors"
	"github.com/couchbaselabs/cbmigrate/internal/common"
	"github.com/couchbaselabs/cbmigrate/internal/couchbase/option"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

//go:generate mockgen -source=migrater.go -destination=../../testhelper/mock/migrater.go -package=mock_test IMigrate
type IMigrate[Options any] interface {
	Copy(mOpts *Options, cbOpts *option.Options, copyIndexes bool, bufferSize int) error
}

type Migrate[Options any] struct {
	Source      common.ISource[Options]
	Destination common.IDestination
}

func (m Migrate[Options]) Copy(mOpts *Options, cbOpts *option.Options, copyIndexes bool, bufferSize int) error {

	documentKey := common.NewCBDocumentKey()
	err := m.Source.Init(mOpts, documentKey)
	if err != nil {
		return err
	}
	err = m.Destination.Init(cbOpts, documentKey)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	zap.S().Info("data migration started")
	var mChan = make(chan map[string]interface{}, bufferSize)
	g := errgroup.Group{}
	var sErr, dErr error
	g.Go(func() error {
		// message channel should be closed properly inside stream data function using syntax:  close(mChan)
		sErr = m.Source.StreamData(ctx, mChan)
		return nil
	})
	g.Go(func() error {
		for data := range mChan {
			dErr = m.Destination.ProcessData(data)
			if dErr != nil {
				cancel()
				return nil
			}
		}
		dErr = m.Destination.Complete()
		return nil
	})
	_ = g.Wait()
	if dErr != nil {
		err = errors.Join(err, dErr)
	}
	if sErr != nil {
		err = errors.Join(err, sErr)
	}
	if err != nil {
		return err
	}
	zap.S().Info("data migration completed")
	if copyIndexes {
		zap.S().Info("index migration started")
		cbIndexes, err := m.Source.GetCouchbaseIndexesQuery(cbOpts.Bucket, cbOpts.Scope, cbOpts.Collection)
		if err != nil {
			return err
		}
		err = m.Destination.CreateIndexes(cbIndexes)
		if err != nil {
			return err
		}
		zap.S().Info("index migration completed")
	}
	return err
}

func NewMigrator[Options any](source common.ISource[Options], destination common.IDestination) IMigrate[Options] {
	return Migrate[Options]{
		Source:      source,
		Destination: destination,
	}
}
