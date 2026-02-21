package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
)

// MavenMetadata represents the maven-metadata.xml structure.
type MavenMetadata struct {
	XMLName    xml.Name   `xml:"metadata"`
	Versioning Versioning `xml:"versioning"`
}

// Versioning contains version information from Maven metadata.
type Versioning struct {
	Latest  string `xml:"latest"`
	Release string `xml:"release"`
}

// FetchLatestVersion fetches the latest version from Maven Central metadata.
func FetchLatestVersion(baseURL string, client *http.Client) (string, error) {
	metadataURL := fmt.Sprintf("%s/maven-metadata.xml", baseURL)

	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Get(metadataURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch metadata: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read metadata: %w", err)
	}

	var metadata MavenMetadata
	if err := xml.Unmarshal(body, &metadata); err != nil {
		return "", fmt.Errorf("failed to parse metadata: %w", err)
	}

	// Prefer <latest>, fall back to <release>
	version := metadata.Versioning.Latest
	if version == "" {
		version = metadata.Versioning.Release
	}

	if version == "" {
		return "", fmt.Errorf("no version found in metadata")
	}

	return version, nil
}
