package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// configureShellAlias sets up the shell alias for the mod command.
func (i *Installer) configureShellAlias() error {
	i.logger.Step("Configuring shell alias")

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

	shellConfigs := i.detectUnixShellConfigs(homeDir)

	for _, configFile := range shellConfigs {
		if err := i.updateShellConfig(configFile, markerComment, aliasLine); err != nil {
			i.logger.Warning("Failed to update %s: %v", configFile, err)
		} else {
			i.logger.Success("Updated %s", configFile)
		}
	}

	return nil
}

func (i *Installer) detectUnixShellConfigs(homeDir string) []string {
	var shellConfigs []string

	bashrc := filepath.Join(homeDir, ".bashrc")
	if _, err := os.Stat(bashrc); err == nil {
		shellConfigs = append(shellConfigs, bashrc)
	}

	zshrc := filepath.Join(homeDir, ".zshrc")
	if _, err := os.Stat(zshrc); err == nil {
		shellConfigs = append(shellConfigs, zshrc)
	}

	// If neither exists, default to .bashrc
	if len(shellConfigs) == 0 {
		shellConfigs = append(shellConfigs, bashrc)
	}

	return shellConfigs
}

func (i *Installer) configureWindowsAlias() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	if err := i.configurePowerShellProfile(homeDir); err != nil {
		i.logger.Warning("Failed to configure PowerShell: %v", err)
	}

	if err := i.createBatchFile(); err != nil {
		i.logger.Warning("Failed to create batch file: %v", err)
	}

	return nil
}

func (i *Installer) configurePowerShellProfile(homeDir string) error {
	psProfileDir := filepath.Join(homeDir, "Documents", "WindowsPowerShell")
	if err := os.MkdirAll(psProfileDir, 0755); err != nil {
		return fmt.Errorf("failed to create PowerShell profile directory: %w", err)
	}

	profilePath := filepath.Join(psProfileDir, "Microsoft.PowerShell_profile.ps1")
	functionDef := fmt.Sprintf(`function %s { java -jar "%s" $args }`, aliasName, i.jarPath)
	markerComment := "# Moderne CLI alias (managed by installer)"

	if err := i.updateShellConfig(profilePath, markerComment, functionDef); err != nil {
		return err
	}

	i.logger.Success("Updated PowerShell profile: %s", profilePath)
	return nil
}

func (i *Installer) createBatchFile() error {
	batchPath := filepath.Join(i.binDir, "mod.bat")
	batchContent := fmt.Sprintf("@echo off\njava -jar \"%s\" %%*\n", i.jarPath)

	if err := os.WriteFile(batchPath, []byte(batchContent), 0755); err != nil {
		return err
	}

	i.logger.Success("Created batch file: %s", batchPath)
	i.logger.Info("Add %s to your PATH to use 'mod' in CMD", i.binDir)
	return nil
}

func (i *Installer) updateShellConfig(configFile, marker, content string) error {
	existingContent, err := os.ReadFile(configFile)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	lines := strings.Split(string(existingContent), "\n")
	newLines := i.removeExistingAlias(lines, marker)

	// Add new alias
	newLines = append(newLines, "", marker, content)

	newContent := strings.Join(newLines, "\n")
	return os.WriteFile(configFile, []byte(newContent), 0644)
}

func (i *Installer) removeExistingAlias(lines []string, marker string) []string {
	var newLines []string
	skipNext := false

	for _, line := range lines {
		if strings.Contains(line, marker) {
			skipNext = true
			i.logger.Info("Replacing existing Moderne CLI alias")
			continue
		}
		if skipNext {
			skipNext = false
			continue
		}
		newLines = append(newLines, line)
	}

	return newLines
}
