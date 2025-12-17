# CLI Command Structure

This document outlines the command structure implemented in the Portainer CLI.

## Global Flags

Available on all commands:

- `--config`: Path to config file (default: `$HOME/.portainer-cli/config.yaml`)
- `--profile`: Profile/context to use
- `--url`: Portainer URL (overrides config)
- `--api-key`: API key for authentication (overrides config)
- `--output, -o`: Output format (table, json, yaml)
- `--verbose, -v`: Verbose output
- `--quiet, -q`: Quiet mode (minimal output)

## Command Hierarchy

```
portainer-cli
├── version                    # Display version information
├── auth                       # Authentication operations
│   ├── login                 # Login to Portainer
│   ├── logout                # Logout from Portainer
│   └── status                # Check authentication status
├── environments (env)         # Manage environments
│   ├── list (ls)             # List all environments
│   └── get [id]              # Get environment details
├── containers                 # Manage Docker containers
│   ├── list (ls)             # List containers
│   └── logs [container]      # View container logs
└── stacks                     # Manage stacks
    ├── list (ls)             # List stacks
    └── deploy                # Deploy a stack
```

## Implementation Status

### Task 2 - Completed ✓
- Root command with Cobra framework
- Global flags configuration
- Viper integration for configuration management
- Version command with build-time injection
- Placeholder commands for future implementation:
  - auth (login, logout, status)
  - environments (list, get)
  - containers (list, logs)
  - stacks (list, deploy)

### Future Tasks
- Task 3: Configuration management implementation
- Task 4: HTTP API client and authentication
- Task 5: Output formatting system
- Task 6-10: Full command implementations

## Configuration Priority

The CLI follows this priority order for configuration values:

1. Command-line flags (highest priority)
2. Environment variables (PORTAINER_*)
3. Config file values
4. Profile-specific values
5. Default values (lowest priority)

## Environment Variables

- `PORTAINER_URL`: Portainer server URL
- `PORTAINER_API_KEY`: API key for authentication
- `PORTAINER_USERNAME`: Username for authentication
- `PORTAINER_PASSWORD`: Password for authentication

## Example Usage

```bash
# Display version
portainer-cli version

# Use specific profile
portainer-cli --profile production environments list

# Override URL via flag
portainer-cli --url https://portainer.example.com auth status

# JSON output
portainer-cli --output json environments list

# Verbose mode
portainer-cli -v containers list --endpoint 1
```
