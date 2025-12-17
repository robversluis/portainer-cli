package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetConfigDir(t *testing.T) {
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", originalXDG)

	t.Run("with XDG_CONFIG_HOME set", func(t *testing.T) {
		os.Setenv("XDG_CONFIG_HOME", "/tmp/xdg-config")
		dir, err := GetConfigDir()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "/tmp/xdg-config/portainer-cli"
		if dir != expected {
			t.Errorf("expected %s, got %s", expected, dir)
		}
	})

	t.Run("without XDG_CONFIG_HOME", func(t *testing.T) {
		os.Unsetenv("XDG_CONFIG_HOME")
		dir, err := GetConfigDir()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if dir == "" {
			t.Error("config dir should not be empty")
		}
		if !filepath.IsAbs(dir) {
			t.Error("config dir should be absolute path")
		}
	})
}

func TestConfig_SetProfile(t *testing.T) {
	cfg := &Config{
		Profiles: make(map[string]*Profile),
	}

	profile := &Profile{
		URL:    "https://test.example.com",
		APIKey: "test-key",
	}

	cfg.SetProfile("test", profile)

	if len(cfg.Profiles) != 1 {
		t.Errorf("expected 1 profile, got %d", len(cfg.Profiles))
	}

	retrieved, err := cfg.GetProfile("test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if retrieved.URL != profile.URL {
		t.Errorf("expected URL %s, got %s", profile.URL, retrieved.URL)
	}

	if retrieved.Name != "test" {
		t.Errorf("expected name 'test', got '%s'", retrieved.Name)
	}
}

func TestConfig_GetProfile(t *testing.T) {
	cfg := &Config{
		CurrentProfile: "default",
		Profiles: map[string]*Profile{
			"default": {
				URL:    "https://default.example.com",
				APIKey: "default-key",
			},
			"prod": {
				URL:    "https://prod.example.com",
				APIKey: "prod-key",
			},
		},
	}

	t.Run("get specific profile", func(t *testing.T) {
		profile, err := cfg.GetProfile("prod")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if profile.URL != "https://prod.example.com" {
			t.Errorf("expected prod URL, got %s", profile.URL)
		}
	})

	t.Run("get current profile", func(t *testing.T) {
		profile, err := cfg.GetProfile("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if profile.URL != "https://default.example.com" {
			t.Errorf("expected default URL, got %s", profile.URL)
		}
	})

	t.Run("get non-existent profile", func(t *testing.T) {
		_, err := cfg.GetProfile("nonexistent")
		if err == nil {
			t.Error("expected error for non-existent profile")
		}
	})

	t.Run("no profile specified and no current", func(t *testing.T) {
		cfg.CurrentProfile = ""
		_, err := cfg.GetProfile("")
		if err == nil {
			t.Error("expected error when no profile specified and no current profile")
		}
	})
}

func TestConfig_DeleteProfile(t *testing.T) {
	cfg := &Config{
		CurrentProfile: "test",
		Profiles: map[string]*Profile{
			"test": {URL: "https://test.example.com"},
			"prod": {URL: "https://prod.example.com"},
		},
	}

	t.Run("delete existing profile", func(t *testing.T) {
		err := cfg.DeleteProfile("test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(cfg.Profiles) != 1 {
			t.Errorf("expected 1 profile remaining, got %d", len(cfg.Profiles))
		}
		if cfg.CurrentProfile != "" {
			t.Error("current profile should be cleared when deleted")
		}
	})

	t.Run("delete non-existent profile", func(t *testing.T) {
		err := cfg.DeleteProfile("nonexistent")
		if err == nil {
			t.Error("expected error when deleting non-existent profile")
		}
	})
}

func TestConfig_SetCurrentProfile(t *testing.T) {
	cfg := &Config{
		Profiles: map[string]*Profile{
			"test": {URL: "https://test.example.com"},
		},
	}

	t.Run("set existing profile", func(t *testing.T) {
		err := cfg.SetCurrentProfile("test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.CurrentProfile != "test" {
			t.Errorf("expected current profile 'test', got '%s'", cfg.CurrentProfile)
		}
	})

	t.Run("set non-existent profile", func(t *testing.T) {
		err := cfg.SetCurrentProfile("nonexistent")
		if err == nil {
			t.Error("expected error when setting non-existent profile as current")
		}
	})
}

