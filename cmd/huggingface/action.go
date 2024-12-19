package huggingface

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/couchbaselabs/cbmigrate/cmd/huggingface/command"
	"github.com/couchbaselabs/cbmigrate/internal/pkg/logger"
	"github.com/google/go-github/v66/github"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

const (
	repoOwner  = "Couchbase-Ecosystem"
	repoName   = "hf-to-cb-dataset-migrator"
	binaryName = "hf_to_cb_dataset_migrator"
	releaseTag = "v1.0.0"
)

func downloadAndExtract(destDir string, releaseTag string) error {

	ctx := context.Background()
	client := github.NewClient(nil)

	// Get the specified release
	release, _, err := client.Repositories.GetReleaseByTag(ctx, repoOwner, repoName, releaseTag)
	if err != nil {
		return err
	}

	// Determine the expected asset name
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	var extension string
	if goos == "windows" {
		extension = ".zip"
	} else {
		extension = ".tar.gz"
	}

	expectedAssetName := fmt.Sprintf("%s_%s_%s_%s%s", binaryName, strings.TrimLeft(releaseTag, "v"), goos, goarch, extension)

	// Find the asset
	var assetURL string
	for _, asset := range release.Assets {
		if asset.GetName() == expectedAssetName {
			assetURL = asset.GetBrowserDownloadURL()
			break
		}
	}
	if assetURL == "" {
		return fmt.Errorf("asset not found for Release: %s, OS: %s, Arch: %s, ", strings.TrimLeft(releaseTag, "v"), goos, goarch)
	}

	zap.S().Warnf("Downloading asset: %s\n", assetURL)

	// Download the asset
	resp, err := http.Get(assetURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download asset: %s", resp.Status)
	}

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", binaryName+"_download_*")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return err
	}
	tmpFile.Close()

	// Extract the binary using command-line tools
	err = os.MkdirAll(destDir, 0755)
	if err != nil {
		return err
	}

	switch extension {
	case ".zip":
		err = extractZipWithCommand(tmpFile.Name(), destDir)
		if err != nil {
			return err
		}
	case ".tar.gz":
		err = extractTarGzWithCommand(tmpFile.Name(), destDir)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported extension: %s", extension)
	}

	return nil
}

func extractZipWithCommand(zipPath, destDir string) error {
	cmd := exec.Command("unzip", zipPath, "-d", destDir)
	output, err := cmd.Output()
	if err != nil {
		zap.S().Errorf("%s", string(output))
		return fmt.Errorf("failed to extract zip: %w", err)
	} else {
		zap.S().Warn(string(output))
	}
	return nil
}

func extractTarGzWithCommand(tarGzPath, destDir string) error {
	cmd := exec.Command("tar", "-xzvf", tarGzPath, "-C", destDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		zap.S().Errorf("%s", string(output))
		return fmt.Errorf("failed to extract zip: %w", err)
	} else {
		zap.S().Warn(string(output))
	}
	return nil
}

func ensureBinary() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	cbmigrateDir := filepath.Join(homeDir, ".cbmigrate")
	binaryNameWithExt := binaryName
	if runtime.GOOS == "windows" {
		binaryNameWithExt += ".exe"
	}
	binaryPath := filepath.Join(cbmigrateDir, binaryName, binaryNameWithExt)

	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		// Binary doesn't exist, download it
		zap.S().Warn("Binary not found, downloading...")

		err = downloadAndExtract(cbmigrateDir, releaseTag)
		if err != nil {
			return "", err
		}
	}
	return binaryPath, nil
}

func executeHuggingFaceCommand(args []string) error {
	binaryPath, err := ensureBinary()
	if err != nil {
		return err
	}

	execCmd := exec.Command(binaryPath, args...)
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	execCmd.Env = []string{"RUN_FROM_CBMIGRATE=true"}
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
