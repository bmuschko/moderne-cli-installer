package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, DefaultBaseURL, config.Download.BaseURL)
	assert.Nil(t, config.Download.Proxy)
}

func TestDownloadConfig_HasProxy(t *testing.T) {
	tests := []struct {
		name     string
		config   DownloadConfig
		expected bool
	}{
		{
			name:     "no proxy configured",
			config:   DownloadConfig{BaseURL: "http://example.com"},
			expected: false,
		},
		{
			name: "proxy is nil",
			config: DownloadConfig{
				BaseURL: "http://example.com",
				Proxy:   nil,
			},
			expected: false,
		},
		{
			name: "proxy with empty URL",
			config: DownloadConfig{
				BaseURL: "http://example.com",
				Proxy:   &ProxyConfig{URL: ""},
			},
			expected: false,
		},
		{
			name: "proxy with URL",
			config: DownloadConfig{
				BaseURL: "http://example.com",
				Proxy:   &ProxyConfig{URL: "http://proxy:8080"},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.config.HasProxy())
		})
	}
}

func TestMergeConfig(t *testing.T) {
	t.Run("merges baseURL", func(t *testing.T) {
		base := DefaultConfig()
		loaded := &Config{
			Download: DownloadConfig{
				BaseURL: "http://custom.repo/maven",
			},
		}

		mergeConfig(base, loaded)

		assert.Equal(t, "http://custom.repo/maven", base.Download.BaseURL)
	})

	t.Run("does not override with empty baseURL", func(t *testing.T) {
		base := DefaultConfig()
		loaded := &Config{
			Download: DownloadConfig{
				BaseURL: "",
			},
		}

		mergeConfig(base, loaded)

		assert.Equal(t, DefaultBaseURL, base.Download.BaseURL)
	})

	t.Run("merges proxy config", func(t *testing.T) {
		base := DefaultConfig()
		loaded := &Config{
			Download: DownloadConfig{
				Proxy: &ProxyConfig{
					URL:      "http://proxy:8080",
					Username: "user",
					Password: "pass",
					NoProxy:  "localhost",
				},
			},
		}

		mergeConfig(base, loaded)

		require.NotNil(t, base.Download.Proxy)
		assert.Equal(t, "http://proxy:8080", base.Download.Proxy.URL)
		assert.Equal(t, "user", base.Download.Proxy.Username)
		assert.Equal(t, "pass", base.Download.Proxy.Password)
		assert.Equal(t, "localhost", base.Download.Proxy.NoProxy)
	})
}

func TestLoadConfigFile(t *testing.T) {
	t.Run("loads valid config file", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")

		content := `download:
  baseUrl: http://custom.repo/maven
  proxy:
    url: http://proxy:8080
    username: testuser
    password: testpass
    noProxy: localhost,127.0.0.1
`
		err := os.WriteFile(configPath, []byte(content), 0644)
		require.NoError(t, err)

		config, err := loadConfigFile(configPath)
		require.NoError(t, err)

		assert.Equal(t, "http://custom.repo/maven", config.Download.BaseURL)
		require.NotNil(t, config.Download.Proxy)
		assert.Equal(t, "http://proxy:8080", config.Download.Proxy.URL)
		assert.Equal(t, "testuser", config.Download.Proxy.Username)
		assert.Equal(t, "testpass", config.Download.Proxy.Password)
		assert.Equal(t, "localhost,127.0.0.1", config.Download.Proxy.NoProxy)
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		_, err := loadConfigFile("/non/existent/path/config.yaml")
		assert.Error(t, err)
	})

	t.Run("returns error for invalid YAML", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")

		content := `invalid: yaml: content: [[[`
		err := os.WriteFile(configPath, []byte(content), 0644)
		require.NoError(t, err)

		_, err = loadConfigFile(configPath)
		assert.Error(t, err)
	})
}
