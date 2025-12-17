# Output Formatting Guide

This document describes the output formatting system for the Portainer CLI.

## Supported Formats

The CLI supports three output formats:

1. **Table** (default) - Human-readable tabular format
2. **JSON** - Machine-readable JSON format
3. **YAML** - Human and machine-readable YAML format

## Usage

### Specifying Output Format

Use the global `--output` or `-o` flag:

```bash
# Table format (default)
portainer-cli environments list

# JSON format
portainer-cli environments list --output json
portainer-cli environments list -o json

# YAML format
portainer-cli environments list --output yaml
portainer-cli environments list -o yaml
```

### Table Format

The default format displays data in a clean, aligned table:

```bash
portainer-cli environments list
```

Output:
```
ID   Name        Type        URL                    Status
1    local       docker      unix:///var/run/dock   Up
2    production  kubernetes  https://k8s.prod.com   Up
3    staging     docker      tcp://staging:2375     Down
```

**Features:**
- Auto-aligned columns
- Clean, minimal borders
- Easy to read in terminal
- Truncates long values when needed

### JSON Format

Perfect for scripting and programmatic access:

```bash
portainer-cli environments list -o json
```

Output:
```json
[
  {
    "Id": 1,
    "Name": "local",
    "Type": "docker",
    "URL": "unix:///var/run/docker.sock",
    "Status": 1
  },
  {
    "Id": 2,
    "Name": "production",
    "Type": "kubernetes",
    "URL": "https://k8s.prod.com",
    "Status": 1
  }
]
```

**Features:**
- Pretty-printed with 2-space indentation
- Valid JSON that can be piped to `jq`
- Preserves all data types
- No HTML escaping

### YAML Format

Human-readable structured format:

```bash
portainer-cli environments list -o yaml
```

Output:
```yaml
- Id: 1
  Name: local
  Type: docker
  URL: unix:///var/run/docker.sock
  Status: 1
- Id: 2
  Name: production
  Type: kubernetes
  URL: https://k8s.prod.com
  Status: 1
```

**Features:**
- Clean, indented structure
- Easy to read and edit
- Compatible with YAML parsers

## Quiet and Verbose Modes

### Quiet Mode

Minimal output for scripting:

```bash
portainer-cli -q environments list
```

Suppresses informational messages and only shows essential data.

### Verbose Mode

Detailed output for debugging:

```bash
portainer-cli -v environments list
```

Shows additional information like:
- HTTP requests and responses
- Detailed error messages
- Internal processing steps

## Scripting Examples

### Parse JSON with jq

```bash
# Get all environment names
portainer-cli environments list -o json | jq -r '.[].Name'

# Filter by status
portainer-cli environments list -o json | jq '.[] | select(.Status == 1)'

# Count environments
portainer-cli environments list -o json | jq 'length'
```

### Parse YAML with yq

```bash
# Get environment IDs
portainer-cli environments list -o yaml | yq '.[].Id'

# Filter by type
portainer-cli environments list -o yaml | yq '.[] | select(.Type == "docker")'
```

### Process in Shell Scripts

```bash
#!/bin/bash

# Get environments as JSON
environments=$(portainer-cli environments list -o json)

# Parse with jq
for env in $(echo "$environments" | jq -r '.[].Id'); do
    echo "Processing environment $env"
    portainer-cli containers list --endpoint "$env"
done
```

### Save to File

```bash
# Save as JSON
portainer-cli environments list -o json > environments.json

# Save as YAML
portainer-cli stacks list -o yaml > stacks.yaml

# Append to file
portainer-cli containers list -o json >> containers.json
```

## Helper Functions

The output package provides utility functions for formatting common data types:

### FormatBool

Convert boolean to Yes/No:

```go
FormatBool(true)   // "Yes"
FormatBool(false)  // "No"
```

### FormatStatus

Convert status code to readable string:

```go
FormatStatus(1)  // "Up"
FormatStatus(2)  // "Down"
FormatStatus(0)  // "Unknown"
```

### FormatSize

Convert bytes to human-readable size:

```go
FormatSize(1024)       // "1.0 KB"
FormatSize(1048576)    // "1.0 MB"
FormatSize(1073741824) // "1.0 GB"
```

### FormatDuration

Convert seconds to readable duration:

```go
FormatDuration(45)     // "45s"
FormatDuration(120)    // "2m"
FormatDuration(7200)   // "2h"
FormatDuration(172800) // "2d"
```

### TruncateString

Truncate long strings with ellipsis:

