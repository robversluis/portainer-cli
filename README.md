# Portainer CLI

A powerful command-line interface tool for managing Portainer environments, written in Go. This single-binary tool provides comprehensive access to Portainer's API, enabling automation and scriptable container orchestration workflows.

## Features

- **Single Binary**: Zero dependencies, cross-platform support (Linux, macOS, Windows)
- **Complete API Coverage**: Access all Portainer functionality from the command line
- **Multiple Authentication Methods**: Username/password, API keys, and OAuth support
- **Flexible Output**: Human-readable tables, JSON, and YAML formats
- **Multi-Environment Support**: Manage multiple Portainer instances with profiles
- **Scriptable**: Perfect for CI/CD pipelines and automation

## Installation

### Download Pre-built Binaries

Download the latest release for your platform from the [releases page](https://github.com/rob/portainer-cli/releases).

#### Linux/macOS
```bash
# Download and extract
tar -xzf portainer-cli_<version>_<os>_<arch>.tar.gz

# Move to PATH
sudo mv portainer-cli /usr/local/bin/

# Verify installation
portainer-cli --version
```

#### Windows
Download the zip file, extract it, and add the binary to your PATH.

### Build from Source

Requires Go 1.21 or later.

```bash
# Clone the repository
git clone https://github.com/rob/portainer-cli.git
cd portainer-cli

# Build
make build

# Install to GOPATH/bin
make install
```

## Quick Start

### Authentication

```bash
# Login with username and password
portainer-cli auth login --url https://portainer.example.com --username admin

# Use API key
portainer-cli --api-key YOUR_API_KEY environments list

# Check authentication status
portainer-cli auth status
```

### Basic Commands

```bash
# List environments
portainer-cli environments list

# List containers
portainer-cli containers list --endpoint 1

# Deploy a stack
portainer-cli stacks deploy --file docker-compose.yml --endpoint 1 --name mystack

# View container logs
portainer-cli containers logs my-container --follow

# List images
portainer-cli images list --endpoint 1
```

## Configuration

The CLI uses a configuration file located at `~/.portainer-cli/config.yaml` (or `$XDG_CONFIG_HOME/portainer-cli/config.yaml` on Unix systems).

### Multiple Profiles

```yaml
current_profile: production

profiles:
  production:
    url: https://portainer.prod.example.com
    api_key: prod_api_key_here
  
  staging:
    url: https://portainer.staging.example.com
    api_key: staging_api_key_here
```

Switch profiles:
```bash
portainer-cli --profile staging environments list
```

### Environment Variables

Override configuration with environment variables:
- `PORTAINER_URL`: Portainer server URL
- `PORTAINER_API_KEY`: API key for authentication
- `PORTAINER_USERNAME`: Username for authentication
- `PORTAINER_PASSWORD`: Password for authentication

## Command Reference

### Global Flags

- `--config`: Path to config file
- `--profile`: Profile/context to use
- `--url`: Portainer URL (override config)
- `--api-key`: API key (override config)
- `--output, -o`: Output format (table, json, yaml)
- `--verbose, -v`: Verbose output
- `--quiet, -q`: Quiet mode
- `--help, -h`: Help information
- `--version`: Show version

### Available Commands

- `auth`: Authentication operations (login, logout, status)
- `environments`: Manage Portainer environments/endpoints
- `containers`: Docker container operations
- `stacks`: Stack deployment and management
- `images`: Docker image operations
- `registries`: Registry management
- `kubernetes`: Kubernetes cluster operations
- `edge`: Edge computing operations
- `users`: User management
- `teams`: Team management
- `templates`: Custom template management
- `backup`: Backup and restore operations
- `system`: System information and settings

Run `portainer-cli <command> --help` for detailed command information.

## Development

### Prerequisites

- Go 1.21 or later
- Make
- golangci-lint (for linting)
- goreleaser (for releases)

### Building

```bash
# Install dependencies
make deps

# Build binary
make build

# Run tests
make test

# Run linter
make lint

# Build for all platforms
make build-all
```

### Project Structure

```
portainer-cli/
├── cmd/portainer-cli/    # Main application entry point
├── internal/             # Internal packages
│   ├── client/          # HTTP API client
│   ├── config/          # Configuration management
│   └── output/          # Output formatters
├── pkg/                 # Public packages
│   └── api/            # API models and types
├── docs/               # Documentation
├── scripts/            # Build and utility scripts
└── .taskmaster/        # Task management
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific package tests
go test ./internal/client/...
```

### Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run `make verify` to ensure tests and linting pass
6. Submit a pull request

## License

[Add your license here]

## Support

- **Issues**: [GitHub Issues](https://github.com/rob/portainer-cli/issues)
- **Documentation**: [Full documentation](https://github.com/rob/portainer-cli/docs)

## Roadmap

See the [PRD document](.taskmaster/docs/prd.txt) for detailed feature roadmap and future enhancements.

### Upcoming Features

- Interactive mode with prompts
- Shell completion (bash, zsh, fish)
- Colored output
- Progress bars for long operations
- Watch mode for continuous monitoring
- Plugin system for extensions
