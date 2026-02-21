package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateHTTPClient(t *testing.T) {
	t.Run("returns default client when no proxy configured", func(t *testing.T) {
		installer := &Installer{
			config: &Config{
				Download: DownloadConfig{
					BaseURL: "http://example.com",
				},
			},
			logger: NewLogger(),
		}

		client, err := installer.createHTTPClient()
		require.NoError(t, err)
		assert.Equal(t, http.DefaultClient, client)
	})

	t.Run("returns custom client with proxy", func(t *testing.T) {
		installer := &Installer{
			config: &Config{
				Download: DownloadConfig{
					BaseURL: "http://example.com",
					Proxy: &ProxyConfig{
						URL: "http://proxy:8080",
					},
				},
			},
			logger: NewLogger(),
		}

		client, err := installer.createHTTPClient()
		require.NoError(t, err)
		assert.NotEqual(t, http.DefaultClient, client)
	})

	t.Run("returns error for invalid proxy URL", func(t *testing.T) {
		installer := &Installer{
			config: &Config{
				Download: DownloadConfig{
					BaseURL: "http://example.com",
					Proxy: &ProxyConfig{
						URL: "://invalid-url",
					},
				},
			},
			logger: NewLogger(),
		}

		_, err := installer.createHTTPClient()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid proxy URL")
	})
}

func TestDownloadJAR(t *testing.T) {
	t.Run("downloads JAR successfully", func(t *testing.T) {
		jarContent := []byte("fake jar content")
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/1.0.0/moderne-cli-1.0.0.jar", r.URL.Path)
			w.Header().Set("Content-Length", "16")
			w.WriteHeader(http.StatusOK)
			w.Write(jarContent)
		}))
		defer server.Close()

		tmpDir := t.TempDir()
		binDir := filepath.Join(tmpDir, ".moderne", "bin")

		installer := &Installer{
			version:     "1.0.0",
			config:      &Config{Download: DownloadConfig{BaseURL: server.URL}},
			binDir:      binDir,
			jarPath:     filepath.Join(binDir, "moderne-cli-1.0.0.jar"),
			jarFileName: "moderne-cli-1.0.0.jar",
			logger:      NewLogger(),
		}

		// Create bin directory
		err := os.MkdirAll(binDir, 0755)
		require.NoError(t, err)

		err = installer.downloadJAR()
		require.NoError(t, err)

		// Verify file was created
		content, err := os.ReadFile(installer.jarPath)
		require.NoError(t, err)
		assert.Equal(t, jarContent, content)
	})

	t.Run("skips download if JAR already exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		binDir := filepath.Join(tmpDir, ".moderne", "bin")
		jarPath := filepath.Join(binDir, "moderne-cli-1.0.0.jar")

		// Create existing JAR
		err := os.MkdirAll(binDir, 0755)
		require.NoError(t, err)
		err = os.WriteFile(jarPath, []byte("existing"), 0644)
		require.NoError(t, err)

		serverCalled := false
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			serverCalled = true
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		installer := &Installer{
			version:     "1.0.0",
			config:      &Config{Download: DownloadConfig{BaseURL: server.URL}},
			binDir:      binDir,
			jarPath:     jarPath,
			jarFileName: "moderne-cli-1.0.0.jar",
			logger:      NewLogger(),
		}

		err = installer.downloadJAR()
		require.NoError(t, err)
		assert.False(t, serverCalled, "server should not have been called")
	})

	t.Run("returns error on HTTP failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		tmpDir := t.TempDir()
		binDir := filepath.Join(tmpDir, ".moderne", "bin")
		err := os.MkdirAll(binDir, 0755)
		require.NoError(t, err)

		installer := &Installer{
			version:     "1.0.0",
			config:      &Config{Download: DownloadConfig{BaseURL: server.URL}},
			binDir:      binDir,
			jarPath:     filepath.Join(binDir, "moderne-cli-1.0.0.jar"),
			jarFileName: "moderne-cli-1.0.0.jar",
			logger:      NewLogger(),
		}

		err = installer.downloadJAR()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "download failed with status")
	})
}

func TestProgressReader(t *testing.T) {
	t.Run("reports progress correctly", func(t *testing.T) {
		data := []byte("hello world")
		var progressCalls []int64

		pr := &progressReader{
			reader: &mockReader{data: data},
			total:  int64(len(data)),
			onProgress: func(downloaded, total int64) {
				progressCalls = append(progressCalls, downloaded)
			},
		}

		buf := make([]byte, 5)
		n, _ := pr.Read(buf)
		assert.Equal(t, 5, n)
		assert.Equal(t, []int64{5}, progressCalls)

		n, _ = pr.Read(buf)
		assert.Equal(t, 5, n)
		assert.Equal(t, []int64{5, 10}, progressCalls)
	})
}

type mockReader struct {
	data   []byte
	offset int
}

func (m *mockReader) Read(p []byte) (int, error) {
	if m.offset >= len(m.data) {
		return 0, nil
	}
	n := copy(p, m.data[m.offset:])
	m.offset += n
	return n, nil
}
