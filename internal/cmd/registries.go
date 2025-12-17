package cmd

import (
	"fmt"

	"github.com/robversluis/portainer-cli/internal/client"
	"github.com/robversluis/portainer-cli/internal/config"
	"github.com/robversluis/portainer-cli/internal/output"
	"github.com/spf13/cobra"
)

var registriesCmd = &cobra.Command{
	Use:   "registries",
	Short: "Manage container registries",
	Long:  `List and manage container registries for pulling and pushing images.`,
}

var registriesListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List registries",
	Long:    `Display a list of all configured container registries.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		profile, err := config.GetProfileFromViper()
		if err != nil {
			return fmt.Errorf("failed to get profile: %w", err)
		}

		c, err := client.NewClient(profile, client.WithVerbose(GetVerbose()))
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		registryService := client.NewRegistryService(c)
		registries, err := registryService.List()
		if err != nil {
			return err
		}

		format := output.ParseFormat(cmd.Flag("output").Value.String())

		switch format {
		case output.FormatJSON, output.FormatYAML:
			formatter := output.NewFormatter(output.Options{Format: format})
			return formatter.Format(registries)

		default:
			table := output.NewTableData([]string{"ID", "Name", "Type", "URL", "Auth"})
			for _, registry := range registries {
				table.AddRow([]string{
					fmt.Sprintf("%d", registry.Id),
					registry.Name,
					registry.TypeString(),
					registry.URL,
					output.FormatBool(registry.Authentication),
				})
			}
			return output.PrintTable(*table)
		}
	},
}

var registriesGetCmd = &cobra.Command{
	Use:   "get [id]",
	Short: "Get registry details",
	Long:  `Retrieve detailed information about a specific registry.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var registryID int
		if _, err := fmt.Sscanf(args[0], "%d", &registryID); err != nil {
			return fmt.Errorf("invalid registry ID: %s", args[0])
		}

		profile, err := config.GetProfileFromViper()
		if err != nil {
			return fmt.Errorf("failed to get profile: %w", err)
		}

		c, err := client.NewClient(profile, client.WithVerbose(GetVerbose()))
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		registryService := client.NewRegistryService(c)
		registry, err := registryService.Get(registryID)
		if err != nil {
			return err
		}

		format := output.ParseFormat(cmd.Flag("output").Value.String())

		switch format {
		case output.FormatJSON, output.FormatYAML:
			formatter := output.NewFormatter(output.Options{Format: format})
			return formatter.Format(registry)

		default:
			fmt.Printf("ID:             %d\n", registry.Id)
			fmt.Printf("Name:           %s\n", registry.Name)
			fmt.Printf("Type:           %s\n", registry.TypeString())
			fmt.Printf("URL:            %s\n", registry.URL)
			fmt.Printf("Authentication: %s\n", output.FormatBool(registry.Authentication))

			if registry.Authentication {
				fmt.Printf("Username:       %s\n", registry.Username)
			}

			return nil
		}
	},
}

var registriesDeleteCmd = &cobra.Command{
	Use:     "delete [id]",
	Aliases: []string{"rm"},
	Short:   "Delete a registry",
	Long:    `Remove a registry configuration.`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var registryID int
		if _, err := fmt.Sscanf(args[0], "%d", &registryID); err != nil {
			return fmt.Errorf("invalid registry ID: %s", args[0])
		}

		profile, err := config.GetProfileFromViper()
		if err != nil {
			return fmt.Errorf("failed to get profile: %w", err)
		}

		c, err := client.NewClient(profile, client.WithVerbose(GetVerbose()))
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		registryService := client.NewRegistryService(c)
		if err := registryService.Delete(registryID); err != nil {
			return err
		}

		if !GetQuiet() {
			fmt.Printf("Registry %d deleted successfully\n", registryID)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(registriesCmd)
	registriesCmd.AddCommand(registriesListCmd)
	registriesCmd.AddCommand(registriesGetCmd)
	registriesCmd.AddCommand(registriesDeleteCmd)
}
