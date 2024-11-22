package command

import (
	"github.com/couchbaselabs/cbmigrate/cmd/common"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	usage := "Migrate data from Hugging face to Couchbase"
	return common.NewCommand(common.HuggingFace, []string{"h"}, nil, usage, usage, nil)
}
