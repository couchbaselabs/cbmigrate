package common

import (
	"context"
	"github.com/couchbaselabs/cbmigrate/internal/option"
)

type ISource interface {
	Init(opts *option.Options) error
	StreamData(context.Context, chan map[string]interface{}) error
}
