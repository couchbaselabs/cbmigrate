package migrater

import (
	"context"
	"errors"
	"github.com/couchbaselabs/cbmigrate/internal/common"
	"github.com/couchbaselabs/cbmigrate/internal/index"
	"github.com/couchbaselabs/cbmigrate/internal/option"
	"golang.org/x/sync/errgroup"
)

type IMigrate interface {
	Copy(opts *option.Options) error
}

type Migrate struct {
	Source      common.ISource
	Analyzer    index.Analyzer
	Destination common.IDestination
}

func (m Migrate) Copy(opts *option.Options) error {

	err := m.Source.Init(opts)
	if err != nil {
		return err
	}
	err = m.Destination.Init(opts)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())

	indexes, err := m.Source.GetIndexes(ctx)
	if err != nil {
		return err
	}
	m.Analyzer.Init(indexes)

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
			m.Analyzer.AnalyzeData(data)
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

	fieldPaths := m.Analyzer.GetIndexFieldPath()
	err = m.Destination.CreateIndexes(indexes, fieldPaths)
	return err
}

func NewMigrator(source common.ISource, destination common.IDestination, analyzer index.Analyzer) IMigrate {
	return Migrate{
		Source:      source,
		Destination: destination,
		Analyzer:    analyzer,
	}
}
