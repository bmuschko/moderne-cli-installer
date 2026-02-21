package main

import (
	"bufio"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

//go:embed commands.txt
var embeddedCommands embed.FS

const commandsFileName = "commands.txt"

// runPostInstallCommands executes configured commands after installation.
func (i *Installer) runPostInstallCommands() error {
	commands, source := i.loadCommands()

	if len(commands) == 0 {
		i.logger.Info("No post-installation commands configured")
		return nil
	}

	i.logger.Step("Running post-installation commands")
	i.logger.Info("Loaded %d command(s) from %s", len(commands), source)

	for _, cmdLine := range commands {
		if err := i.executeCommand(cmdLine); err != nil {
			i.logger.Warning("Command '%s' failed: %v", cmdLine, err)
		} else {
			i.logger.Success("Executed: %s", cmdLine)
		}
	}

	return nil
}

// loadCommands tries to load commands from an external file first,
// then falls back to the embedded file.
func (i *Installer) loadCommands() ([]string, string) {
	// Try external file first (next to the binary)
	exePath, err := os.Executable()
	if err == nil {
		externalPath := filepath.Join(filepath.Dir(exePath), commandsFileName)
		if commands, err := i.parseCommandsFile(externalPath); err == nil && len(commands) > 0 {
			return commands, externalPath
		}
	}

	// Try external file in current working directory
	if commands, err := i.parseCommandsFile(commandsFileName); err == nil && len(commands) > 0 {
		cwd, _ := os.Getwd()
		return commands, filepath.Join(cwd, commandsFileName)
	}

	// Fall back to embedded file
	if commands, err := i.parseEmbeddedCommands(); err == nil && len(commands) > 0 {
		return commands, "embedded"
	}

	return nil, ""
}

func (i *Installer) parseCommandsFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return i.parseCommands(bufio.NewScanner(file))
}

func (i *Installer) parseEmbeddedCommands() ([]string, error) {
	file, err := embeddedCommands.Open(commandsFileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return i.parseCommands(bufio.NewScanner(file))
}

func (i *Installer) parseCommands(scanner *bufio.Scanner) ([]string, error) {
	var commands []string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		commands = append(commands, line)
	}

	return commands, scanner.Err()
}

// executeCommand runs a command through the shell with MOD variable defined.
func (i *Installer) executeCommand(cmdLine string) error {
	var cmd *exec.Cmd

	// Define MOD as "java -jar <path>" so commands can use $MOD or %MOD%
	modValue := fmt.Sprintf("java -jar %s", i.jarPath)

	if runtime.GOOS == "windows" {
		// PowerShell: define $env:MOD and run command
		script := fmt.Sprintf(`$env:MOD = '%s'; %s`, modValue, cmdLine)
		cmd = exec.Command("powershell", "-NoProfile", "-Command", script)
	} else {
		// Bash: define MOD and run command
		script := fmt.Sprintf(`MOD="%s"; %s`, modValue, cmdLine)
		cmd = exec.Command("bash", "-c", script)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	return cmd.Run()
}
