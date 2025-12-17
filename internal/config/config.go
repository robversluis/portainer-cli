package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type Config struct {
	CurrentProfile string              `yaml:"current_profile" mapstructure:"current_profile"`
	Profiles       map[string]*Profile `yaml:"profiles" mapstructure:"profiles"`
}

type Profile struct {
	Name     string `yaml:"name,omitempty" mapstructure:"name"`
	URL      string `yaml:"url" mapstructure:"url"`
	APIKey   string `yaml:"api_key,omitempty" mapstructure:"api_key"`
	Username string `yaml:"username,omitempty" mapstructure:"username"`
	Token    string `yaml:"token,omitempty" mapstructure:"token"`
	Insecure bool   `yaml:"insecure,omitempty" mapstructure:"insecure"`
}

func GetConfigDir() (string, error) {
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return filepath.Join(xdgConfig, "portainer-cli"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return filepath.Join(home, ".portainer-cli"), nil
}

func GetConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.yaml"), nil
}

func EnsureConfigDir() error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	return nil
}

func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{
			Profiles: make(map[string]*Profile),
		}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]*Profile)
	}

	return &cfg, nil
}

func (c *Config) Save() error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func (c *Config) GetProfile(name string) (*Profile, error) {
	if name == "" {
		name = c.CurrentProfile
	}

	if name == "" {
		return nil, fmt.Errorf("no profile specified and no current profile set")
	}

	profile, exists := c.Profiles[name]
	if !exists {
		return nil, fmt.Errorf("profile '%s' not found", name)
	}

	return profile, nil
}

func (c *Config) SetProfile(name string, profile *Profile) {
	if c.Profiles == nil {
		c.Profiles = make(map[string]*Profile)
	}
	profile.Name = name
	c.Profiles[name] = profile
}

func (c *Config) DeleteProfile(name string) error {
	if _, exists := c.Profiles[name]; !exists {
		return fmt.Errorf("profile '%s' not found", name)
	}

	delete(c.Profiles, name)

	if c.CurrentProfile == name {
		c.CurrentProfile = ""
	}

	return nil
}

func (c *Config) SetCurrentProfile(name string) error {
	if _, exists := c.Profiles[name]; !exists {
		return fmt.Errorf("profile '%s' not found", name)
	}

	c.CurrentProfile = name
	return nil
}

func (c *Config) ListProfiles() []string {
	profiles := make([]string, 0, len(c.Profiles))
	for name := range c.Profiles {
		profiles = append(profiles, name)
	}
	return profiles
}

func (p *Profile) Validate() error {
	if p.URL == "" {
		return fmt.Errorf("URL is required")
	}

	if p.APIKey == "" && p.Username == "" && p.Token == "" {
		return fmt.Errorf("at least one authentication method is required (api_key, username, or token)")
	}

	return nil
}

func GetCurrentProfile() (*Profile, error) {
	cfg, err := Load()
	if err != nil {
		return nil, err
	}

	profileName := viper.GetString("current_profile")
	if profileName == "" {
		profileName = cfg.CurrentProfile
	}

	if profileName == "" {
		return nil, fmt.Errorf("no current profile set")
	}

	return cfg.GetProfile(profileName)
}

func GetProfileFromViper() (*Profile, error) {
	url := viper.GetString("url")
	apiKey := viper.GetString("api_key")
	username := viper.GetString("username")
	token := viper.GetString("token")
	insecure := viper.GetBool("insecure")

	if url == "" {
		profile, err := GetCurrentProfile()
		if err != nil {
			return nil, fmt.Errorf("no URL specified and failed to get current profile: %w", err)
		}
		return profile, nil
	}

	profile := &Profile{
		URL:      url,
		APIKey:   apiKey,
		Username: username,
		Token:    token,
		Insecure: insecure,
	}

	if err := profile.Validate(); err != nil {
		return nil, err
	}

	return profile, nil
}
