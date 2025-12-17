package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestNewFormatter(t *testing.T) {
	tests := []struct {
		name           string
		format         Format
		expectedType   string
	}{
		{
			name:         "table format",
			format:       FormatTable,
			expectedType: "*output.TableFormatter",
		},
		{
			name:         "json format",
			format:       FormatJSON,
			expectedType: "*output.JSONFormatter",
		},
		{
			name:         "yaml format",
			format:       FormatYAML,
			expectedType: "*output.YAMLFormatter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewFormatter(Options{Format: tt.format})
			if formatter == nil {
				t.Fatal("formatter should not be nil")
			}
		})
	}
}

func TestJSONFormatter(t *testing.T) {
	tests := []struct {
		name     string
		data     interface{}
		wantErr  bool
		validate func(string) bool
	}{
		{
			name: "simple map",
			data: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			wantErr: false,
			validate: func(output string) bool {
				var result map[string]string
				return json.Unmarshal([]byte(output), &result) == nil
			},
		},
		{
			name: "slice of maps",
			data: []map[string]interface{}{
				{"id": 1, "name": "test1"},
				{"id": 2, "name": "test2"},
			},
			wantErr: false,
			validate: func(output string) bool {
				var result []map[string]interface{}
				return json.Unmarshal([]byte(output), &result) == nil
			},
		},
		{
			name: "struct",
			data: struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			}{
				ID:   1,
				Name: "test",
			},
			wantErr: false,
			validate: func(output string) bool {
				return strings.Contains(output, `"id"`) && strings.Contains(output, `"name"`)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			formatter := &JSONFormatter{writer: buf}

			err := formatter.Format(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Format() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				output := buf.String()
				if !tt.validate(output) {
					t.Errorf("JSON validation failed for output: %s", output)
				}
			}
		})
	}
}

func TestYAMLFormatter(t *testing.T) {
	tests := []struct {
		name     string
		data     interface{}
		wantErr  bool
		validate func(string) bool
	}{
		{
			name: "simple map",
			data: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			wantErr: false,
			validate: func(output string) bool {
				var result map[string]string
				return yaml.Unmarshal([]byte(output), &result) == nil
			},
		},
		{
			name: "nested structure",
			data: map[string]interface{}{
				"parent": map[string]string{
					"child": "value",
				},
			},
			wantErr: false,
			validate: func(output string) bool {
				return strings.Contains(output, "parent:") && strings.Contains(output, "child:")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			formatter := &YAMLFormatter{writer: buf}

			err := formatter.Format(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Format() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				output := buf.String()
				if !tt.validate(output) {
					t.Errorf("YAML validation failed for output: %s", output)
				}
			}
		})
	}
}

func TestTableFormatter(t *testing.T) {
	tests := []struct {
		name     string
		data     interface{}
		wantErr  bool
		validate func(string) bool
	}{
		{
			name: "string slice",
			data: [][]string{
				{"Header1", "Header2", "Header3"},
				{"Value1", "Value2", "Value3"},
				{"Value4", "Value5", "Value6"},
			},
			wantErr: false,
			validate: func(output string) bool {
				upper := strings.ToUpper(output)
				return strings.Contains(upper, "HEADER1") &&
					strings.Contains(output, "Value1") &&
					strings.Contains(output, "Value6")
			},
		},
		{
			name: "table data",
			data: TableData{
				Headers: []string{"ID", "Name", "Status"},
				Rows: [][]string{
					{"1", "Test1", "Active"},
					{"2", "Test2", "Inactive"},
				},
			},
			wantErr: false,
			validate: func(output string) bool {
				return strings.Contains(output, "ID") &&
					strings.Contains(output, "Test1") &&
					strings.Contains(output, "Active")
			},
		},
		{
			name: "empty table data",
			data: TableData{
				Headers: []string{"ID", "Name"},
				Rows:    [][]string{},
			},
			wantErr: false,
			validate: func(output string) bool {
				return strings.Contains(output, "No data available")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			formatter := &TableFormatter{writer: buf}

			err := formatter.Format(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Format() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				output := buf.String()
				if !tt.validate(output) {
					t.Errorf("Table validation failed for output: %s", output)
				}
			}
		})
	}
}

