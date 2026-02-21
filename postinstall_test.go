package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCommands(t *testing.T) {
	installer := &Installer{logger: NewLogger()}

	t.Run("parses simple commands", func(t *testing.T) {
		input := `config license KEY123
config moderne https://app.moderne.io`

		scanner := bufio.NewScanner(strings.NewReader(input))
		commands, err := installer.parseCommands(scanner)
		require.NoError(t, err)

		assert.Len(t, commands, 2)
		assert.Equal(t, "config license KEY123", commands[0])
		assert.Equal(t, "config moderne https://app.moderne.io", commands[1])
	})

	t.Run("skips empty lines", func(t *testing.T) {
		input := `command1

command2

`
		scanner := bufio.NewScanner(strings.NewReader(input))
		commands, err := installer.parseCommands(scanner)
		require.NoError(t, err)

		assert.Len(t, commands, 2)
		assert.Equal(t, "command1", commands[0])
		assert.Equal(t, "command2", commands[1])
	})

	t.Run("skips comment lines", func(t *testing.T) {
		input := `# This is a comment
command1
# Another comment
command2`

		scanner := bufio.NewScanner(strings.NewReader(input))
		commands, err := installer.parseCommands(scanner)
		require.NoError(t, err)

		assert.Len(t, commands, 2)
		assert.Equal(t, "command1", commands[0])
		assert.Equal(t, "command2", commands[1])
	})

	t.Run("trims whitespace", func(t *testing.T) {
		input := `  command1
	command2	`

		scanner := bufio.NewScanner(strings.NewReader(input))
		commands, err := installer.parseCommands(scanner)
		require.NoError(t, err)

		assert.Len(t, commands, 2)
		assert.Equal(t, "command1", commands[0])
		assert.Equal(t, "command2", commands[1])
	})

	t.Run("handles empty input", func(t *testing.T) {
		input := ``

		scanner := bufio.NewScanner(strings.NewReader(input))
		commands, err := installer.parseCommands(scanner)
		require.NoError(t, err)

		assert.Len(t, commands, 0)
	})

	t.Run("handles only comments and empty lines", func(t *testing.T) {
		input := `# Comment 1
# Comment 2

`
		scanner := bufio.NewScanner(strings.NewReader(input))
		commands, err := installer.parseCommands(scanner)
		require.NoError(t, err)

		assert.Len(t, commands, 0)
	})
}

func TestParseCommandsFile(t *testing.T) {
	installer := &Installer{logger: NewLogger()}

	t.Run("parses file successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "commands.txt")

		content := `# Post-install commands
$MOD config license KEY123
echo "Done"
`
		err := os.WriteFile(filePath, []byte(content), 0644)
		require.NoError(t, err)

		commands, err := installer.parseCommandsFile(filePath)
		require.NoError(t, err)

		assert.Len(t, commands, 2)
		assert.Equal(t, "$MOD config license KEY123", commands[0])
		assert.Equal(t, `echo "Done"`, commands[1])
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		_, err := installer.parseCommandsFile("/non/existent/file.txt")
		assert.Error(t, err)
	})
}

func TestLoadCommands(t *testing.T) {
	t.Run("loads from external file in cwd", func(t *testing.T) {
		// Save current directory
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer os.Chdir(originalDir)

		// Create temp directory and change to it
		tmpDir := t.TempDir()
		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		// Create commands file
		content := `command1
command2`
		err = os.WriteFile("post-install-commands.txt", []byte(content), 0644)
		require.NoError(t, err)

		installer := &Installer{logger: NewLogger()}
		commands, source := installer.loadCommands()

		assert.Len(t, commands, 2)
		assert.Contains(t, source, "post-install-commands.txt")
	})

	t.Run("falls back to embedded when no external file", func(t *testing.T) {
		// Save current directory
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer os.Chdir(originalDir)

		// Create temp directory with no commands file
		tmpDir := t.TempDir()
		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		installer := &Installer{logger: NewLogger()}
		commands, source := installer.loadCommands()

		// Embedded file has no commands (only comments)
		assert.Len(t, commands, 0)
		assert.Equal(t, "", source)
	})
}
