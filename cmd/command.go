package cmd

import (
	"fmt"
	"github.com/couchbaselabs/cbmigrate/cmd/common"
	"github.com/couchbaselabs/cbmigrate/cmd/mongo"
	"github.com/spf13/cobra"
	"os"

	"github.com/couchbaselabs/cbmigrate/cmd/flag"
)

const (
	Version = "version"
)

// to hind the default auto-completion script generation command
func completionCommand() *cobra.Command {
	return &cobra.Command{
		Use:    "completion",
		Short:  "Generate the autocompletion script for the specified shell",
		Hidden: true,
	}
}

// Execute executes root command
func Execute() {
	flags := []flag.Flag{
		&flag.BoolFlag{
			Name:  Version,
			Alias: "v",
			Usage: "Display the version of this tool.",
		},
	}
	cmd := common.NewCommand("cbmigrate", nil, nil, "", "", flags)
	cmd.AddCommand(completionCommand())
	cmd.AddCommand(mongo.GetMongoMigrateCommand())
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if cmd.Flags().Changed(Version) {
			version, _ := cmd.Flags().GetString(Version)
			fmt.Println("Version: " + version)
			return nil
		}
		return cmd.Help()
	}
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
