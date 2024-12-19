package command

import (
	"github.com/couchbaselabs/cbmigrate/cmd/common"
	"github.com/couchbaselabs/cbmigrate/cmd/flag"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	const (
		shortDesc = "Migrate data from Hugging Face to Couchbase"
		longDesc  = `Migrate datasets from Hugging Face to Couchbase.
This command allows you to download and import Hugging Face datasets into Couchbase.
Use --help to see available options.`
	)

	flags := []flag.Flag{common.GetDebugFlag()}
	cmd := common.NewCommand(
		common.HuggingFace,
		[]string{"h", "hf"}, // Add "hf" as additional alias
		nil,
		shortDesc,
		longDesc,
		flags,
	)

	// Disable flag parsing since we pass flags to the underlying binary
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.DisableFlagParsing = true

	return cmd
}
