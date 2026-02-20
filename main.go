package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	defaultBaseURL  = "https://repo1.maven.org/maven2/io/moderne/moderne-cli"
	installDirName  = ".moderne"
	binDirName      = "bin"
	jarFilePrefix   = "moderne-cli-"
	jarFileSuffix   = ".jar"
	aliasName       = "mod"
)

// ModCommand represents a command to run after installation
type ModCommand struct {
	Args        []string
	Description string
}

// Default mod commands to run after installation
// Customize this list based on your requirements
var postInstallCommands = []ModCommand{
	// Example commands - uncomment and modify as needed:
	// {Args: []string{"config", "license", "YOUR_LICENSE_KEY"}, Description: "Configuring license"},
	// {Args: []string{"config", "moderne", "https://app.moderne.io"}, Description: "Configuring Moderne platform URL"},
}

type Installer struct {
	version     string
	baseURL     string
	installDir  string
	binDir      string
	jarPath     string
	jarFileName string
}

func main() {
	version := flag.String("version", "", "Version of the Moderne CLI to install (required)")
	baseURL := flag.String("url", defaultBaseURL, "Base URL for downloading the CLI JAR")
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

func NewInstaller(version, baseURL string) *Installer {
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
		baseURL:     baseURL,
		installDir:  installDir,
		binDir:      binDir,
		jarPath:     jarPath,
		jarFileName: jarFileName,
	}
}

func (i *Installer) Run() error {
	i.logStep("Starting Moderne CLI installation")
	i.logInfo("Version: %s", i.version)
	i.logInfo("Install directory: %s", i.installDir)

	// Step 1: Create installation directories
	if err := i.createDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Step 2: Download the JAR file
	if err := i.downloadJAR(); err != nil {
		return fmt.Errorf("failed to download JAR: %w", err)
	}

	// Step 3: Configure shell alias
	if err := i.configureShellAlias(); err != nil {
		return fmt.Errorf("failed to configure shell alias: %w", err)
	}

	// Step 4: Run post-installation commands
	if err := i.runPostInstallCommands(); err != nil {
		return fmt.Errorf("failed to run post-install commands: %w", err)
	}

	i.printCompletionMessage()
	return nil
}

func (i *Installer) createDirectories() error {
	i.logStep("Creating installation directories")

	if err := os.MkdirAll(i.binDir, 0755); err != nil {
		return err
	}

	i.logSuccess("Created directory: %s", i.binDir)
	return nil
}

func (i *Installer) downloadJAR() error {
	i.logStep("Downloading Moderne CLI JAR")

	// Check if JAR already exists
	if _, err := os.Stat(i.jarPath); err == nil {
		i.logInfo("JAR file already exists at %s, skipping download", i.jarPath)
		return nil
	}

	// Construct download URL
	// Maven Central format: baseURL/version/moderne-cli-version.jar
	downloadURL := fmt.Sprintf("%s/%s/%s", i.baseURL, i.version, i.jarFileName)
	i.logInfo("Downloading from: %s", downloadURL)

	// Create HTTP request
	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	// Create the output file
	out, err := os.Create(i.jarPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// Download with progress
	contentLength := resp.ContentLength
	if contentLength > 0 {
		i.logInfo("File size: %.2f MB", float64(contentLength)/(1024*1024))
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
	fmt.Println() // New line after progress

	if err != nil {
		os.Remove(i.jarPath) // Clean up partial download
		return fmt.Errorf("failed to write file: %w", err)
	}

	i.logSuccess("Downloaded %.2f MB to %s", float64(written)/(1024*1024), i.jarPath)
	return nil
}

func (i *Installer) configureShellAlias() error {
	i.logStep("Configuring shell alias")

	switch runtime.GOOS {
	case "windows":
		return i.configureWindowsAlias()
	default:
		return i.configureUnixAlias()
	}
}

func (i *Installer) configureUnixAlias() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	aliasLine := fmt.Sprintf(`alias %s="java -jar %s"`, aliasName, i.jarPath)
	markerComment := "# Moderne CLI alias (managed by installer)"

	// Determine which shell config files to update
	shellConfigs := []string{}

	// Check for bash
	bashrc := filepath.Join(homeDir, ".bashrc")
	if _, err := os.Stat(bashrc); err == nil {
		shellConfigs = append(shellConfigs, bashrc)
	}

	// Check for zsh
	zshrc := filepath.Join(homeDir, ".zshrc")
	if _, err := os.Stat(zshrc); err == nil {
		shellConfigs = append(shellConfigs, zshrc)
	}

	// If neither exists, create .bashrc
	if len(shellConfigs) == 0 {
		shellConfigs = append(shellConfigs, bashrc)
	}

	for _, configFile := range shellConfigs {
		if err := i.updateShellConfig(configFile, markerComment, aliasLine); err != nil {
			i.logWarning("Failed to update %s: %v", configFile, err)
		} else {
			i.logSuccess("Updated %s", configFile)
		}
	}

	return nil
}

func (i *Installer) configureWindowsAlias() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// Create PowerShell profile directory if it doesn't exist
	psProfileDir := filepath.Join(homeDir, "Documents", "WindowsPowerShell")
	if err := os.MkdirAll(psProfileDir, 0755); err != nil {
		return fmt.Errorf("failed to create PowerShell profile directory: %w", err)
	}

	profilePath := filepath.Join(psProfileDir, "Microsoft.PowerShell_profile.ps1")

	// PowerShell function instead of alias (more flexible)
	functionDef := fmt.Sprintf(`function %s { java -jar "%s" $args }`, aliasName, i.jarPath)
	markerComment := "# Moderne CLI alias (managed by installer)"

	if err := i.updateShellConfig(profilePath, markerComment, functionDef); err != nil {
		i.logWarning("Failed to update PowerShell profile: %v", err)
	} else {
		i.logSuccess("Updated PowerShell profile: %s", profilePath)
	}

	// Also create a batch file for CMD
	batchPath := filepath.Join(i.binDir, "mod.bat")
	batchContent := fmt.Sprintf(`@echo off
java -jar "%s" %%*
`, i.jarPath)

	if err := os.WriteFile(batchPath, []byte(batchContent), 0755); err != nil {
		i.logWarning("Failed to create batch file: %v", err)
	} else {
		i.logSuccess("Created batch file: %s", batchPath)
		i.logInfo("Add %s to your PATH to use 'mod' in CMD", i.binDir)
	}

	return nil
}

