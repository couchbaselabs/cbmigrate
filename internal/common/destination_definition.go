package common

//go:generate mockgen -source=destination_definition.go -destination=../../testhelper/mock/destination_definition.go -package=mock IDestination

import (
	"github.com/couchbaselabs/cbmigrate/internal/couchbase/option"
)

type IDestination interface {
	Init(opts *option.Options, documentKey ICBDocumentKey) error
	ProcessData(map[string]interface{}) error
	Complete() error
	CreateIndexes(indexes []Index) error
}
