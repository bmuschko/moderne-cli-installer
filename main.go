package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	version := flag.String("version", "", "Version of the Moderne CLI to install (required)")
	baseURL := flag.String("url", DefaultBaseURL, "Base URL for downloading the CLI JAR")
	flag.Parse()

	if *version == "" {
		fmt.Println("Error: version is required")
		fmt.Println("Usage: moderne-cli-installer -version <version> [-url <base-url>]")
		fmt.Println("Example: moderne-cli-installer -version 3.57.9")
		os.Exit(1)
	}

	installer := NewInstaller(*version, *baseURL)
	if err := installer.Run(); err != nil {
		fmt.Printf("Installation failed: %v\n", err)
		os.Exit(1)
	}
}
