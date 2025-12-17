# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.4] - 2025-12-17

### Fixed
- **API Key Flag Binding**: Fixed viper binding for `--api-key` flag
  - Changed viper key from `api-key` to `api_key` to match config struct
  - CLI flags now properly override config file values
  - Resolves authentication errors when using `--api-key` flag
- **Additional Lint Fixes**: Resolved remaining lint errors
  - Fixed `viper.BindPFlag` errcheck errors in root.go (3 instances)
  - Replaced deprecated `cobra.ExactValidArgs` with `cobra.MatchAll`
  - Updated golangci-lint configuration for compatibility
  - Applied `go fmt` formatting across all files

## [1.0.3] - 2025-12-17

### Fixed
- **Lint Errors**: Resolved all errcheck lint errors from watch mode implementation
  - Added proper error checking for `cmd.Run()` in watch.go
  - Added blank identifier for `MarkFlagRequired` return values in containers.go (7 instances)
  - Added blank identifier for `MarkFlagRequired` return values in images.go (2 instances)
  - All tests passing

## [1.0.2] - 2025-12-17

### Added
- **Watch Mode for Continuous Monitoring**: Real-time monitoring of resources with automatic refresh
  - Added `--watch` flag to `containers list`, `stacks list`, and `images list` commands
  - Added `--interval` flag to control refresh rate (default: 2 seconds)
  - Clear screen between updates for clean display
  - Show timestamp of last update
  - Graceful exit with Ctrl+C (SIGINT/SIGTERM handling)
  - Works with all output formats (table, JSON, YAML)
  - Cross-platform support (Linux, macOS, Windows)

### Technical Details
- Created `internal/watch` package for reusable watch functionality
- Context-based cancellation for proper cleanup
- Signal handling for graceful shutdown

## [1.0.1] - 2025-12-17

### Added
- **Stack Update Command**: New `stacks update` command to update existing stacks with new compose files
  - Update stack file content from local files
  - Update environment variables with `--env` flag
  - Requires `--endpoint` and `--file` flags

### Fixed
- **Lint Errors**: Resolved all errcheck lint errors in containers.go and client.go
  - Added proper error checking for all flag Get methods
  - Added type assertion checks in WithInsecure and WithCustomCA functions
- **Stack Update API**: Fixed hardcoded endpointId in stack update API call
  - Now properly accepts endpoint parameter
  - Aligns with Portainer API specification

## [1.0.0] - 2025-12-17

### Added

#### Core Features
- Multi-profile configuration management with secure credential storage
- JWT token and API key authentication support
- Flexible output formatting (table, JSON, YAML)
- Verbose and quiet modes for different use cases
- Shell completion support for bash, zsh, fish, and powershell

#### Environment Management
- List all Portainer environments with filtering
- Get detailed environment information by ID or name
- Environment status and health monitoring
- Support for Docker, Kubernetes, and Edge environments

#### Container Management
- List containers with filtering by environment and status
- Inspect container details and configuration
- View container logs with follow, tail, and since options
- Start, stop, restart, and remove containers
- Container lifecycle management with proper error handling

#### Stack Management
- List stacks across environments
- Deploy stacks from local files, Git repositories, or inline content
- Get detailed stack information and status
- Update and remove stacks with cleanup options
- Support for Docker Compose and Kubernetes stacks
- Environment variable injection in stack deployments

#### Image Management
- List Docker images with size and age information
- Inspect image details including layers and configuration
- Pull images from registries with authentication
- Remove images with force option
- Tag images with new names
- Prune unused and dangling images

#### Registry Management
- List configured container registries
- Get detailed registry information
- Delete registry configurations
- Support for Docker Hub, private registries, and custom registries

#### Volume Management
- List Docker volumes with driver and scope information
- Inspect volume details including usage data
- Create new volumes with custom drivers
- Remove volumes with force option
- Prune unused volumes

#### Network Management
- List Docker networks with driver information
- Inspect network details including IPAM configuration
- Create networks with custom drivers and options
- Remove networks
- Prune unused networks
- Support for internal and attachable network options

#### Configuration & Authentication
- Initialize configuration with `config init`
- Set and get configuration values
- Manage multiple profiles/contexts
- Create and switch between profiles
- Login and logout with username/password
- API key authentication support

### Technical Details
- Built with Go 1.21+
- Uses Cobra for CLI framework
- Uses Viper for configuration management
- Comprehensive error handling and validation
- HTTP client with retry logic and timeout configuration
- Proper HTTPS certificate validation
- Test coverage for core packages (62.1% config, 96.1% output)

### Documentation
- Comprehensive README with installation and usage instructions
- Command-line help for all commands
- Output formatting documentation
- Environment management documentation
- Configuration examples and best practices

[1.0.4]: https://github.com/robversluis/portainer-cli/releases/tag/v1.0.4
[1.0.3]: https://github.com/robversluis/portainer-cli/releases/tag/v1.0.3
[1.0.2]: https://github.com/robversluis/portainer-cli/releases/tag/v1.0.2
[1.0.1]: https://github.com/robversluis/portainer-cli/releases/tag/v1.0.1
[1.0.0]: https://github.com/robversluis/portainer-cli/releases/tag/v1.0.0
