package cmd

import (
	"fmt"

	"github.com/rob/portainer-cli/internal/client"
	"github.com/rob/portainer-cli/internal/config"
	"github.com/rob/portainer-cli/internal/output"
	"github.com/spf13/cobra"
)

var networksCmd = &cobra.Command{
	Use:   "networks",
	Short: "Manage Docker networks",
	Long:  `List, create, inspect, and manage Docker networks across environments.`,
}

var networksListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List networks",
	Long:    `Display a list of Docker networks in the specified environment.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		endpointID, err := cmd.Flags().GetInt("endpoint")
		if err != nil {
			return err
		}
		if endpointID == 0 {
			return fmt.Errorf("--endpoint flag is required")
		}

		profile, err := config.GetProfileFromViper()
		if err != nil {
			return fmt.Errorf("failed to get profile: %w", err)
		}

		c, err := client.NewClient(profile, client.WithVerbose(GetVerbose()))
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		networkService := client.NewNetworkService(c)
		networks, err := networkService.List(endpointID)
		if err != nil {
			return err
		}

		format := output.ParseFormat(cmd.Flag("output").Value.String())

		switch format {
		case output.FormatJSON, output.FormatYAML:
			formatter := output.NewFormatter(output.Options{Format: format})
			return formatter.Format(networks)

		default:
			table := output.NewTableData([]string{"ID", "Name", "Driver", "Scope"})
			for _, network := range networks {
				table.AddRow([]string{
					network.GetShortID(),
					network.Name,
					network.Driver,
					network.Scope,
				})
			}
			return output.PrintTable(*table)
		}
	},
}

var networksInspectCmd = &cobra.Command{
	Use:   "inspect [network]",
	Short: "Inspect a network",
	Long:  `Display detailed information about a specific network.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		endpointID, err := cmd.Flags().GetInt("endpoint")
		if err != nil {
			return err
		}
		if endpointID == 0 {
			return fmt.Errorf("--endpoint flag is required")
		}

		networkID := args[0]

		profile, err := config.GetProfileFromViper()
		if err != nil {
			return fmt.Errorf("failed to get profile: %w", err)
		}

		c, err := client.NewClient(profile, client.WithVerbose(GetVerbose()))
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		networkService := client.NewNetworkService(c)
		network, err := networkService.Inspect(endpointID, networkID)
		if err != nil {
			return err
		}

		format := output.ParseFormat(cmd.Flag("output").Value.String())

		switch format {
		case output.FormatJSON, output.FormatYAML:
			formatter := output.NewFormatter(output.Options{Format: format})
			return formatter.Format(network)

		default:
			fmt.Printf("ID:         %s\n", network.Id)
			fmt.Printf("Name:       %s\n", network.Name)
			fmt.Printf("Driver:     %s\n", network.Driver)
			fmt.Printf("Scope:      %s\n", network.Scope)
			fmt.Printf("Internal:   %s\n", output.FormatBool(network.Internal))
			fmt.Printf("Attachable: %s\n", output.FormatBool(network.Attachable))
			fmt.Printf("IPv6:       %s\n", output.FormatBool(network.EnableIPv6))

			if len(network.IPAM.Config) > 0 {
				fmt.Printf("\nIPAM Config:\n")
				for i, config := range network.IPAM.Config {
					fmt.Printf("  [%d] Subnet:  %s\n", i, config.Subnet)
					if config.Gateway != "" {
						fmt.Printf("      Gateway: %s\n", config.Gateway)
					}
					if config.IPRange != "" {
						fmt.Printf("      IPRange: %s\n", config.IPRange)
					}
				}
			}

			if len(network.Containers) > 0 {
				fmt.Printf("\nConnected Containers:\n")
				for _, container := range network.Containers {
					fmt.Printf("  %s: %s\n", container.Name, container.IPv4Address)
				}
			}

			if len(network.Labels) > 0 {
				fmt.Printf("\nLabels:\n")
				for k, v := range network.Labels {
					fmt.Printf("  %s=%s\n", k, v)
				}
			}

			return nil
		}
	},
}

var networksCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a network",
	Long:  `Create a new Docker network.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		endpointID, err := cmd.Flags().GetInt("endpoint")
		if err != nil {
			return err
		}
		if endpointID == 0 {
			return fmt.Errorf("--endpoint flag is required")
		}

		networkName := args[0]
		driver, err := cmd.Flags().GetString("driver")
		if err != nil {
			return err
		}
		internal, err := cmd.Flags().GetBool("internal")
		if err != nil {
			return err
		}
		attachable, err := cmd.Flags().GetBool("attachable")
		if err != nil {
			return err
		}

		profile, err := config.GetProfileFromViper()
		if err != nil {
			return fmt.Errorf("failed to get profile: %w", err)
		}

		c, err := client.NewClient(profile, client.WithVerbose(GetVerbose()))
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		req := &client.NetworkCreateRequest{
			Name:       networkName,
			Driver:     driver,
			Internal:   internal,
			Attachable: attachable,
		}

		networkService := client.NewNetworkService(c)
		response, err := networkService.Create(endpointID, req)
		if err != nil {
			return err
		}

		if !GetQuiet() {
			fmt.Printf("Network '%s' created successfully (ID: %s)\n", networkName, response.Id[:12])
			if response.Warning != "" {
				fmt.Printf("Warning: %s\n", response.Warning)
			}
		}

		return nil
	},
}

var networksRemoveCmd = &cobra.Command{
	Use:     "remove [network]",
	Aliases: []string{"rm"},
	Short:   "Remove a network",
	Long:    `Remove a Docker network.`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		endpointID, err := cmd.Flags().GetInt("endpoint")
		if err != nil {
			return err
		}
		if endpointID == 0 {
			return fmt.Errorf("--endpoint flag is required")
		}

		networkID := args[0]

		profile, err := config.GetProfileFromViper()
		if err != nil {
			return fmt.Errorf("failed to get profile: %w", err)
		}

		c, err := client.NewClient(profile, client.WithVerbose(GetVerbose()))
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		networkService := client.NewNetworkService(c)
		if err := networkService.Remove(endpointID, networkID); err != nil {
			return err
		}

		if !GetQuiet() {
			fmt.Printf("Network '%s' removed successfully\n", networkID)
		}

		return nil
	},
}

var networksPruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove unused networks",
	Long:  `Remove all unused networks.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		endpointID, err := cmd.Flags().GetInt("endpoint")
		if err != nil {
			return err
		}
		if endpointID == 0 {
			return fmt.Errorf("--endpoint flag is required")
		}

		profile, err := config.GetProfileFromViper()
		if err != nil {
			return fmt.Errorf("failed to get profile: %w", err)
		}

		c, err := client.NewClient(profile, client.WithVerbose(GetVerbose()))
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		networkService := client.NewNetworkService(c)
		if err := networkService.Prune(endpointID); err != nil {
			return err
		}

		if !GetQuiet() {
			fmt.Println("Networks pruned successfully")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(networksCmd)
	networksCmd.AddCommand(networksListCmd)
	networksCmd.AddCommand(networksInspectCmd)
	networksCmd.AddCommand(networksCreateCmd)
	networksCmd.AddCommand(networksRemoveCmd)
	networksCmd.AddCommand(networksPruneCmd)

	networksListCmd.Flags().Int("endpoint", 0, "Environment endpoint ID (required)")
	networksListCmd.MarkFlagRequired("endpoint")

	networksInspectCmd.Flags().Int("endpoint", 0, "Environment endpoint ID (required)")
	networksInspectCmd.MarkFlagRequired("endpoint")

	networksCreateCmd.Flags().Int("endpoint", 0, "Environment endpoint ID (required)")
	networksCreateCmd.Flags().String("driver", "bridge", "Network driver")
	networksCreateCmd.Flags().Bool("internal", false, "Restrict external access to the network")
	networksCreateCmd.Flags().Bool("attachable", false, "Enable manual container attachment")
	networksCreateCmd.MarkFlagRequired("endpoint")

	networksRemoveCmd.Flags().Int("endpoint", 0, "Environment endpoint ID (required)")
	networksRemoveCmd.MarkFlagRequired("endpoint")

	networksPruneCmd.Flags().Int("endpoint", 0, "Environment endpoint ID (required)")
	networksPruneCmd.MarkFlagRequired("endpoint")
}
