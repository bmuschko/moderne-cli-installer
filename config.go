package main

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const configFileName = "config.yaml"

// Config holds the application configuration.
type Config struct {
	Download DownloadConfig `yaml:"download"`
}

// DownloadConfig holds download-related settings.
type DownloadConfig struct {
	BaseURL string       `yaml:"baseUrl"`
	Proxy   *ProxyConfig `yaml:"proxy,omitempty"`
}

// ProxyConfig holds proxy settings.
type ProxyConfig struct {
	URL      string `yaml:"url"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
	NoProxy  string `yaml:"noProxy,omitempty"`
}

// HasProxy returns true if proxy configuration is provided.
func (d *DownloadConfig) HasProxy() bool {
	return d.Proxy != nil && d.Proxy.URL != ""
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		Download: DownloadConfig{
			BaseURL: DefaultBaseURL,
		},
	}
}

// LoadConfig loads configuration from file, with fallback to defaults.
// Priority: external file (next to binary or cwd) > defaults
func LoadConfig() (*Config, string, error) {
	config := DefaultConfig()

	// Try external file next to binary
	exePath, err := os.Executable()
	if err == nil {
		externalPath := filepath.Join(filepath.Dir(exePath), configFileName)
		if loaded, err := loadConfigFile(externalPath); err == nil {
			mergeConfig(config, loaded)
			return config, externalPath, nil
		}
	}

	// Try external file in current working directory
	if loaded, err := loadConfigFile(configFileName); err == nil {
		mergeConfig(config, loaded)
		cwd, _ := os.Getwd()
		return config, filepath.Join(cwd, configFileName), nil
	}

	return config, "defaults", nil
}

func loadConfigFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// mergeConfig merges loaded config into base config (non-empty values override).
func mergeConfig(base, loaded *Config) {
	if loaded.Download.BaseURL != "" {
		base.Download.BaseURL = loaded.Download.BaseURL
	}
	if loaded.Download.Proxy != nil {
		base.Download.Proxy = loaded.Download.Proxy
	}
}
