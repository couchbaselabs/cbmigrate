package migrater

import (
	"context"
	"errors"
	"github.com/couchbaselabs/cbmigrate/internal/common"
	"github.com/couchbaselabs/cbmigrate/internal/option"
	"golang.org/x/sync/errgroup"
)

type IMigrate interface {
	Copy(opts *option.Options) error
}

type Migrate struct {
	Source      common.ISource
	Analyzer    common.Analyzer
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
	m.Analyzer.Init(indexes)

	var sChan = make(chan map[string]interface{}, 10000)
	var dChan = make(chan map[string]interface{}, 10000)
	g := errgroup.Group{}
	var sErr, dErr error
	g.Go(func() error {
		sErr = m.Source.StreamData(ctx, sChan)
		close(sChan)
		return nil
	})
	g.Go(func() error {
		for data := range sChan {
			m.Analyzer.AnalyzeData(data)
			dChan <- data
		}
		close(dChan)
		return nil
	})
	g.Go(func() error {
		for data := range dChan {
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

func NewMigrator(source common.ISource, destination common.IDestination, analyzer common.Analyzer) IMigrate {
	return Migrate{
		Source:      source,
		Destination: destination,
		Analyzer:    analyzer,
	}
}
