package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/robversluis/portainer-cli/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
	Long:  `Manage configuration profiles, settings, and credentials for the Portainer CLI.`,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration",
	Long:  `Create the configuration directory and file with default settings.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.EnsureConfigDir(); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		cfg := &config.Config{
			Profiles: make(map[string]*config.Profile),
		}

		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		configPath, err := config.GetConfigPath()
		if err != nil {
			return fmt.Errorf("failed to get config path: %w", err)
		}
		if !GetQuiet() {
			fmt.Printf("Configuration initialized at: %s\n", configPath)
		}
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value for the current or specified profile.

Examples:
  portainer-cli config set url https://portainer.example.com
  portainer-cli config set api_key YOUR_API_KEY
  portainer-cli config set --profile prod url https://prod.example.com`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		profileName, err := cmd.Flags().GetString("profile")
		if err != nil {
			return err
		}

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if profileName == "" {
			profileName = cfg.CurrentProfile
		}

		if profileName == "" {
			return fmt.Errorf("no profile specified and no current profile set. Use --profile flag or set a current profile")
		}

		profile, exists := cfg.Profiles[profileName]
		if !exists {
			profile = &config.Profile{}
			cfg.SetProfile(profileName, profile)
		}

		switch key {
		case "url":
			profile.URL = value
		case "api_key":
			profile.APIKey = value
		case "username":
			profile.Username = value
		case "token":
			profile.Token = value
		case "insecure":
			profile.Insecure = strings.ToLower(value) == "true"
		default:
			return fmt.Errorf("unknown configuration key: %s", key)
		}

		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("Set %s = %s for profile '%s'\n", key, value, profileName)
		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get configuration value(s)",
	Long: `Display configuration values for the current or specified profile.

Examples:
  portainer-cli config get
  portainer-cli config get url
  portainer-cli config get --profile prod`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		profileName, err := cmd.Flags().GetString("profile")
		if err != nil {
			return err
		}
		if profileName == "" {
			profileName = cfg.CurrentProfile
		}

		if profileName == "" {
			return fmt.Errorf("no profile specified and no current profile set")
		}

		profile, err := cfg.GetProfile(profileName)
		if err != nil {
			return err
		}

		if len(args) == 0 {
			fmt.Printf("Profile: %s\n", profileName)
			fmt.Printf("URL: %s\n", profile.URL)
			fmt.Printf("API Key: %s\n", maskSecret(profile.APIKey))
			fmt.Printf("Username: %s\n", profile.Username)
			fmt.Printf("Token: %s\n", maskSecret(profile.Token))
			fmt.Printf("Insecure: %t\n", profile.Insecure)
		} else {
			key := args[0]
			switch key {
			case "url":
				fmt.Println(profile.URL)
			case "api_key":
				fmt.Println(profile.APIKey)
			case "username":
				fmt.Println(profile.Username)
			case "token":
				fmt.Println(profile.Token)
			case "insecure":
				fmt.Println(profile.Insecure)
			default:
				return fmt.Errorf("unknown configuration key: %s", key)
			}
		}

		return nil
	},
}

var configListProfilesCmd = &cobra.Command{
	Use:     "list-profiles",
	Aliases: []string{"profiles"},
	Short:   "List all profiles",
	Long:    `Display a list of all configured profiles.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if len(cfg.Profiles) == 0 {
			fmt.Println("No profiles configured")
			return nil
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Profile", "URL", "Auth Method", "Current"})
		table.SetBorder(false)
		table.SetColumnSeparator("")
		table.SetHeaderLine(false)

		for name, profile := range cfg.Profiles {
			authMethod := "None"
			if profile.APIKey != "" {
				authMethod = "API Key"
			} else if profile.Token != "" {
				authMethod = "Token"
			} else if profile.Username != "" {
				authMethod = "Username"
			}

			current := ""
			if name == cfg.CurrentProfile {
				current = "*"
			}

			table.Append([]string{name, profile.URL, authMethod, current})
		}

		table.Render()
		return nil
	},
}

var configUseProfileCmd = &cobra.Command{
	Use:   "use-profile <name>",
	Short: "Set the current profile",
	Long:  `Set the specified profile as the current/default profile.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := args[0]

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if err := cfg.SetCurrentProfile(profileName); err != nil {
			return err
		}

		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("Switched to profile '%s'\n", profileName)
		return nil
	},
}

var configCreateProfileCmd = &cobra.Command{
	Use:   "create-profile <name>",
	Short: "Create a new profile",
	Long: `Create a new configuration profile.

Examples:
  portainer-cli config create-profile production --url https://portainer.prod.com --api-key KEY`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := args[0]

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if _, exists := cfg.Profiles[profileName]; exists {
			return fmt.Errorf("profile '%s' already exists", profileName)
		}

		url, err := cmd.Flags().GetString("url")
		if err != nil {
			return err
		}
		apiKey, err := cmd.Flags().GetString("api-key")
		if err != nil {
			return err
		}
		username, err := cmd.Flags().GetString("username")
		if err != nil {
			return err
		}
		insecure, err := cmd.Flags().GetBool("insecure")
		if err != nil {
			return err
		}

		profile := &config.Profile{
			URL:      url,
			APIKey:   apiKey,
			Username: username,
			Insecure: insecure,
		}

		cfg.SetProfile(profileName, profile)

		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("Created profile '%s'\n", profileName)
		return nil
	},
}

var configDeleteProfileCmd = &cobra.Command{
	Use:     "delete-profile <name>",
	Aliases: []string{"remove-profile"},
	Short:   "Delete a profile",
	Long:    `Remove a configuration profile.`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := args[0]

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if err := cfg.DeleteProfile(profileName); err != nil {
			return err
		}

		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("Deleted profile '%s'\n", profileName)
		return nil
	},
}

var configViewCmd = &cobra.Command{
	Use:   "view",
	Short: "View the entire configuration",
	Long:  `Display the complete configuration file contents.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath, err := config.GetConfigPath()
		if err != nil {
			return err
		}

		data, err := os.ReadFile(configPath)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("No configuration file found")
				return nil
			}
			return fmt.Errorf("failed to read config: %w", err)
		}

		fmt.Println(string(data))
		return nil
	},
}

func maskSecret(secret string) string {
	if secret == "" {
		return ""
	}
	if len(secret) <= 8 {
		return "****"
	}
	return secret[:4] + "****" + secret[len(secret)-4:]
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configListProfilesCmd)
	configCmd.AddCommand(configUseProfileCmd)
	configCmd.AddCommand(configCreateProfileCmd)
	configCmd.AddCommand(configDeleteProfileCmd)
	configCmd.AddCommand(configViewCmd)

	configSetCmd.Flags().String("profile", "", "Profile to modify")
	configGetCmd.Flags().String("profile", "", "Profile to view")

	configCreateProfileCmd.Flags().String("url", "", "Portainer URL")
	configCreateProfileCmd.Flags().String("api-key", "", "API key")
	configCreateProfileCmd.Flags().String("username", "", "Username")
	configCreateProfileCmd.Flags().Bool("insecure", false, "Skip TLS verification")
}