func (i *Installer) updateShellConfig(configFile, marker, content string) error {
	// Read existing content
	existingContent, err := os.ReadFile(configFile)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	lines := strings.Split(string(existingContent), "\n")
	var newLines []string
	skipNext := false
	found := false

	// Remove old alias entries
	for _, line := range lines {
		if strings.Contains(line, marker) {
			skipNext = true
			found = true
			continue
		}
		if skipNext {
			skipNext = false
			continue
		}
		newLines = append(newLines, line)
	}

	// Add new alias
	newLines = append(newLines, "", marker, content)

	// Write back
	newContent := strings.Join(newLines, "\n")
	if err := os.WriteFile(configFile, []byte(newContent), 0644); err != nil {
		return err
	}

	if found {
		i.logInfo("Replaced existing Moderne CLI alias in %s", configFile)
	}

	return nil
}

func (i *Installer) runPostInstallCommands() error {
	if len(postInstallCommands) == 0 {
		i.logInfo("No post-installation commands configured")
		return nil
	}

	i.logStep("Running post-installation commands")

	// Determine how to run the mod command
	var modCmd string
	var modArgs []string

	if runtime.GOOS == "windows" {
		modCmd = "java"
		modArgs = []string{"-jar", i.jarPath}
	} else {
		modCmd = "java"
		modArgs = []string{"-jar", i.jarPath}
	}

	for _, cmd := range postInstallCommands {
		i.logInfo("%s...", cmd.Description)

		args := append(modArgs, cmd.Args...)
		execCmd := exec.Command(modCmd, args...)
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr

		if err := execCmd.Run(); err != nil {
			i.logWarning("Command failed: %v", err)
			// Continue with other commands even if one fails
		} else {
			i.logSuccess("Completed: %s", cmd.Description)
		}
	}

	return nil
}

func (i *Installer) printCompletionMessage() {
	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))
	i.logSuccess("Moderne CLI installation completed!")
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

// Logging helpers
func (i *Installer) logStep(format string, args ...interface{}) {
	fmt.Printf("\n[*] "+format+"\n", args...)
}

func (i *Installer) logInfo(format string, args ...interface{}) {
	fmt.Printf("    "+format+"\n", args...)
}

func (i *Installer) logSuccess(format string, args ...interface{}) {
	fmt.Printf("    [OK] "+format+"\n", args...)
}

func (i *Installer) logWarning(format string, args ...interface{}) {
	fmt.Printf("    [WARN] "+format+"\n", args...)
}

// progressReader wraps an io.Reader to report progress
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
