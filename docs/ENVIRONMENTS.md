# Environment Management Guide

This document describes how to manage Portainer environments (endpoints) using the CLI.

## Overview

Environments in Portainer represent the Docker, Kubernetes, or other container orchestration platforms you want to manage. Each environment has a unique ID, name, type, and connection details.

## Commands

### List Environments

Display all environments with their status:

```bash
portainer-cli environments list
# or
portainer-cli env ls
```

Output:
```
ID   Name        Type                URL                                Status
1    local       Docker (Local)      unix:///var/run/docker.sock       Up
2    production  Kubernetes (Agent)  https://k8s.prod.com              Up
3    staging     Docker (Agent)      tcp://staging.example.com:2375    Down
```

**JSON output:**
```bash
portainer-cli environments list -o json
```

**YAML output:**
```bash
portainer-cli environments list -o yaml
```

### Get Environment Details

Get detailed information about a specific environment by ID or name:

```bash
# By ID
portainer-cli environments get 1

# By name
portainer-cli environments get local

# Alias
portainer-cli env get production
```

Output:
```
ID:          1
Name:        local
Type:        Docker (Local)
URL:         unix:///var/run/docker.sock
Status:      Up
Group ID:    1

Agent:
  Version:   2.19.0

Snapshot:
  Containers:  5 running, 2 stopped
  Health:      4 healthy, 1 unhealthy
  Images:      15
  Volumes:     8
  Stacks:      3
  CPU:         8
  Memory:      16.0 GB
```

### Inspect Environment

Alias for `get` command:

```bash
portainer-cli environments inspect 1
portainer-cli env inspect production
```

## Environment Types

The CLI supports all Portainer environment types:

| Type | Description |
|------|-------------|
| Docker (Local) | Local Docker socket |
| Docker (Agent) | Remote Docker via Portainer Agent |
| Docker (Edge) | Edge Docker environment |
| Kubernetes (Local) | Local Kubernetes cluster |
| Kubernetes (Agent) | Remote Kubernetes via Portainer Agent |
| Kubernetes (Edge) | Edge Kubernetes environment |
| Azure | Azure Container Instances |

## Environment Status

Environments can have the following statuses:

- **Up**: Environment is reachable and healthy
- **Down**: Environment is unreachable or unhealthy
- **Unknown**: Status cannot be determined

## Output Formats

### Table Format (Default)

Clean, human-readable table:

```bash
portainer-cli environments list
```

### JSON Format

Machine-readable JSON for scripting:

```bash
portainer-cli environments list -o json | jq '.[] | select(.Status == 1)'
```

### YAML Format

Structured YAML output:

```bash
portainer-cli environments list -o yaml
```

## Examples

### List All Environments

```bash
portainer-cli environments list
```

### Get Specific Environment

```bash
# By ID
portainer-cli environments get 1

# By name
portainer-cli environments get production
```

### Filter Environments with jq

```bash
# Get only running environments
portainer-cli environments list -o json | jq '.[] | select(.Status == 1)'

# Get environment names
portainer-cli environments list -o json | jq -r '.[].Name'

# Count environments by type
portainer-cli environments list -o json | jq 'group_by(.Type) | map({type: .[0].Type, count: length})'
```

### Check Environment Status

```bash
# Check if environment is up
if portainer-cli environments get production -o json | jq -e '.Status == 1' > /dev/null; then
    echo "Production is up"
else
    echo "Production is down"
fi
```

### Export Environment Configuration

```bash
# Export single environment
portainer-cli environments get 1 -o yaml > environment-1.yaml

# Export all environments
portainer-cli environments list -o json > environments-backup.json
```

## Scripting

### Bash Script Example

```bash
#!/bin/bash

# Get all environment IDs
env_ids=$(portainer-cli environments list -o json | jq -r '.[].Id')

# Loop through environments
for id in $env_ids; do
    echo "Processing environment $id"
    
    # Get environment details
    env=$(portainer-cli environments get $id -o json)
    name=$(echo "$env" | jq -r '.Name')
    status=$(echo "$env" | jq -r '.Status')
    
    if [ "$status" == "1" ]; then
        echo "  $name is UP"
        # Do something with running environment
    else
        echo "  $name is DOWN"
        # Handle down environment
    fi
done
```

### Python Script Example

```python
#!/usr/bin/env python3
import json
import subprocess

def get_environments():
    result = subprocess.run(
        ['portainer-cli', 'environments', 'list', '-o', 'json'],
        capture_output=True,
        text=True
    )
    return json.loads(result.stdout)

def get_environment(env_id):
    result = subprocess.run(
        ['portainer-cli', 'environments', 'get', str(env_id), '-o', 'json'],
        capture_output=True,
        text=True
    )
    return json.loads(result.stdout)

# Get all environments
environments = get_environments()

for env in environments:
    print(f"Environment: {env['Name']}")
    print(f"  Type: {env['Type']}")
    print(f"  Status: {'Up' if env['Status'] == 1 else 'Down'}")
    
    # Get detailed info
    details = get_environment(env['Id'])
    if details.get('Snapshots'):
        snapshot = details['Snapshots'][0]
        print(f"  Containers: {snapshot.get('RunningContainerCount', 0)}")
```

