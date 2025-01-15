package huggingface

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/couchbaselabs/cbmigrate/cmd/common"
	"github.com/couchbaselabs/cbmigrate/cmd/huggingface/command"
	"github.com/couchbaselabs/cbmigrate/internal/pkg/logger"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

const (
	repoOwner  = "Couchbase-Ecosystem"
	repoName   = "hf-to-cb-dataset-migrator"
	binaryName = "hf_to_cb_dataset_migrator"
	releaseTag = "v1.0.2"
)

func executeHuggingFaceCommand(args []string) error {
	info := common.BinaryInfo{
		RepoOwner:  repoOwner,
		RepoName:   repoName,
		BinaryName: binaryName,
		Version:    releaseTag,
	}

	binaryPath, err := common.EnsureBinary(info)
	if err != nil {
		return err
	}

	execCmd := exec.Command(binaryPath, args...)
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	execCmd.Env = []string{"RUN_FROM_CBMIGRATE=true"}
	if os.Getenv("MOCK_CLI_FOR_CBMIGRATE") == "true" {
		execCmd.Env = append(execCmd.Env, "MOCK_CLI_FOR_CBMIGRATE=true")
	}
	return execCmd.Run()
}

func GetHuggingFaceMigrateCommand() *cobra.Command {
	cmd := command.NewCommand()

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if isJsonOutputPreset(args) {
			logger.EnableErrorLevel()
		}

		if err := executeHuggingFaceCommand(args); err != nil {
			return fmt.Errorf("error executing binary: %w", err)
		}
		return nil
	}

	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		if isJsonOutputPreset(args) {
			logger.EnableErrorLevel()
		}

		cmdArgs := args[1:]
		if err := executeHuggingFaceCommand(cmdArgs); err != nil {
			zap.S().Fatal(fmt.Errorf("error executing binary: %w", err))
		}
	})
	return cmd
}

func isJsonOutputPreset(arg []string) bool {
	for _, arg := range arg {
		if arg == "--json-output" {
			return true
		}
	}
	return false
}
