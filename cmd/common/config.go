package common

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type CommandConfig struct {
	Version string `json:"version"`
}

type Config struct {
	HuggingFace *CommandConfig `json:"hugging_face,omitempty"`
	LastUpdated time.Time      `json:"last_updated"`
}

func ReadConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(homeDir, ".cbmigrate", "cbmigrate.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func WriteBinaryConfig(version string, commandType string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// Read existing config
	config, err := ReadConfig()
	if err != nil {
		return fmt.Errorf("failed to read existing config: %w", err)
	}

	// Update specific command config
	cmdConfig := &CommandConfig{
		Version: version,
	}

	switch commandType {
	case "hugging_face":
		config.HuggingFace = cmdConfig
	// Add other command types here
	default:
		return fmt.Errorf("unknown command type: %s", commandType)
	}

	// Update last updated time
	config.LastUpdated = time.Now()

	configPath := filepath.Join(homeDir, ".cbmigrate", "cbmigrate.json")

	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
