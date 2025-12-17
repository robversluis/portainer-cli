package cmd

import (
	"testing"
)

func TestRootCommand(t *testing.T) {
	if rootCmd == nil {
		t.Fatal("rootCmd should not be nil")
	}

	if rootCmd.Use != "portainer-cli" {
		t.Errorf("Expected Use to be 'portainer-cli', got '%s'", rootCmd.Use)
	}

	if rootCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestGlobalFlags(t *testing.T) {
	flags := rootCmd.PersistentFlags()

	if flags.Lookup("config") == nil {
		t.Error("config flag should be defined")
	}

	if flags.Lookup("profile") == nil {
		t.Error("profile flag should be defined")
	}

	if flags.Lookup("url") == nil {
		t.Error("url flag should be defined")
	}

	if flags.Lookup("api-key") == nil {
		t.Error("api-key flag should be defined")
	}

	if flags.Lookup("output") == nil {
		t.Error("output flag should be defined")
	}

	if flags.Lookup("verbose") == nil {
		t.Error("verbose flag should be defined")
	}

	if flags.Lookup("quiet") == nil {
		t.Error("quiet flag should be defined")
	}
}

func TestGetters(t *testing.T) {
	verbose = true
	if !GetVerbose() {
		t.Error("GetVerbose should return true")
	}

	quiet = true
	if !GetQuiet() {
		t.Error("GetQuiet should return true")
	}

	output = "json"
	if GetOutput() != "json" {
		t.Errorf("Expected output to be 'json', got '%s'", GetOutput())
	}

	url = "https://test.example.com"
	if GetURL() != "https://test.example.com" {
		t.Errorf("Expected URL to be 'https://test.example.com', got '%s'", GetURL())
	}

	apiKey = "test-key"
	if GetAPIKey() != "test-key" {
		t.Errorf("Expected API key to be 'test-key', got '%s'", GetAPIKey())
	}
}
