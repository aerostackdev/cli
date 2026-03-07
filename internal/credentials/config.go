package credentials

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const configFile = "config.json"

// CLIConfig holds user-level CLI preferences (active workspace, etc.)
type CLIConfig struct {
	ActiveWorkspace string `json:"activeWorkspace,omitempty"`
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, credentialsDir, configFile), nil
}

// LoadConfig reads CLI config from ~/.aerostack/config.json.
// Returns an empty config (not an error) when the file doesn't exist yet.
func LoadConfig() (*CLIConfig, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &CLIConfig{}, nil
		}
		return nil, err
	}
	var cfg CLIConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return &CLIConfig{}, nil
	}
	return &cfg, nil
}

// SaveConfig writes CLI config to ~/.aerostack/config.json.
func SaveConfig(cfg *CLIConfig) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
