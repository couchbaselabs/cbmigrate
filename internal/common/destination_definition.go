package common

import "github.com/couchbaselabs/cbmigrate/internal/option"

type IDestination interface {
	Init(opts *option.Options) error
	ProcessData(map[string]interface{}) error
	Complete() error
}
