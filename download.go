package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// downloadJAR downloads the Moderne CLI JAR file.
func (i *Installer) downloadJAR() error {
	i.logger.Step("Downloading Moderne CLI JAR")

	// Check if JAR already exists
	if _, err := os.Stat(i.jarPath); err == nil {
		i.logger.Info("JAR file already exists at %s, skipping download", i.jarPath)
		return nil
	}

	// Construct download URL (Maven Central format: baseURL/version/moderne-cli-version.jar)
	downloadURL := fmt.Sprintf("%s/%s/%s", i.config.Download.BaseURL, i.version, i.jarFileName)
	i.logger.Info("Downloading from: %s", downloadURL)

	// Create HTTP client with optional proxy
	client, err := i.createHTTPClient()
	if err != nil {
		return fmt.Errorf("failed to create HTTP client: %w", err)
	}

	resp, err := client.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	out, err := os.Create(i.jarPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	contentLength := resp.ContentLength
	if contentLength > 0 {
		i.logger.Info("File size: %.2f MB", float64(contentLength)/(1024*1024))
	}

	written, err := io.Copy(out, &progressReader{
		reader: resp.Body,
		total:  contentLength,
		onProgress: func(downloaded, total int64) {
			if total > 0 {
				percent := float64(downloaded) / float64(total) * 100
				fmt.Printf("\r    Downloading: %.1f%%", percent)
			}
		},
	})
	fmt.Println()

	if err != nil {
		os.Remove(i.jarPath)
		return fmt.Errorf("failed to write file: %w", err)
	}

	i.logger.Success("Downloaded %.2f MB to %s", float64(written)/(1024*1024), i.jarPath)
	return nil
}

// createHTTPClient creates an HTTP client with optional proxy configuration.
func (i *Installer) createHTTPClient() (*http.Client, error) {
	if !i.config.Download.HasProxy() {
		return http.DefaultClient, nil
	}

	proxy := i.config.Download.Proxy
	i.logger.Info("Using proxy: %s", proxy.URL)

	proxyURL, err := url.Parse(proxy.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URL: %w", err)
	}

	// Add authentication if provided
	if proxy.Username != "" {
		proxyURL.User = url.UserPassword(proxy.Username, proxy.Password)
	}

	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			// Check noProxy list
			if proxy.NoProxy != "" {
				host := req.URL.Hostname()
				for _, noProxyHost := range strings.Split(proxy.NoProxy, ",") {
					noProxyHost = strings.TrimSpace(noProxyHost)
					if noProxyHost == "" {
						continue
					}
					// Match exact host or suffix (e.g., .internal.domain)
					if host == noProxyHost || strings.HasSuffix(host, noProxyHost) {
						return nil, nil // No proxy for this host
					}
				}
			}
			return proxyURL, nil
		},
	}

	return &http.Client{Transport: transport}, nil
}

// progressReader wraps an io.Reader to report download progress.
type progressReader struct {
	reader     io.Reader
	total      int64
	downloaded int64
	onProgress func(downloaded, total int64)
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.downloaded += int64(n)
	if pr.onProgress != nil {
		pr.onProgress(pr.downloaded, pr.total)
	}
	return n, err
}
