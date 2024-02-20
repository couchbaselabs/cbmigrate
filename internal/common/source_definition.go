package common

//go:generate mockgen -source=source_definition.go -destination=../../testhelper/mock/source_definition.go -package=mock_test ISource
import (
	"context"
	"github.com/couchbaselabs/cbmigrate/internal/index"
	"github.com/couchbaselabs/cbmigrate/internal/option"
)

type ISource interface {
	Init(opts *option.Options) error
	StreamData(context.Context, chan map[string]interface{}) error
	GetIndexes(ctx context.Context) ([]index.Index, error)
}
