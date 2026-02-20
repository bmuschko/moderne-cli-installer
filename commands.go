package main

import (
	"os"
	"os/exec"
)

// ModCommand represents a command to run after installation.
type ModCommand struct {
	Args        []string
	Description string
}

// PostInstallCommands defines the commands to run after installation.
// Customize this list based on your requirements.
var PostInstallCommands = []ModCommand{
	// Example commands - uncomment and modify as needed:
	// {Args: []string{"config", "license", "YOUR_LICENSE_KEY"}, Description: "Configuring license"},
	// {Args: []string{"config", "moderne", "https://app.moderne.io"}, Description: "Configuring Moderne platform URL"},
}

// runPostInstallCommands executes configured commands after installation.
func (i *Installer) runPostInstallCommands() error {
	if len(PostInstallCommands) == 0 {
		i.logger.Info("No post-installation commands configured")
		return nil
	}

	i.logger.Step("Running post-installation commands")

	for _, cmd := range PostInstallCommands {
		if err := i.executeModCommand(cmd); err != nil {
			i.logger.Warning("Command failed: %v", err)
			// Continue with other commands even if one fails
		} else {
			i.logger.Success("Completed: %s", cmd.Description)
		}
	}

	return nil
}

func (i *Installer) executeModCommand(cmd ModCommand) error {
	i.logger.Info("%s...", cmd.Description)

	args := append([]string{"-jar", i.jarPath}, cmd.Args...)
	execCmd := exec.Command("java", args...)
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	return execCmd.Run()
}