func TestTableData(t *testing.T) {
	t.Run("add single row", func(t *testing.T) {
		td := NewTableData([]string{"Col1", "Col2"})
		td.AddRow([]string{"Val1", "Val2"})

		if len(td.Rows) != 1 {
			t.Errorf("expected 1 row, got %d", len(td.Rows))
		}

		if td.Rows[0][0] != "Val1" {
			t.Errorf("expected 'Val1', got '%s'", td.Rows[0][0])
		}
	})

	t.Run("add multiple rows", func(t *testing.T) {
		td := NewTableData([]string{"Col1", "Col2"})
		rows := [][]string{
			{"Val1", "Val2"},
			{"Val3", "Val4"},
			{"Val5", "Val6"},
		}
		td.AddRows(rows)

		if len(td.Rows) != 3 {
			t.Errorf("expected 3 rows, got %d", len(td.Rows))
		}
	})
}

func TestFormatBool(t *testing.T) {
	tests := []struct {
		input    bool
		expected string
	}{
		{true, "Yes"},
		{false, "No"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := FormatBool(tt.input)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestFormatStatus(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{1, "Up"},
		{2, "Down"},
		{0, "Unknown"},
		{99, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := FormatStatus(tt.input)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "no truncation needed",
			input:    "short",
			maxLen:   10,
			expected: "short",
		},
		{
			name:     "truncate with ellipsis",
			input:    "this is a very long string",
			maxLen:   10,
			expected: "this is...",
		},
		{
			name:     "exact length",
			input:    "exactly10c",
			maxLen:   10,
			expected: "exactly10c",
		},
		{
			name:     "very short maxLen",
			input:    "test",
			maxLen:   2,
			expected: "te",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateString(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestFormatList(t *testing.T) {
	tests := []struct {
		name      string
		items     []string
		separator string
		expected  string
	}{
		{
			name:      "empty list",
			items:     []string{},
			separator: ", ",
			expected:  "-",
		},
		{
			name:      "single item",
			items:     []string{"item1"},
			separator: ", ",
			expected:  "item1",
		},
		{
			name:      "multiple items with comma",
			items:     []string{"item1", "item2", "item3"},
			separator: ", ",
			expected:  "item1, item2, item3",
		},
		{
			name:      "multiple items with pipe",
			items:     []string{"a", "b", "c"},
			separator: " | ",
			expected:  "a | b | c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatList(tt.items, tt.separator)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{
			name:     "bytes",
			bytes:    512,
			expected: "512 B",
		},
		{
			name:     "kilobytes",
			bytes:    1536,
			expected: "1.5 KB",
		},
		{
			name:     "megabytes",
			bytes:    1048576,
			expected: "1.0 MB",
		},
		{
			name:     "gigabytes",
			bytes:    1073741824,
			expected: "1.0 GB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatSize(tt.bytes)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		seconds  int64
		expected string
	}{
		{
			name:     "seconds",
			seconds:  45,
			expected: "45s",
		},
		{
			name:     "minutes",
			seconds:  120,
			expected: "2m",
		},
		{
			name:     "hours",
			seconds:  7200,
			expected: "2h",
		},
		{
			name:     "days",
			seconds:  172800,
			expected: "2d",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDuration(tt.seconds)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestParseFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected Format
	}{
		{"json", FormatJSON},
		{"JSON", FormatJSON},
		{"yaml", FormatYAML},
		{"YAML", FormatYAML},
		{"yml", FormatYAML},
		{"table", FormatTable},
		{"TABLE", FormatTable},
		{"invalid", FormatTable},
		{"", FormatTable},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseFormat(tt.input)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestPrintFunctions(t *testing.T) {
	data := map[string]string{"key": "value"}

	t.Run("PrintJSON", func(t *testing.T) {
		err := PrintJSON(data)
		if err != nil {
			t.Errorf("PrintJSON failed: %v", err)
		}
	})

	t.Run("PrintYAML", func(t *testing.T) {
		err := PrintYAML(data)
		if err != nil {
			t.Errorf("PrintYAML failed: %v", err)
		}
	})

	t.Run("PrintTable", func(t *testing.T) {
		tableData := TableData{
			Headers: []string{"Key", "Value"},
			Rows:    [][]string{{"key", "value"}},
		}
		err := PrintTable(tableData)
		if err != nil {
			t.Errorf("PrintTable failed: %v", err)
		}
	})
}
