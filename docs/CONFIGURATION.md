# Configuration Management

This document describes the configuration system for Portainer CLI.

## Configuration File Location

The CLI stores configuration in a YAML file at:

- **Linux/macOS**: `~/.portainer-cli/config.yaml`
- **With XDG_CONFIG_HOME**: `$XDG_CONFIG_HOME/portainer-cli/config.yaml`

The configuration directory is created with `0700` permissions and the config file with `0600` permissions for security.

## Configuration Structure

```yaml
current_profile: production

profiles:
  production:
    url: https://portainer.prod.example.com
    api_key: your-api-key-here
    insecure: false
  
  staging:
    url: https://portainer.staging.example.com
    username: admin
    token: jwt-token-here
    insecure: false
  
  development:
    url: http://localhost:9000
    api_key: dev-key
    insecure: true
```

## Profile Fields

- **url** (required): Portainer server URL
- **api_key** (optional): API key for authentication
- **username** (optional): Username for authentication
- **token** (optional): JWT token for authentication
- **insecure** (optional): Skip TLS certificate verification (default: false)

At least one authentication method (api_key, username, or token) is required.

## Configuration Commands

### Initialize Configuration

```bash
portainer-cli config init
```

Creates the configuration directory and file.

### Create a Profile

```bash
# Create with API key
portainer-cli config create-profile production \
  --url https://portainer.example.com \
  --api-key YOUR_API_KEY

# Create with username
portainer-cli config create-profile staging \
  --url https://staging.example.com \
  --username admin

# Create with insecure flag
portainer-cli config create-profile local \
  --url http://localhost:9000 \
  --api-key local-key \
  --insecure
```

### Set Configuration Values

```bash
# Set for current profile
portainer-cli config set url https://new-url.example.com

# Set for specific profile
portainer-cli config set --profile production api_key NEW_KEY

# Available keys: url, api_key, username, token, insecure
portainer-cli config set insecure true
```

### Get Configuration Values

```bash
# Get all values for current profile
portainer-cli config get

# Get specific value
portainer-cli config get url

# Get for specific profile
portainer-cli config get --profile production url
```

### List Profiles

```bash
portainer-cli config list-profiles
# or
portainer-cli config profiles
```

Output:
```
Profile      URL                              Auth Method  Current
production   https://portainer.prod.com       API Key      *
staging      https://portainer.staging.com    Token        
development  http://localhost:9000            API Key      
```

### Switch Profile

```bash
portainer-cli config use-profile staging
```

### Delete Profile

```bash
portainer-cli config delete-profile old-profile
# or
portainer-cli config remove-profile old-profile
```

### View Configuration

```bash
portainer-cli config view
```

Displays the entire configuration file contents.

## Configuration Priority

The CLI resolves configuration values in the following order (highest to lowest priority):

1. **Command-line flags**: `--url`, `--api-key`, etc.
2. **Environment variables**: `PORTAINER_URL`, `PORTAINER_API_KEY`, etc.
3. **Profile-specific values**: Values from the specified or current profile
4. **Default values**: Built-in defaults

### Example

```bash
# Uses URL from command-line flag (highest priority)
portainer-cli --url https://override.com environments list

# Uses URL from environment variable
export PORTAINER_URL=https://env.com
portainer-cli environments list

# Uses URL from current profile
portainer-cli config use-profile production
portainer-cli environments list
```

## Environment Variables

The following environment variables are supported:

- `PORTAINER_URL`: Portainer server URL
- `PORTAINER_API_KEY`: API key for authentication
- `PORTAINER_USERNAME`: Username for authentication
- `PORTAINER_PASSWORD`: Password for authentication
- `PORTAINER_TOKEN`: JWT token for authentication
- `XDG_CONFIG_HOME`: Base directory for configuration files (Unix only)

## Security Best Practices

### File Permissions

The CLI automatically sets secure permissions:
- Config directory: `0700` (rwx------)
- Config file: `0600` (rw-------)

### Credential Storage

- API keys and tokens are stored in plain text in the config file
- Ensure the config file has proper permissions (0600)
- Never commit the config file to version control
- Consider using environment variables in CI/CD pipelines

### Masking Secrets

When displaying configuration with `config get`, secrets are masked:
```
API Key: test****key
Token: jwt-****-end
```

To view the actual value, use:
```bash
portainer-cli config get api_key
```

## Multiple Environments

Manage multiple Portainer instances using profiles:

```bash
# Create profiles for each environment
portainer-cli config create-profile prod --url https://prod.example.com --api-key PROD_KEY
portainer-cli config create-profile staging --url https://staging.example.com --api-key STAGING_KEY
portainer-cli config create-profile dev --url http://localhost:9000 --api-key DEV_KEY

# Switch between environments
portainer-cli config use-profile prod
portainer-cli environments list

portainer-cli config use-profile staging
portainer-cli environments list

# Or use --profile flag without switching
portainer-cli --profile dev environments list
```

## Troubleshooting

### Configuration Not Found

```bash
# Initialize configuration
portainer-cli config init

# Create a profile
portainer-cli config create-profile default \
  --url https://portainer.example.com \
  --api-key YOUR_KEY

# Set as current
portainer-cli config use-profile default
```

### Permission Denied

If you get permission errors:

```bash
# Check permissions
ls -la ~/.portainer-cli/

# Fix permissions if needed
chmod 700 ~/.portainer-cli
chmod 600 ~/.portainer-cli/config.yaml
```

### Invalid Configuration

```bash
# View current configuration
portainer-cli config view

# Validate by trying to get profile
portainer-cli config get

# Re-initialize if corrupted
rm ~/.portainer-cli/config.yaml
portainer-cli config init
```

## Examples

### Quick Setup

```bash
# Initialize and create first profile
portainer-cli config init
portainer-cli config create-profile default \
  --url https://portainer.example.com \
  --api-key YOUR_API_KEY
portainer-cli config use-profile default

# Test connection
portainer-cli auth status
```

### Multi-Environment Workflow

```bash
# Set up all environments
portainer-cli config create-profile prod --url https://prod.example.com --api-key PROD_KEY
portainer-cli config create-profile staging --url https://staging.example.com --api-key STAGING_KEY
portainer-cli config create-profile dev --url http://localhost:9000 --api-key DEV_KEY

# Use production by default
portainer-cli config use-profile prod

# Deploy to staging
portainer-cli --profile staging stacks deploy --file stack.yml --name myapp

# Check dev environment
portainer-cli --profile dev environments list
```

### CI/CD Integration

```bash
# Use environment variables in CI/CD
export PORTAINER_URL=https://portainer.example.com
export PORTAINER_API_KEY=$CI_PORTAINER_API_KEY

# No profile needed when using env vars
portainer-cli stacks deploy --file stack.yml --name myapp
```
