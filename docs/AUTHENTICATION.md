# Authentication Guide

This document describes authentication methods and usage for the Portainer CLI.

## Authentication Methods

The CLI supports three authentication methods:

1. **API Key** - Long-lived token for programmatic access
2. **JWT Token** - Session token obtained via username/password login
3. **Username/Password** - Interactive login (stores JWT token)

## Quick Start

### Login with Username/Password

```bash
# Interactive login (prompts for password)
portainer-cli auth login --username admin

# Non-interactive login
portainer-cli auth login --username admin --password mypassword

# Login to specific URL
portainer-cli --url https://portainer.example.com auth login --username admin
```

The JWT token is automatically stored in your current profile for future use.

### Using API Keys

```bash
# Set API key in profile
portainer-cli config set api_key YOUR_API_KEY

# Or use via flag
portainer-cli --api-key YOUR_API_KEY environments list

# Or via environment variable
export PORTAINER_API_KEY=YOUR_API_KEY
portainer-cli environments list
```

### Check Authentication Status

```bash
portainer-cli auth status
```

Output:
```
Portainer URL: https://portainer.example.com
Portainer Version: 2.19.0
Authentication Method: JWT Token
Authentication Status: Valid
Logged in as: admin (ID: 1, Role: 1)
```

### Logout

```bash
portainer-cli auth logout
```

This clears the stored JWT token from your profile.

## Authentication Flow

### Username/Password Flow

1. User provides username and password
2. CLI sends credentials to `/api/auth` endpoint
3. Portainer returns JWT token
4. CLI stores token in current profile
5. Token is used for subsequent requests

### API Key Flow

1. User configures API key in profile or via flag
2. CLI sends `X-API-KEY` header with each request
3. No login required

### Token Validation

The CLI automatically validates tokens by making a test API call. Invalid tokens will result in authentication errors.

## Security Best Practices

### Secure Password Entry

The CLI uses secure password input that doesn't echo to the terminal:

```bash
portainer-cli auth login --username admin
Password: [hidden input]
```

### Token Storage

- JWT tokens are stored in `~/.portainer-cli/config.yaml`
- Config file has `0600` permissions (read/write for owner only)
- Never commit config files to version control

### API Key Management

- Generate API keys in Portainer UI: User Settings → API Keys
- Use different API keys for different purposes
- Rotate API keys regularly
- Revoke unused API keys

### Environment Variables

For CI/CD pipelines, use environment variables instead of config files:

```bash
export PORTAINER_URL=https://portainer.example.com
export PORTAINER_API_KEY=$SECRET_API_KEY
portainer-cli stacks deploy --file stack.yml
```

## Certificate Validation

### HTTPS with Valid Certificates

By default, the CLI validates TLS certificates:

```bash
portainer-cli --url https://portainer.example.com auth status
```

### Skip Certificate Validation (Not Recommended)

For development or self-signed certificates:

```bash
# Set in profile
portainer-cli config set insecure true

# Or via profile creation
portainer-cli config create-profile dev \
  --url https://localhost:9443 \
  --api-key KEY \
  --insecure
```

**Warning**: Only use `--insecure` for development environments. Never use in production.

### Custom CA Certificates

For custom certificate authorities, the CLI respects system certificate stores. Add your CA certificate to your system's trusted certificates.

## Error Handling

### Common Authentication Errors

#### Invalid Credentials
```
Error: login failed: invalid credentials
```

**Solution**: Verify username and password are correct.

#### Unauthorized (401)
```
Error: API error (HTTP 401): Unauthorized
```

**Solution**: Token expired or invalid. Run `portainer-cli auth login` again.

#### Forbidden (403)
```
Error: API error (HTTP 403): Forbidden
```

**Solution**: User lacks permissions for the operation. Check user role in Portainer.

#### Connection Refused
```
Error: request failed: connection refused
```

**Solution**: Verify Portainer URL is correct and server is running.

#### Certificate Validation Failed
```
Error: x509: certificate signed by unknown authority
```

**Solution**: Add CA certificate to system trust store or use `--insecure` for development.

## Multiple Environments

### Separate Profiles for Each Environment

```bash
# Production with API key
portainer-cli config create-profile prod \
  --url https://portainer.prod.com \
  --api-key $PROD_API_KEY

# Staging with username
portainer-cli config create-profile staging \
  --url https://portainer.staging.com \
  --username admin

# Login to staging
portainer-cli --profile staging auth login

# Use production
portainer-cli --profile prod environments list
```

## Advanced Usage

### Verbose Mode

See detailed authentication information:

```bash
portainer-cli -v auth login --username admin
```

Output includes:
- HTTP requests and responses
- Token information
- API endpoints called

### Programmatic Authentication

For scripts and automation:

```bash
#!/bin/bash
set -e

# Login and capture output
if portainer-cli auth login --username admin --password "$PASSWORD" > /dev/null 2>&1; then
    echo "Login successful"
    portainer-cli environments list
else
    echo "Login failed"
    exit 1
fi
```

### Token Refresh

JWT tokens expire after a period (configured in Portainer). When a token expires:

1. CLI returns 401 Unauthorized error
2. User must login again: `portainer-cli auth login`
3. New token is stored in profile

**Note**: API keys don't expire but can be revoked in Portainer UI.

## Troubleshooting

### Check Current Authentication

```bash
# View profile configuration
portainer-cli config get

# Test authentication
portainer-cli auth status

# Verbose status check
portainer-cli -v auth status
```

### Clear Stored Credentials

```bash
# Logout (clears token)
portainer-cli auth logout

# Or manually edit config
portainer-cli config view
portainer-cli config set token ""
```

### Reset Authentication

```bash
# Remove profile and recreate
portainer-cli config delete-profile myprofile
portainer-cli config create-profile myprofile \
  --url https://portainer.example.com \
  --username admin
portainer-cli --profile myprofile auth login
```

## API Reference

### Authentication Endpoints

The CLI uses these Portainer API endpoints:

- `POST /api/auth` - Login with username/password
- `POST /api/auth/logout` - Logout (clear session)
- `GET /api/status` - Get Portainer version and status
- `GET /api/users` - Validate token and get user info

### Request Headers

**With API Key:**
```
X-API-KEY: your-api-key-here
```

**With JWT Token:**
```
Authorization: Bearer your-jwt-token-here
```

## Examples

### Complete Login Flow

```bash
# 1. Create profile
portainer-cli config create-profile production \
  --url https://portainer.example.com \
  --username admin

# 2. Set as current profile
portainer-cli config use-profile production

# 3. Login
portainer-cli auth login
Username: admin
Password: [hidden]
Successfully logged in as admin

# 4. Verify authentication
portainer-cli auth status
Portainer URL: https://portainer.example.com
Portainer Version: 2.19.0
Authentication Method: JWT Token
Authentication Status: Valid
Logged in as: admin (ID: 1, Role: 1)

# 5. Use CLI
portainer-cli environments list
```

### CI/CD Pipeline

```yaml
# .github/workflows/deploy.yml
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Deploy to Portainer
        env:
          PORTAINER_URL: ${{ secrets.PORTAINER_URL }}
          PORTAINER_API_KEY: ${{ secrets.PORTAINER_API_KEY }}
        run: |
          portainer-cli stacks deploy \
            --file docker-compose.yml \
            --name myapp \
            --endpoint 1
```

### Switching Between Environments

```bash
# Deploy to staging
portainer-cli --profile staging stacks deploy --file stack.yml --name app

# Test on staging
portainer-cli --profile staging stacks list

# Deploy to production
portainer-cli --profile prod stacks deploy --file stack.yml --name app
```