func TestConfig_ListProfiles(t *testing.T) {
	cfg := &Config{
		Profiles: map[string]*Profile{
			"test": {URL: "https://test.example.com"},
			"prod": {URL: "https://prod.example.com"},
			"dev":  {URL: "https://dev.example.com"},
		},
	}

	profiles := cfg.ListProfiles()
	if len(profiles) != 3 {
		t.Errorf("expected 3 profiles, got %d", len(profiles))
	}

	profileMap := make(map[string]bool)
	for _, name := range profiles {
		profileMap[name] = true
	}

	if !profileMap["test"] || !profileMap["prod"] || !profileMap["dev"] {
		t.Error("not all profiles returned")
	}
}

func TestProfile_Validate(t *testing.T) {
	tests := []struct {
		name      string
		profile   *Profile
		wantError bool
	}{
		{
			name: "valid with API key",
			profile: &Profile{
				URL:    "https://test.example.com",
				APIKey: "test-key",
			},
			wantError: false,
		},
		{
			name: "valid with username",
			profile: &Profile{
				URL:      "https://test.example.com",
				Username: "admin",
			},
			wantError: false,
		},
		{
			name: "valid with token",
			profile: &Profile{
				URL:   "https://test.example.com",
				Token: "jwt-token",
			},
			wantError: false,
		},
		{
			name: "missing URL",
			profile: &Profile{
				APIKey: "test-key",
			},
			wantError: true,
		},
		{
			name: "missing auth method",
			profile: &Profile{
				URL: "https://test.example.com",
			},
			wantError: true,
		},
		{
			name:      "empty profile",
			profile:   &Profile{},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.profile.Validate()
			if tt.wantError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestConfig_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	cfg := &Config{
		CurrentProfile: "test",
		Profiles: map[string]*Profile{
			"test": {
				URL:      "https://test.example.com",
				APIKey:   "test-key",
				Username: "admin",
				Insecure: true,
			},
		},
	}

	err := cfg.Save()
	if err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	configPath, _ := GetConfigPath()
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("config file not created: %v", err)
	}

	if info.Mode().Perm() != 0600 {
		t.Errorf("expected permissions 0600, got %o", info.Mode().Perm())
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if loaded.CurrentProfile != cfg.CurrentProfile {
		t.Errorf("expected current profile '%s', got '%s'", cfg.CurrentProfile, loaded.CurrentProfile)
	}

	if len(loaded.Profiles) != len(cfg.Profiles) {
		t.Errorf("expected %d profiles, got %d", len(cfg.Profiles), len(loaded.Profiles))
	}

	testProfile := loaded.Profiles["test"]
	if testProfile.URL != "https://test.example.com" {
		t.Errorf("expected URL 'https://test.example.com', got '%s'", testProfile.URL)
	}

	if testProfile.APIKey != "test-key" {
		t.Errorf("expected API key 'test-key', got '%s'", testProfile.APIKey)
	}

	if !testProfile.Insecure {
		t.Error("expected Insecure to be true")
	}
}

func TestLoad_NonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error loading non-existent config: %v", err)
	}

	if cfg == nil {
		t.Fatal("config should not be nil")
	}

	if cfg.Profiles == nil {
		t.Error("profiles map should be initialized")
	}

	if len(cfg.Profiles) != 0 {
		t.Errorf("expected 0 profiles, got %d", len(cfg.Profiles))
	}
}

func TestEnsureConfigDir(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	err := EnsureConfigDir()
	if err != nil {
		t.Fatalf("failed to ensure config dir: %v", err)
	}

	configDir, _ := GetConfigDir()
	info, err := os.Stat(configDir)
	if err != nil {
		t.Fatalf("config dir not created: %v", err)
	}

	if !info.IsDir() {
		t.Error("config path should be a directory")
	}

	if info.Mode().Perm() != 0700 {
		t.Errorf("expected directory permissions 0700, got %o", info.Mode().Perm())
	}
}
