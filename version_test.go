package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchLatestVersion(t *testing.T) {
	t.Run("returns latest version from metadata", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/maven-metadata.xml", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<metadata>
  <groupId>io.moderne</groupId>
  <artifactId>moderne-cli</artifactId>
  <versioning>
    <latest>3.57.9</latest>
    <release>3.57.8</release>
  </versioning>
</metadata>`))
		}))
		defer server.Close()

		version, err := FetchLatestVersion(server.URL, nil)
		require.NoError(t, err)
		assert.Equal(t, "3.57.9", version)
	})

	t.Run("falls back to release when latest is empty", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<metadata>
  <versioning>
    <latest></latest>
    <release>3.57.8</release>
  </versioning>
</metadata>`))
		}))
		defer server.Close()

		version, err := FetchLatestVersion(server.URL, nil)
		require.NoError(t, err)
		assert.Equal(t, "3.57.8", version)
	})

	t.Run("returns error when no version found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<metadata>
  <versioning>
  </versioning>
</metadata>`))
		}))
		defer server.Close()

		_, err := FetchLatestVersion(server.URL, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no version found")
	})

	t.Run("returns error on HTTP failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		_, err := FetchLatestVersion(server.URL, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to fetch metadata")
	})

	t.Run("returns error on invalid XML", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`not valid xml`))
		}))
		defer server.Close()

		_, err := FetchLatestVersion(server.URL, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse metadata")
	})

	t.Run("returns error on connection failure", func(t *testing.T) {
		_, err := FetchLatestVersion("http://localhost:99999", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to fetch metadata")
	})

	t.Run("uses provided HTTP client", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<metadata>
  <versioning>
    <latest>1.0.0</latest>
  </versioning>
</metadata>`))
		}))
		defer server.Close()

		customClient := &http.Client{}
		version, err := FetchLatestVersion(server.URL, customClient)
		require.NoError(t, err)
		assert.Equal(t, "1.0.0", version)
	})
}
