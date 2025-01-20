package common

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"go.uber.org/zap"
)

type BinaryInfo struct {
	RepoOwner  string
	RepoName   string
	BinaryName string
	Version    string
}

func EnsureBinary(info BinaryInfo) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	cbmigrateDir := filepath.Join(homeDir, ".cbmigrate")
	binaryNameWithExt := info.BinaryName
	if runtime.GOOS == "windows" {
		binaryNameWithExt += ".exe"
	}
	// .cbmigrate/binary_foldername/binary
	binaryPath := filepath.Join(cbmigrateDir, info.BinaryName, binaryNameWithExt)

	// Check if binary exists
	if _, err := os.Stat(binaryPath); err == nil {
		// Read config to check version
		config, err := ReadConfig()
		if err != nil {
			zap.S().Warnf("Failed to read config: %v", err)
		} else if config != nil {
			if cmdConfig, exists := config.Commands[info.BinaryName]; exists && cmdConfig.Version == info.Version {
				// Version matches, use existing binary
				return binaryPath, nil
			}
		}

		// Version mismatch or couldn't read config, remove old binary
		zap.S().Warn("Removing old binary version...")
		if err := os.RemoveAll(filepath.Join(cbmigrateDir, info.BinaryName)); err != nil {
			return "", fmt.Errorf("failed to remove old binary: %w", err)
		}
	}

	// Binary doesn't exist or was removed, download it
	zap.S().Warn("Binary not found or outdated, downloading...")
	err = downloadAndExtract(cbmigrateDir, info)
	if err != nil {
		return "", err
	}

	// Write the new config since this is either initial setup or version changed
	if err := WriteBinaryConfig(info.Version, info.BinaryName); err != nil {
		zap.S().Warnf("Failed to write config: %v", err)
		// Continue even if config write fails
	}

	return binaryPath, nil
}

func downloadAndExtract(destDir string, info BinaryInfo) error {
	// Determine the expected asset name
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	var extension string
	if goos == "windows" {
		extension = ".zip"
	} else {
		extension = ".tar.gz"
	}

	githubDownloadAssertPrefix := fmt.Sprintf("https://github.com/%s/%s/releases/download/%s",
		info.RepoOwner, info.RepoName, info.Version)
	assetURL := fmt.Sprintf("%s/%s_%s_%s_%s%s", githubDownloadAssertPrefix, info.BinaryName,
		info.Version[1:], goos, goarch, extension)

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
	tmpFile, err := os.CreateTemp("", info.BinaryName+"_download_*_"+goos+"_"+goarch+extension)
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
		err = extractZipWithCommand(tmpFile.Name(), destDir, info.BinaryName)
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

func extractZipWithCommand(zipPath, destDir, binaryName string) error {
	cmd := exec.Command("powershell", "-NoProfile", "-Command",
		fmt.Sprintf(`Expand-Archive -Path "%s" -DestinationPath "%s" -Force`,
			zipPath, filepath.Join(destDir, binaryName)))
	output, err := cmd.CombinedOutput()
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
