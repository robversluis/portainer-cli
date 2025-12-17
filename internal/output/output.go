package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v3"
)

type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatYAML  Format = "yaml"
)

type Formatter interface {
	Format(data interface{}) error
}

type Options struct {
	Format  Format
	Writer  io.Writer
	Quiet   bool
	Verbose bool
	Fields  []string
}

func NewFormatter(opts Options) Formatter {
	if opts.Writer == nil {
		opts.Writer = os.Stdout
	}

	switch opts.Format {
	case FormatJSON:
		return &JSONFormatter{writer: opts.Writer}
	case FormatYAML:
		return &YAMLFormatter{writer: opts.Writer}
	default:
		return &TableFormatter{
			writer:  opts.Writer,
			verbose: opts.Verbose,
		}
	}
}

func Print(data interface{}, format Format) error {
	formatter := NewFormatter(Options{
		Format: format,
		Writer: os.Stdout,
	})
	return formatter.Format(data)
}

func PrintTable(data interface{}) error {
	return Print(data, FormatTable)
}

func PrintJSON(data interface{}) error {
	return Print(data, FormatJSON)
}

func PrintYAML(data interface{}) error {
	return Print(data, FormatYAML)
}

type TableFormatter struct {
	writer  io.Writer
	verbose bool
}

func (f *TableFormatter) Format(data interface{}) error {
	switch v := data.(type) {
	case [][]string:
		return f.formatStringSlice(v)
	case TableData:
		return f.formatTableData(v)
	default:
		return fmt.Errorf("unsupported data type for table format: %T", data)
	}
}

func (f *TableFormatter) formatStringSlice(data [][]string) error {
	if len(data) == 0 {
		return nil
	}

	table := tablewriter.NewWriter(f.writer)

	if len(data) > 0 {
		table.SetHeader(data[0])
		if len(data) > 1 {
			table.AppendBulk(data[1:])
		}
	}

	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)

	table.Render()
	return nil
}

func (f *TableFormatter) formatTableData(data TableData) error {
	if len(data.Rows) == 0 {
		fmt.Fprintln(f.writer, "No data available")
		return nil
	}

	table := tablewriter.NewWriter(f.writer)
	table.SetHeader(data.Headers)
	table.AppendBulk(data.Rows)

	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)

	table.Render()
	return nil
}

type JSONFormatter struct {
	writer io.Writer
}

func (f *JSONFormatter) Format(data interface{}) error {
	encoder := json.NewEncoder(f.writer)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

type YAMLFormatter struct {
	writer io.Writer
}

func (f *YAMLFormatter) Format(data interface{}) error {
	encoder := yaml.NewEncoder(f.writer)
	encoder.SetIndent(2)

	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode YAML: %w", err)
	}

	return nil
}

type TableData struct {
	Headers []string
	Rows    [][]string
}

func NewTableData(headers []string) *TableData {
	return &TableData{
		Headers: headers,
		Rows:    make([][]string, 0),
	}
}

func (t *TableData) AddRow(row []string) {
	t.Rows = append(t.Rows, row)
}

func (t *TableData) AddRows(rows [][]string) {
	t.Rows = append(t.Rows, rows...)
}

func FormatBool(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

func FormatStatus(status int) string {
	switch status {
	case 1:
		return "Up"
	case 2:
		return "Down"
	default:
		return "Unknown"
	}
}

func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

func FormatList(items []string, separator string) string {
	if len(items) == 0 {
		return "-"
	}
	return strings.Join(items, separator)
}

func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func FormatDuration(seconds int64) string {
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	if seconds < 3600 {
		return fmt.Sprintf("%dm", seconds/60)
	}
	if seconds < 86400 {
		return fmt.Sprintf("%dh", seconds/3600)
	}
	return fmt.Sprintf("%dd", seconds/86400)
}

func ParseFormat(format string) Format {
	switch strings.ToLower(format) {
	case "json":
		return FormatJSON
	case "yaml", "yml":
		return FormatYAML
	case "table":
		return FormatTable
	default:
		return FormatTable
	}
}
