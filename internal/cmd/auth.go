package cmd

import (
	"fmt"
	"syscall"

	"github.com/rob/portainer-cli/internal/client"
	"github.com/rob/portainer-cli/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication operations",
	Long:  `Manage authentication with Portainer API including login, logout, and status checks.`,
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to Portainer",
	Long: `Authenticate with Portainer using username and password.
The JWT token will be stored in the current profile for future use.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		username, err := cmd.Flags().GetString("username")
		if err != nil {
			return err
		}
		password, err := cmd.Flags().GetString("password")
		if err != nil {
			return err
		}

		if username == "" {
			fmt.Print("Username: ")
			if _, err := fmt.Scanln(&username); err != nil {
				return fmt.Errorf("failed to read username: %w", err)
			}
		}

		if password == "" {
			fmt.Print("Password: ")
			passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
			fmt.Println()
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}
			password = string(passwordBytes)
		}

		if username == "" || password == "" {
			return fmt.Errorf("username and password are required")
		}

		profile, err := config.GetProfileFromViper()
		if err != nil {
			return fmt.Errorf("failed to get profile: %w", err)
		}

		if GetVerbose() {
			fmt.Printf("Logging in to %s as %s...\n", profile.URL, username)
		}

		token, err := client.LoginAndSaveToken(profile, username, password)
		if err != nil {
			return fmt.Errorf("login failed: %w", err)
		}

		if !GetQuiet() {
			fmt.Printf("Successfully logged in as %s\n", username)
			if GetVerbose() {
				fmt.Printf("Token: %s\n", token)
			}
		}

		return nil
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from Portainer",
	Long:  `Clear stored authentication token from the current profile.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		profileName := cfg.CurrentProfile
		if profileName == "" {
			return fmt.Errorf("no current profile set")
		}

		profile, err := cfg.GetProfile(profileName)
		if err != nil {
			return err
		}

		profile.Token = ""

		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		if !GetQuiet() {
			fmt.Printf("Logged out from profile '%s'\n", profileName)
		}

		return nil
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check authentication status",
	Long:  `Display current authentication status and validate credentials.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		profile, err := config.GetProfileFromViper()
		if err != nil {
			return fmt.Errorf("failed to get profile: %w", err)
		}

		c, err := client.NewClient(profile, client.WithVerbose(GetVerbose()))
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		authService := client.NewAuthService(c)

		status, err := authService.GetStatus()
		if err != nil {
			return fmt.Errorf("failed to get status: %w", err)
		}

		fmt.Printf("Portainer URL: %s\n", profile.URL)
		fmt.Printf("Portainer Version: %s\n", status.Version)

		authMethod := "None"
		if profile.APIKey != "" {
			authMethod = "API Key"
		} else if profile.Token != "" {
			authMethod = "JWT Token"
		} else if profile.Username != "" {
			authMethod = "Username (no token)"
		}
		fmt.Printf("Authentication Method: %s\n", authMethod)

		if profile.Token != "" || profile.APIKey != "" {
			userInfo, err := authService.ValidateToken()
			if err != nil {
				fmt.Printf("Authentication Status: Invalid (%v)\n", err)
				return nil
			}

			fmt.Printf("Authentication Status: Valid\n")
			fmt.Printf("Logged in as: %s (ID: %d, Role: %d)\n", userInfo.Username, userInfo.ID, userInfo.Role)
		} else {
			fmt.Printf("Authentication Status: Not authenticated\n")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authLogoutCmd)
	authCmd.AddCommand(authStatusCmd)

	authLoginCmd.Flags().String("username", "", "Username for authentication")
	authLoginCmd.Flags().String("password", "", "Password for authentication")
}