## Monitoring

### Watch Environment Status

```bash
# Watch environments every 5 seconds
watch -n 5 'portainer-cli environments list'
```

### Log Status Changes

```bash
#!/bin/bash

while true; do
    timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    portainer-cli environments list -o json | \
        jq -r '.[] | "\(.Name): \(.Status)"' | \
        while read line; do
            echo "[$timestamp] $line"
        done
    
    sleep 60
done >> environment-status.log
```

### Alert on Down Environments

```bash
#!/bin/bash

down_envs=$(portainer-cli environments list -o json | \
    jq -r '.[] | select(.Status == 2) | .Name')

if [ -n "$down_envs" ]; then
    echo "ALERT: The following environments are down:"
    echo "$down_envs"
    # Send notification (email, Slack, etc.)
fi
```

## Environment Snapshots

Snapshots contain resource usage and statistics:

- **Containers**: Running, stopped, healthy, unhealthy counts
- **Images**: Total image count
- **Volumes**: Total volume count
- **Services**: Swarm service count (Docker only)
- **Stacks**: Stack count
- **CPU**: Total CPU cores
- **Memory**: Total memory in bytes

Access snapshot data:

```bash
# Get latest snapshot
portainer-cli environments get 1 -o json | jq '.Snapshots[0]'

# Get container counts
portainer-cli environments get 1 -o json | \
    jq '{running: .Snapshots[0].RunningContainerCount, stopped: .Snapshots[0].StoppedContainerCount}'
```

## Troubleshooting

### Environment Not Found

```
Error: environment 'myenv' not found
```

**Solution**: Check environment name or use ID instead:
```bash
portainer-cli environments list
portainer-cli environments get 1
```

### Connection Refused

```
Error: failed to list environments: request failed: connection refused
```

**Solution**: Verify Portainer URL and authentication:
```bash
portainer-cli config get url
portainer-cli auth status
```

### Unauthorized

```
Error: API error (HTTP 401): Unauthorized
```

**Solution**: Login or check API key:
```bash
portainer-cli auth login
# or
portainer-cli config set api_key YOUR_KEY
```

### Environment Down

If an environment shows as "Down":

1. Check the environment URL is correct
2. Verify network connectivity
3. Check Portainer Agent is running (for Agent environments)
4. Review Portainer logs for connection errors

## Best Practices

### 1. Use Environment Names

Use descriptive names for easier identification:

```bash
portainer-cli environments get production
```

### 2. Monitor Regularly

Set up monitoring for environment status:

```bash
# Cron job to check every 5 minutes
*/5 * * * * /path/to/check-environments.sh
```

### 3. Export Configurations

Regularly backup environment configurations:

```bash
portainer-cli environments list -o json > backups/environments-$(date +%Y%m%d).json
```

### 4. Use JSON for Automation

Use JSON output for scripting and automation:

```bash
portainer-cli environments list -o json | jq '.[] | select(.Status == 1) | .Id'
```

### 5. Filter by Type

Filter environments by type for specific operations:

```bash
# Get only Kubernetes environments
portainer-cli environments list -o json | jq '.[] | select(.Type == 4 or .Type == 6 or .Type == 7)'
```

## Integration Examples

### Ansible Playbook

```yaml
---
- name: Check Portainer Environments
  hosts: localhost
  tasks:
    - name: Get environments
      command: portainer-cli environments list -o json
      register: environments
    
    - name: Parse environments
      set_fact:
        env_list: "{{ environments.stdout | from_json }}"
    
    - name: Display environment status
      debug:
        msg: "{{ item.Name }}: {{ 'Up' if item.Status == 1 else 'Down' }}"
      loop: "{{ env_list }}"
```

### Terraform Data Source

```bash
# Get environment ID for Terraform
export TF_VAR_portainer_env_id=$(portainer-cli environments get production -o json | jq -r '.Id')
```

### CI/CD Pipeline

```yaml
# GitHub Actions example
- name: Check Environment Status
  run: |
    STATUS=$(portainer-cli environments get production -o json | jq -r '.Status')
    if [ "$STATUS" != "1" ]; then
      echo "Production environment is down!"
      exit 1
    fi
```

## API Reference

The CLI uses these Portainer API endpoints:

- `GET /api/endpoints` - List all environments
- `GET /api/endpoints/{id}` - Get environment details
- `DELETE /api/endpoints/{id}` - Delete environment

For more information, see the [Portainer API documentation](https://docs.portainer.io/api/docs).
