package cmd

import (
	"fmt"

	"github.com/robversluis/portainer-cli/internal/client"
	"github.com/robversluis/portainer-cli/internal/config"
	"github.com/robversluis/portainer-cli/internal/output"
	"github.com/spf13/cobra"
)

var volumesCmd = &cobra.Command{
	Use:   "volumes",
	Short: "Manage Docker volumes",
	Long:  `List, create, inspect, and manage Docker volumes across environments.`,
}

var volumesListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List volumes",
	Long:    `Display a list of Docker volumes in the specified environment.`,
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

		c, err := client.NewClient(profile, GetClientOptions()...)
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		volumeService := client.NewVolumeService(c)
		volumes, err := volumeService.List(endpointID)
		if err != nil {
			return err
		}

		format := output.ParseFormat(cmd.Flag("output").Value.String())

		switch format {
		case output.FormatJSON, output.FormatYAML:
			formatter := output.NewFormatter(output.Options{Format: format})
			return formatter.Format(volumes)

		default:
			table := output.NewTableData([]string{"Name", "Driver", "Scope", "Mountpoint"})
			for _, volume := range volumes {
				mountpoint := volume.Mountpoint
				if len(mountpoint) > 50 {
					mountpoint = output.TruncateString(mountpoint, 50)
				}
				table.AddRow([]string{
					volume.Name,
					volume.Driver,
					volume.Scope,
					mountpoint,
				})
			}
			return output.PrintTable(*table)
		}
	},
}

var volumesInspectCmd = &cobra.Command{
	Use:   "inspect [volume]",
	Short: "Inspect a volume",
	Long:  `Display detailed information about a specific volume.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		endpointID, err := cmd.Flags().GetInt("endpoint")
		if err != nil {
			return err
		}
		if endpointID == 0 {
			return fmt.Errorf("--endpoint flag is required")
		}

		volumeName := args[0]

		profile, err := config.GetProfileFromViper()
		if err != nil {
			return fmt.Errorf("failed to get profile: %w", err)
		}

		c, err := client.NewClient(profile, GetClientOptions()...)
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		volumeService := client.NewVolumeService(c)
		volume, err := volumeService.Inspect(endpointID, volumeName)
		if err != nil {
			return err
		}

		format := output.ParseFormat(cmd.Flag("output").Value.String())

		switch format {
		case output.FormatJSON, output.FormatYAML:
			formatter := output.NewFormatter(output.Options{Format: format})
			return formatter.Format(volume)

		default:
			fmt.Printf("Name:       %s\n", volume.Name)
			fmt.Printf("Driver:     %s\n", volume.Driver)
			fmt.Printf("Scope:      %s\n", volume.Scope)
			fmt.Printf("Mountpoint: %s\n", volume.Mountpoint)
			fmt.Printf("Created:    %s\n", volume.CreatedAt)

			if len(volume.Labels) > 0 {
				fmt.Printf("\nLabels:\n")
				for k, v := range volume.Labels {
					fmt.Printf("  %s=%s\n", k, v)
				}
			}

			if len(volume.Options) > 0 {
				fmt.Printf("\nOptions:\n")
				for k, v := range volume.Options {
					fmt.Printf("  %s=%s\n", k, v)
				}
			}

			if volume.UsageData != nil {
				fmt.Printf("\nUsage:\n")
				fmt.Printf("  Size:      %s\n", output.FormatSize(volume.UsageData.Size))
				fmt.Printf("  RefCount:  %d\n", volume.UsageData.RefCount)
			}

			return nil
		}
	},
}

var volumesCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a volume",
	Long:  `Create a new Docker volume.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		endpointID, err := cmd.Flags().GetInt("endpoint")
		if err != nil {
			return err
		}
		if endpointID == 0 {
			return fmt.Errorf("--endpoint flag is required")
		}

		volumeName := args[0]
		driver, err := cmd.Flags().GetString("driver")
		if err != nil {
			return err
		}

		profile, err := config.GetProfileFromViper()
		if err != nil {
			return fmt.Errorf("failed to get profile: %w", err)
		}

		c, err := client.NewClient(profile, GetClientOptions()...)
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		req := &client.VolumeCreateRequest{
			Name:   volumeName,
			Driver: driver,
		}

		volumeService := client.NewVolumeService(c)
		volume, err := volumeService.Create(endpointID, req)
		if err != nil {
			return err
		}

		if !GetQuiet() {
			fmt.Printf("Volume '%s' created successfully\n", volume.Name)
		}

		return nil
	},
}

var volumesRemoveCmd = &cobra.Command{
	Use:     "remove [volume]",
	Aliases: []string{"rm"},
	Short:   "Remove a volume",
	Long:    `Remove a Docker volume.`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		endpointID, err := cmd.Flags().GetInt("endpoint")
		if err != nil {
			return err
		}
		if endpointID == 0 {
			return fmt.Errorf("--endpoint flag is required")
		}

		volumeName := args[0]
		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}

		profile, err := config.GetProfileFromViper()
		if err != nil {
			return fmt.Errorf("failed to get profile: %w", err)
		}

		c, err := client.NewClient(profile, GetClientOptions()...)
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		volumeService := client.NewVolumeService(c)
		if err := volumeService.Remove(endpointID, volumeName, force); err != nil {
			return err
		}

		if !GetQuiet() {
			fmt.Printf("Volume '%s' removed successfully\n", volumeName)
		}

		return nil
	},
}

var volumesPruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove unused volumes",
	Long:  `Remove all unused local volumes.`,
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

		c, err := client.NewClient(profile, GetClientOptions()...)
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		volumeService := client.NewVolumeService(c)
		if err := volumeService.Prune(endpointID); err != nil {
			return err
		}

		if !GetQuiet() {
			fmt.Println("Volumes pruned successfully")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(volumesCmd)
	volumesCmd.AddCommand(volumesListCmd)
	volumesCmd.AddCommand(volumesInspectCmd)
	volumesCmd.AddCommand(volumesCreateCmd)
	volumesCmd.AddCommand(volumesRemoveCmd)
	volumesCmd.AddCommand(volumesPruneCmd)

	volumesListCmd.Flags().Int("endpoint", 0, "Environment endpoint ID (required)")
	_ = volumesListCmd.MarkFlagRequired("endpoint")

	volumesInspectCmd.Flags().Int("endpoint", 0, "Environment endpoint ID (required)")
	_ = volumesInspectCmd.MarkFlagRequired("endpoint")

	volumesCreateCmd.Flags().Int("endpoint", 0, "Environment endpoint ID (required)")
	volumesCreateCmd.Flags().String("driver", "local", "Volume driver")
	_ = volumesCreateCmd.MarkFlagRequired("endpoint")

	volumesRemoveCmd.Flags().Int("endpoint", 0, "Environment endpoint ID (required)")
	volumesRemoveCmd.Flags().BoolP("force", "f", false, "Force removal of the volume")
	_ = volumesRemoveCmd.MarkFlagRequired("endpoint")

	volumesPruneCmd.Flags().Int("endpoint", 0, "Environment endpoint ID (required)")
	_ = volumesPruneCmd.MarkFlagRequired("endpoint")
}
