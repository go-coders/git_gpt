package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
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

const (
	DefaultModel     = "gpt-4o-mini"
	DefaultMaxTokens = 4000
	DefaultBaseURL   = "https://api.openai.com/v1"
	ConfigFileName   = "config.json"
)

func Load() (*Config, error) {
	cfg := &Config{
		LLM: LLMConfig{
			Model:     DefaultModel,
			BaseURL:   DefaultBaseURL,
			MaxTokens: DefaultMaxTokens,
		},
	}
	configPath := getConfigPath()
	cfg.ConfigPath = configPath
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
	}
	return cfg, nil
}

func (c *Config) Save() error {
	data, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return err
	}

	configPath := getConfigPath()
	if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}

func getConfigPath() string {
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
