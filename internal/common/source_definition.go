package common

//go:generate mockgen -source=source_definition.go -destination=../../testhelper/mock/source_definition.go -package=mock_test ISource
import (
	"context"
)

type ISource[Options any] interface {
	Init(opts *Options, documentKey IDocumentKey) error
	StreamData(context.Context, chan map[string]interface{}) error
	GetCouchbaseIndexesQuery(bucket string, scope string, collection string) []Index
}
