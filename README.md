# Moderne CLI Installer

A cross-platform installer for the [Moderne CLI](https://docs.moderne.io/moderne-cli/getting-started/moderne-cli-intro). Downloads and configures the CLI JAR, sets up shell aliases, and runs post-installation commands.

## Features

- Cross-platform support (Windows, macOS, Linux)
- Automatic latest version detection from Maven Central
- Configurable download source (Maven Central, Artifactory, or custom HTTP server)
- Proxy support with authentication
- Shell alias configuration (bash, zsh, PowerShell, CMD)
- Customizable post-installation commands

## Usage

### Basic Usage

```bash
# Install the latest version (auto-detected from Maven Central)
./moderne-cli-installer

# Install a specific version
./moderne-cli-installer -version 3.57.9
```

### Command Line Options

| Flag | Description | Default |
|------|-------------|---------|
| `-version` | Version to install | Latest (auto-detected) |

### Examples

```bash
# Install latest version
./moderne-cli-installer

# Install specific version
./moderne-cli-installer -version 3.57.9
```

## Configuration

The installer supports a YAML configuration file for advanced settings. Place a `config.yaml` file either:
- Next to the installer binary, or
- In the current working directory

### Configuration File

```yaml
download:
  # Base URL for downloading the CLI JAR
  # Default: https://repo1.maven.org/maven2/io/moderne/moderne-cli
  baseUrl: https://repo1.maven.org/maven2/io/moderne/moderne-cli

  # Proxy settings (optional)
  proxy:
    url: http://proxy.example.com:8080
    username: proxyuser
    password: proxypass
    noProxy: localhost,127.0.0.1,.internal.domain
```

### Configuration Options

#### Download Settings

| Option | Description | Required |
|--------|-------------|----------|
| `download.baseUrl` | Base URL for the Maven repository | No (defaults to Maven Central) |
| `download.proxy.url` | HTTP proxy URL | No |
| `download.proxy.username` | Proxy authentication username | No |
| `download.proxy.password` | Proxy authentication password | No |
| `download.proxy.noProxy` | Comma-separated list of hosts to bypass proxy | No |

### Using with Different Repository Types

#### Maven Central (default)

No configuration needed. The installer uses Maven Central by default.

#### Artifactory

```yaml
download:
  baseUrl: https://artifactory.example.com/artifactory/libs-release/io/moderne/moderne-cli
```

Note: Auto-version detection works with Artifactory as it generates `maven-metadata.xml`.

#### Simple HTTP Server

For non-Maven repositories, configure the base URL and specify the version explicitly:

```yaml
download:
  baseUrl: https://files.example.com/moderne-cli
```

```bash
./moderne-cli-installer -version 3.57.9
```

The JAR must be available at: `<baseUrl>/<version>/moderne-cli-<version>.jar`

Note: Auto-version detection requires `maven-metadata.xml`, so you must use `-version` with simple HTTP servers.

## Post-Installation Commands

The installer can run commands automatically after installation. Create a `post-install-commands.txt` file next to the installer binary or in the current working directory.

### File Format

```bash
# Lines starting with # are comments
# Empty lines are ignored
# Each line is executed as a shell command

# The $MOD variable is pre-defined as "java -jar <path-to-jar>"
$MOD config license YOUR_LICENSE_KEY
$MOD config moderne https://app.moderne.io

# You can also run any shell command
echo "Moderne CLI installed successfully"
```

### Command Execution

- **Unix (Linux/macOS)**: Commands run via `bash -c`
- **Windows**: Commands run via PowerShell

The `$MOD` variable is automatically set to `java -jar <path-to-jar>`, allowing you to run Moderne CLI commands without knowing the exact JAR path.

### Example Commands

```bash
# Configure license
$MOD config license YOUR_LICENSE_KEY

# Configure Moderne platform URL
$MOD config moderne https://app.moderne.io

# Configure Java home
$MOD config java home /usr/lib/jvm/java-17

# Build a project
$MOD build /path/to/project
```

## Installation Directory

The CLI JAR is installed to:

| Platform | Location |
|----------|----------|
| Unix (Linux/macOS) | `~/.moderne/bin/moderne-cli-<version>.jar` |
| Windows | `%USERPROFILE%\.moderne\bin\moderne-cli-<version>.jar` |

## Shell Alias

The installer configures a `mod` alias/function:

| Shell | Configuration File |
|-------|-------------------|
| Bash | `~/.bashrc` |
| Zsh | `~/.zshrc` |
| PowerShell | `~/Documents/WindowsPowerShell/Microsoft.PowerShell_profile.ps1` |
| CMD | `mod.bat` in the bin directory (add to PATH) |

After installation, restart your shell or source the configuration file:

```bash
# Bash
source ~/.bashrc

# Zsh
source ~/.zshrc

# PowerShell
. $PROFILE
```

Then verify:

```bash
mod --version
```

## Building from Source

### Prerequisites

- Go 1.24 or later

### Build

```bash
# Build for current platform
go build -o moderne-cli-installer .

# Build for all platforms
./build.sh
```

### Run Tests

```bash
go test -v ./...
```

## License

[Add your license here]
