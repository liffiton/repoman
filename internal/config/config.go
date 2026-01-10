// Package config handles the configuration for the repoman tool.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/zalando/go-keyring"
)

const (
	serviceName       = "repoman"
	keyName           = "api_key"
	configFileName    = "config.json"
	workspaceFileName = ".repoman.json"
	defaultBaseURL    = "https://crm.unsatisfiable.net"
)

// WorkspaceConfig holds directory-specific configuration.
type WorkspaceConfig struct {
	CourseID       string `json:"course_id"`
	AssignmentID   string `json:"assignment_id"`
	AssignmentName string `json:"assignment_name"`
}

// LoadWorkspace loads the workspace configuration from the current directory.
func LoadWorkspace() (*WorkspaceConfig, error) {
	data, err := os.ReadFile(workspaceFileName)
	if err != nil {
		return nil, err
	}
	var wcfg WorkspaceConfig
	if err := json.Unmarshal(data, &wcfg); err != nil {
		return nil, fmt.Errorf("could not unmarshal workspace config: %w", err)
	}
	return &wcfg, nil
}

// SaveWorkspace saves the workspace configuration to the current directory.
func (wcfg *WorkspaceConfig) SaveWorkspace() error {
	data, err := json.MarshalIndent(wcfg, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal workspace config: %w", err)
	}
	if err := os.WriteFile(workspaceFileName, data, 0o600); err != nil {
		return fmt.Errorf("could not write workspace config: %w", err)
	}
	return nil
}

// Config holds the configuration for repoman.
type Config struct {
	APIKey  string `json:"api_key,omitempty"`
	BaseURL string `json:"base_url,omitempty"`
}

// SaveResult describes where the configuration was saved.
type SaveResult struct {
	ConfigPath  string
	KeyringUsed bool
	FileWritten bool
}

// GetBaseURL returns the configured base URL or the default one.
func (cfg *Config) GetBaseURL() string {
	if cfg.BaseURL != "" {
		return cfg.BaseURL
	}
	return defaultBaseURL
}

// GetConfigPath returns the path to the repoman config file without creating directories.
func GetConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("could not get user config dir: %w", err)
	}
	return filepath.Join(configDir, "repoman", configFileName), nil
}

// EnsureConfigDir creates the repoman config directory if it doesn't exist.
func EnsureConfigDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("could not get user config dir: %w", err)
	}
	repomanDir := filepath.Join(configDir, "repoman")
	if err := os.MkdirAll(repomanDir, 0o700); err != nil {
		return "", fmt.Errorf("could not create config directory: %w", err)
	}
	return repomanDir, nil
}

// Load loads the configuration. It tries the keyring first for the API key,
// then falls back to the config file.
func Load() (*Config, error) {
	cfg := &Config{}

	// 1. Try to get API key from keyring
	apiKey, err := keyring.Get(serviceName, keyName)
	if err == nil {
		cfg.APIKey = apiKey
	}

	// 2. Load from config file
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// #nosec G304
	data, err := os.ReadFile(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return nil, fmt.Errorf("could not read config file: %w", err)
	}

	var fileCfg Config
	if err := json.Unmarshal(data, &fileCfg); err != nil {
		return nil, fmt.Errorf("could not unmarshal config: %w", err)
	}

	// If APIKey wasn't in keyring, use the one from the file
	if cfg.APIKey == "" {
		cfg.APIKey = fileCfg.APIKey
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = fileCfg.BaseURL
	}

	return cfg, nil
}

// Save saves the configuration. It attempts to save the API key to the keyring,
// but falls back to saving it in the config file if necessary.
func (cfg *Config) Save() (*SaveResult, error) {
	result := &SaveResult{}

	keyringErr := keyring.Set(serviceName, keyName, cfg.APIKey)
	if keyringErr == nil {
		result.KeyringUsed = true
	}

	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}
	result.ConfigPath = configPath

	saveCfg := *cfg
	if result.KeyringUsed {
		saveCfg.APIKey = ""
	}

	// Only write the file if there's actually something to save that isn't empty.
	if saveCfg.APIKey != "" || saveCfg.BaseURL != "" {
		if _, err := EnsureConfigDir(); err != nil {
			return nil, err
		}

		data, err := json.MarshalIndent(saveCfg, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("could not marshal config: %w", err)
		}

		if err := os.WriteFile(configPath, data, 0o600); err != nil {
			return nil, fmt.Errorf("could not write config file: %w", err)
		}
		result.FileWritten = true
	}

	return result, nil
}

// SetAPIKey specifically updates the API key.
func (cfg *Config) SetAPIKey(key string) (*SaveResult, error) {
	cfg.APIKey = key
	return cfg.Save()
}
