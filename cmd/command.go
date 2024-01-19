package cmd

import (
	"github.com/couchbaselabs/cbmigrate/cmd/common"
	"github.com/couchbaselabs/cbmigrate/cmd/mongo"
	"github.com/spf13/cobra"
	"os"
)

func completionCommand() *cobra.Command {
	return &cobra.Command{
		Use:    "completion",
		Short:  "Generate the autocompletion script for the specified shell",
		Hidden: true,
	}
}

// Execute executes root command
func Execute() {
	cmd := common.NewCommand("cbmigrate", nil, nil, "", "", nil)
	cmd.AddCommand(completionCommand())
	cmd.AddCommand(mongo.GetMongoMigrateCommand())
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
