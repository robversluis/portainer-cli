package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile      string
	profile      string
	url          string
	apiKey       string
	outputFormat string
	verbose      bool
	quiet        bool
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
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "output format (table, json, yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "quiet mode (minimal output)")

	viper.BindPFlag("url", rootCmd.PersistentFlags().Lookup("url"))
	viper.BindPFlag("api_key", rootCmd.PersistentFlags().Lookup("api-key"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))

	rootCmd.AddCommand(completionCmd)
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

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for portainer-cli.

To load completions:

Bash:
  $ source <(portainer-cli completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ portainer-cli completion bash > /etc/bash_completion.d/portainer-cli
  # macOS:
  $ portainer-cli completion bash > /usr/local/etc/bash_completion.d/portainer-cli

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc
  # To load completions for each session, execute once:
  $ portainer-cli completion zsh > "${fpath[1]}/_portainer-cli"
  # You will need to start a new shell for this setup to take effect.

Fish:
  $ portainer-cli completion fish | source
  # To load completions for each session, execute once:
  $ portainer-cli completion fish > ~/.config/fish/completions/portainer-cli.fish

PowerShell:
  PS> portainer-cli completion powershell | Out-String | Invoke-Expression
  # To load completions for every new session, run:
  PS> portainer-cli completion powershell > portainer-cli.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactValidArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(cmd.OutOrStdout())
		case "zsh":
			return rootCmd.GenZshCompletion(cmd.OutOrStdout())
		case "fish":
			return rootCmd.GenFishCompletion(cmd.OutOrStdout(), true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
		default:
			return fmt.Errorf("unsupported shell: %s", args[0])
		}
	},
}

func GetOutput() string {
	return outputFormat
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
