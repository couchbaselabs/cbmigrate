package common

//go:generate mockgen -source=source_definition.go -destination=../../testhelper/mock/source_definition.go -package=mock_test ISource
import (
	"context"
)

type ISource[Index any, Options any] interface {
	Init(opts *Options) error
	StreamData(context.Context, chan map[string]interface{}) error
	GetIndexes(ctx context.Context) ([]Index, error)
}
