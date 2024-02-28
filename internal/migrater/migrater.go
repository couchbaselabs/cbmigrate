package migrater

import (
	"context"
	"errors"
	"github.com/couchbaselabs/cbmigrate/internal/common"
	"github.com/couchbaselabs/cbmigrate/internal/couchbase/option"
	"github.com/couchbaselabs/cbmigrate/internal/index"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type IMigrate[Options any] interface {
	Copy(mOpts *Options, cbOpts *option.Options, copyIndexes bool) error
}

type Migrate[T any, Options any] struct {
	Source      common.ISource[T, Options]
	Analyzer    index.Analyzer[T]
	Destination common.IDestination
}

func (m Migrate[T, Options]) Copy(mOpts *Options, cbOpts *option.Options, copyIndexes bool) error {

	err := m.Source.Init(mOpts)
	if err != nil {
		return err
	}
	err = m.Destination.Init(cbOpts)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var indexes []T
	if copyIndexes {
		indexes, err = m.Source.GetIndexes(ctx)
		if err != nil {
			return err
		}
	}
	m.Analyzer.Init(indexes)

	zap.S().Info("data migration started")
	var mChan = make(chan map[string]interface{}, 10000)
	g := errgroup.Group{}
	var sErr, dErr error
	g.Go(func() error {
		sErr = m.Source.StreamData(ctx, mChan)
		close(mChan)
		return nil
	})
	g.Go(func() error {
		for data := range mChan {
			if copyIndexes {
				m.Analyzer.AnalyzeData(data)
			}
			dErr = m.Destination.ProcessData(data)
			if dErr != nil {
				cancel()
				return nil
			}
		}
		dErr = m.Destination.Complete()
		return nil
	})
	g.Wait()
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

	cbIndexes := m.Analyzer.GetCouchbaseQuery(cbOpts.Bucket, cbOpts.Scope, cbOpts.Collection)
	if copyIndexes {
		zap.S().Info("index migration started")
		err = m.Destination.CreateIndexes(cbIndexes)
		if err != nil {
			return err
		}
		zap.S().Info("index migration completed")
	}
	return err
}

func NewMigrator[T any, Options any](source common.ISource[T, Options], destination common.IDestination, analyzer index.Analyzer[T]) IMigrate[Options] {
	return Migrate[T, Options]{
		Source:      source,
		Destination: destination,
		Analyzer:    analyzer,
	}
}
