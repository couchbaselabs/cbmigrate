package common

//go:generate mockgen -source=source_definition.go -destination=../../testhelper/mock/source_definition.go -package=mock ISource
import (
	"context"
)

type ISource[Options any] interface {
	Init(opts *Options, documentKey ICBDocumentKey) error
	StreamData(context.Context, chan map[string]interface{}) error
	GetCouchbaseIndexesQuery(bucket string, scope string, collection string) ([]Index, error)
}
