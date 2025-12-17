package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	profile string
	url     string
	apiKey  string
	output  string
	verbose bool
	quiet   bool
)

var rootCmd = &cobra.Command{
	Use:   "portainer-cli",
	Short: "A CLI tool for managing Portainer",
	Long: `Portainer CLI is a command-line interface tool that provides comprehensive 
access to Portainer's API, enabling automation and scriptable container 
orchestration workflows.

Manage Docker, Kubernetes, and Edge environments from the terminal without 
requiring the web UI.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.portainer-cli/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&profile, "profile", "", "profile/context to use")
	rootCmd.PersistentFlags().StringVar(&url, "url", "", "Portainer URL (overrides config)")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "API key for authentication (overrides config)")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "table", "output format (table, json, yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "quiet mode (minimal output)")

	viper.BindPFlag("url", rootCmd.PersistentFlags().Lookup("url"))
	viper.BindPFlag("api_key", rootCmd.PersistentFlags().Lookup("api-key"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
			os.Exit(1)
		}

		configDir := home + "/.portainer-cli"
		viper.AddConfigPath(configDir)
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.SetEnvPrefix("PORTAINER")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintf(os.Stderr, "Error reading config file: %v\n", err)
		}
	}

	if profile != "" {
		viper.Set("current_profile", profile)
	}

	currentProfile := viper.GetString("current_profile")
	if currentProfile != "" {
		profileConfig := viper.Sub("profiles." + currentProfile)
		if profileConfig != nil {
			if url == "" {
				url = profileConfig.GetString("url")
				viper.Set("url", url)
			}
			if apiKey == "" {
				apiKey = profileConfig.GetString("api_key")
				viper.Set("api_key", apiKey)
			}
		}
	}
}

func GetVerbose() bool {
	return verbose
}

func GetQuiet() bool {
	return quiet
}

func GetOutput() string {
	return output
}

func GetURL() string {
	if url != "" {
		return url
	}
	return viper.GetString("url")
}

func GetAPIKey() string {
	if apiKey != "" {
		return apiKey
	}
	return viper.GetString("api_key")
}
