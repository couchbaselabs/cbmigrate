package command

import (
	"github.com/couchbaselabs/cbmigrate/cmd/common"
	"github.com/couchbaselabs/cbmigrate/cmd/flag"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	usage := "Migrate data from Hugging face to Couchbase"
	flags := []flag.Flag{common.GetDebugFlag()}
	cmd := common.NewCommand(common.HuggingFace, []string{"h"}, nil, usage, usage, flags)
	// Disable flag parsing
	cmd.DisableFlagParsing = true
	return cmd
}
