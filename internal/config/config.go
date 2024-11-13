// internal/config/config.go
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const (
	DefaultModel     = "gpt-4o"
	DefaultMaxTokens = 4000
	DefaultBaseURL   = "https://api.openai.com/v1"
	ConfigFileName   = "config.json"
)

type Config struct {
	LLM        LLMConfig `json:"llm"`
	ConfigPath string    `json:"config_path"`
}

type LLMConfig struct {
	APIKey    string `json:"api_key"`
	Model     string `json:"model"`
	BaseURL   string `json:"base_url"`
	MaxTokens int    `json:"max_tokens"`
}

// Load loads the configuration from the specified path
// If path is empty, it uses the default config location
func Load(path ...string) (*Config, error) {
	configPath := getConfigPath(path...)

	cfg := &Config{
		LLM: LLMConfig{
			Model:     DefaultModel,
			BaseURL:   DefaultBaseURL,
			MaxTokens: DefaultMaxTokens,
		},
		ConfigPath: configPath,
	}

	if err := cfg.loadFromFile(configPath); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
		// If file doesn't exist, return default config
		return cfg, nil
	}

	return cfg, nil
}

func (c *Config) loadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, c); err != nil {
		return fmt.Errorf("invalid config file format: %w", err)
	}

	return nil
}

func (c *Config) Save() error {
	data, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	dir := filepath.Dir(c.ConfigPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(c.ConfigPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// getConfigPath determines the configuration file path
func getConfigPath(customPath ...string) string {
	if len(customPath) > 0 && customPath[0] != "" {
		return customPath[0]
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ConfigFileName)
	}

	var configDir string
	switch runtime.GOOS {
	case "windows":
		configDir = filepath.Join(homeDir, "AppData", "Local", "GitChat")
	default:
		configDir = filepath.Join(homeDir, ".config", "gitchat")
	}

	return filepath.Join(configDir, ConfigFileName)
}
