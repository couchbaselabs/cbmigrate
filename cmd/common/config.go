package common

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type CommandConfig struct {
	Version     string    `json:"version"`
	LastUpdated time.Time `json:"last_updated"`
}

type Config struct {
	Commands map[string]*CommandConfig `json:"commands,omitempty"`
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
			return &Config{
				Commands: make(map[string]*CommandConfig),
			}, nil
		}
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	if config.Commands == nil {
		config.Commands = make(map[string]*CommandConfig)
	}

	return &config, nil
}

func WriteBinaryConfig(version string, binaryName string) error {
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
		Version:     version,
		LastUpdated: time.Now(),
	}

	config.Commands[binaryName] = cmdConfig

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
