package common

//go:generate mockgen -source=destination_definition.go -destination=../../testhelper/mock/destination_definition.go -package=mock_test IDestination

import "github.com/couchbaselabs/cbmigrate/internal/option"

type IDestination interface {
	Init(opts *option.Options) error
	ProcessData(map[string]interface{}) error
	Complete() error
	CreateIndexes(indexes []Index, fieldPaths IndexFieldPath) error
}
