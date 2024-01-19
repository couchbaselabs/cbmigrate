package option

import (
	cbOpts "github.com/couchbaselabs/cbmigrate/internal/couchbase/option"
	mOpts "github.com/couchbaselabs/cbmigrate/internal/mongo/option"
)

type Options struct {
	MOpts  *mOpts.Options
	CBOpts *cbOpts.Options
}
