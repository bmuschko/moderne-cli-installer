package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	DefaultBaseURL = "https://repo1.maven.org/maven2/io/moderne/moderne-cli"
	installDirName = ".moderne"
	binDirName     = "bin"
	jarFilePrefix  = "moderne-cli-"
	jarFileSuffix  = ".jar"
	aliasName      = "mod"
)

// Installer manages the Moderne CLI installation process.
type Installer struct {
	version     string
	config      *Config
	installDir  string
	binDir      string
	jarPath     string
	jarFileName string
	logger      *Logger
}

// NewInstallerWithConfig creates a new Installer instance with the given config.
func NewInstallerWithConfig(version string, config *Config) *Installer {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Warning: could not determine home directory: %v\n", err)
		homeDir = "."
	}

	installDir := filepath.Join(homeDir, installDirName)
	binDir := filepath.Join(installDir, binDirName)
	jarFileName := fmt.Sprintf("%s%s%s", jarFilePrefix, version, jarFileSuffix)
	jarPath := filepath.Join(binDir, jarFileName)

	return &Installer{
		version:     version,
		config:      config,
		installDir:  installDir,
		binDir:      binDir,
		jarPath:     jarPath,
		jarFileName: jarFileName,
		logger:      NewLogger(),
	}
}

// Run executes the full installation process.
func (i *Installer) Run() error {
	i.logger.Step("Starting Moderne CLI installation")
	i.logger.Info("Version: %s", i.version)
	i.logger.Info("Download URL: %s", i.config.Download.BaseURL)
	i.logger.Info("Install directory: %s", i.installDir)

	if err := i.createDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	if err := i.downloadJAR(); err != nil {
		return fmt.Errorf("failed to download JAR: %w", err)
	}

	if err := i.configureShellAlias(); err != nil {
		return fmt.Errorf("failed to configure shell alias: %w", err)
	}

	if err := i.runPostInstallCommands(); err != nil {
		return fmt.Errorf("failed to run post-install commands: %w", err)
	}

	i.printCompletionMessage()
	return nil
}

func (i *Installer) createDirectories() error {
	i.logger.Step("Creating installation directories")

	if err := os.MkdirAll(i.binDir, 0755); err != nil {
		return err
	}

	i.logger.Success("Created directory: %s", i.binDir)
	return nil
}

func (i *Installer) printCompletionMessage() {
	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))
	i.logger.Success("Moderne CLI installation completed!")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()
	fmt.Println("JAR location:", i.jarPath)
	fmt.Println()

	switch runtime.GOOS {
	case "windows":
		fmt.Println("To use the 'mod' command:")
		fmt.Println("  - In PowerShell: Restart PowerShell or run:")
		fmt.Println("    . $PROFILE")
		fmt.Println("  - In CMD: Add the following to your PATH:")
		fmt.Printf("    %s\n", i.binDir)
	default:
		fmt.Println("To use the 'mod' command, restart your shell or run:")
		fmt.Println("  source ~/.bashrc")
		fmt.Println("  # or")
		fmt.Println("  source ~/.zshrc")
	}

	fmt.Println()
	fmt.Println("Then verify the installation:")
	fmt.Println("  mod --version")
	fmt.Println()
}
