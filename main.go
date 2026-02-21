package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	// Load configuration
	config, configSource, err := LoadConfig()
	if err != nil {
		fmt.Printf("Warning: failed to load config: %v\n", err)
		config = DefaultConfig()
		configSource = "defaults"
	}

	// Parse CLI flags (override config values)
	version := flag.String("version", "", "Version of the Moderne CLI to install (required)")
	baseURL := flag.String("url", "", "Base URL for downloading the CLI JAR")
	flag.Parse()

	if *version == "" {
		fmt.Println("Error: version is required")
		fmt.Println("Usage: moderne-cli-installer -version <version> [-url <base-url>]")
		fmt.Println("Example: moderne-cli-installer -version 3.57.9")
		os.Exit(1)
	}

	// CLI flag overrides config
	if *baseURL != "" {
		config.Download.BaseURL = *baseURL
	}

	fmt.Printf("Using configuration from: %s\n", configSource)

	installer := NewInstallerWithConfig(*version, config)
	if err := installer.Run(); err != nil {
		fmt.Printf("Installation failed: %v\n", err)
		os.Exit(1)
	}
}