```go
TruncateString("very long string", 10) // "very lo..."
```

### FormatList

Join list items with separator:

```go
FormatList([]string{"a", "b", "c"}, ", ") // "a, b, c"
FormatList([]string{}, ", ")              // "-"
```

## Custom Table Output

For custom commands, use the `TableData` structure:

```go
import "github.com/rob/portainer-cli/internal/output"

// Create table
table := output.NewTableData([]string{"ID", "Name", "Status"})

// Add rows
table.AddRow([]string{"1", "Container1", "Running"})
table.AddRow([]string{"2", "Container2", "Stopped"})

// Format and print
formatter := output.NewFormatter(output.Options{
    Format: output.ParseFormat(outputFlag),
})
formatter.Format(table)
```

## Integration with Commands

Commands should support all output formats:

```go
func runListCommand(cmd *cobra.Command, args []string) error {
    // Get data from API
    items, err := client.GetItems()
    if err != nil {
        return err
    }

    // Get output format from flag
    format := output.ParseFormat(cmd.Flag("output").Value.String())

    switch format {
    case output.FormatJSON, output.FormatYAML:
        // For JSON/YAML, output raw data
        formatter := output.NewFormatter(output.Options{Format: format})
        return formatter.Format(items)
    
    default:
        // For table, format as TableData
        table := output.NewTableData([]string{"ID", "Name", "Status"})
        for _, item := range items {
            table.AddRow([]string{
                fmt.Sprintf("%d", item.ID),
                item.Name,
                output.FormatStatus(item.Status),
            })
        }
        return output.PrintTable(table)
    }
}
```

## Best Practices

### 1. Always Support All Formats

Every command should support table, JSON, and YAML output:

```bash
✓ portainer-cli environments list -o json
✓ portainer-cli environments list -o yaml
✓ portainer-cli environments list -o table
```

### 2. Use Consistent Field Names

Keep field names consistent across formats:

```json
{
  "Id": 1,        // Not "id" or "ID"
  "Name": "test", // Not "name"
  "Status": 1     // Not "status"
}
```

### 3. Preserve Data Types

In JSON/YAML, preserve original data types:

```json
{
  "Id": 1,              // number, not "1"
  "Active": true,       // boolean, not "true"
  "Tags": ["a", "b"]    // array, not "a, b"
}
```

### 4. Format for Humans in Tables

In table format, make data human-readable:

```
Status: Up          (not 1)
Size: 1.5 GB        (not 1610612736)
Uptime: 2d          (not 172800)
Active: Yes         (not true)
```

### 5. Handle Empty Data

Gracefully handle empty results:

```bash
# Table format
portainer-cli stacks list
No data available

# JSON format
portainer-cli stacks list -o json
[]

# YAML format
portainer-cli stacks list -o yaml
[]
```

## Troubleshooting

### Invalid JSON Output

If JSON output is malformed:

```bash
# Validate JSON
portainer-cli environments list -o json | jq .

# Pretty print
portainer-cli environments list -o json | jq '.'
```

### Table Alignment Issues

If table columns are misaligned, the data may contain tabs or newlines. Use `TruncateString` to limit field length.

### YAML Parsing Errors

Ensure YAML output is valid:

```bash
# Validate YAML
portainer-cli environments list -o yaml | yq .
```

## Examples

### Compare Environments

```bash
# Get production environment
portainer-cli --profile prod environments list -o json > prod.json

# Get staging environment
portainer-cli --profile staging environments list -o json > staging.json

# Compare
diff <(jq -S . prod.json) <(jq -S . staging.json)
```

### Export Configuration

```bash
# Export all stacks
portainer-cli stacks list -o yaml > stacks-backup.yaml

# Export specific stack
portainer-cli stacks get 5 -o yaml > mystack.yaml
```

### Monitor Status

```bash
# Watch environment status
watch -n 5 'portainer-cli environments list'

# Log status changes
while true; do
    portainer-cli environments list -o json | \
        jq -r '.[] | "\(.Name): \(.Status)"' | \
        ts >> status.log
    sleep 60
done
```

### Generate Reports

```bash
# Create HTML report
portainer-cli environments list -o json | \
    jq -r '.[] | "<tr><td>\(.Name)</td><td>\(.Type)</td></tr>"' | \
    cat header.html - footer.html > report.html
```

## Performance Considerations

- **Table format**: Fastest for terminal display
- **JSON format**: Efficient for parsing and processing
- **YAML format**: Slightly slower due to formatting

For large datasets, consider using JSON format with streaming parsers.
