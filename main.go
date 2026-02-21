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

	// Parse CLI flags
	version := flag.String("version", "", "Version of the Moderne CLI to install (default: latest)")
	flag.Parse()

	fmt.Printf("Using configuration from: %s\n", configSource)

	// Determine version
	targetVersion := *version
	if targetVersion == "" {
		fmt.Println("No version specified, fetching latest version...")
		latest, err := FetchLatestVersion(config.Download.BaseURL, nil)
		if err != nil {
			fmt.Printf("Error: failed to determine latest version: %v\n", err)
			fmt.Println("Please specify a version using -version flag")
			os.Exit(1)
		}
		targetVersion = latest
		fmt.Printf("Latest version: %s\n", targetVersion)
	}

	installer := NewInstallerWithConfig(targetVersion, config)
	if err := installer.Run(); err != nil {
		fmt.Printf("Installation failed: %v\n", err)
		os.Exit(1)
	}
}
