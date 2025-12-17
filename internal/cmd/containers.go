package cmd

import (
	"bufio"
	"fmt"
	"io"

	"github.com/robversluis/portainer-cli/internal/client"
	"github.com/robversluis/portainer-cli/internal/config"
	"github.com/robversluis/portainer-cli/internal/output"
	"github.com/spf13/cobra"
)

var containersCmd = &cobra.Command{
	Use:   "containers",
	Short: "Manage Docker containers",
	Long:  `List, start, stop, and manage Docker containers across environments.`,
}

var containersListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List containers",
	Long:    `Display a list of containers in the specified environment.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		endpointID, _ := cmd.Flags().GetInt("endpoint")
		if endpointID == 0 {
			return fmt.Errorf("--endpoint flag is required")
		}

		all, _ := cmd.Flags().GetBool("all")

		profile, err := config.GetProfileFromViper()
		if err != nil {
			return fmt.Errorf("failed to get profile: %w", err)
		}

		c, err := client.NewClient(profile, client.WithVerbose(GetVerbose()))
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		containerService := client.NewContainerService(c)
		containers, err := containerService.List(endpointID, all)
		if err != nil {
			return err
		}

		format := output.ParseFormat(cmd.Flag("output").Value.String())

		switch format {
		case output.FormatJSON, output.FormatYAML:
			formatter := output.NewFormatter(output.Options{Format: format})
			return formatter.Format(containers)

		default:
			table := output.NewTableData([]string{"ID", "Name", "Image", "Status", "Ports"})
			for _, container := range containers {
				ports := container.GetPorts()
				if len(ports) > 50 {
					ports = output.TruncateString(ports, 50)
				}
				table.AddRow([]string{
					container.GetShortID(),
					container.GetName(),
					container.Image,
					container.GetStatus(),
					ports,
				})
			}
			return output.PrintTable(*table)
		}
	},
}

var containersLogsCmd = &cobra.Command{
	Use:   "logs [container]",
	Short: "View container logs",
	Long:  `Display logs from a specific container.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		endpointID, _ := cmd.Flags().GetInt("endpoint")
		if endpointID == 0 {
			return fmt.Errorf("--endpoint flag is required")
		}

		containerID := args[0]
		follow, _ := cmd.Flags().GetBool("follow")
		tail, _ := cmd.Flags().GetInt("tail")

		profile, err := config.GetProfileFromViper()
		if err != nil {
			return fmt.Errorf("failed to get profile: %w", err)
		}

		c, err := client.NewClient(profile, client.WithVerbose(GetVerbose()))
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		containerService := client.NewContainerService(c)
		logReader, err := containerService.Logs(endpointID, containerID, follow, tail, true, true)
		if err != nil {
			return err
		}
		defer logReader.Close()

		scanner := bufio.NewScanner(logReader)
		for scanner.Scan() {
			line := scanner.Text()
			if len(line) > 8 {
				line = line[8:]
			}
			fmt.Println(line)
		}

		if err := scanner.Err(); err != nil && err != io.EOF {
			return fmt.Errorf("error reading logs: %w", err)
		}

		return nil
	},
}

var containersInspectCmd = &cobra.Command{
	Use:   "inspect [container]",
	Short: "Inspect container details",
	Long:  `Display detailed information about a specific container.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		endpointID, _ := cmd.Flags().GetInt("endpoint")
		if endpointID == 0 {
			return fmt.Errorf("--endpoint flag is required")
		}

		containerID := args[0]

		profile, err := config.GetProfileFromViper()
		if err != nil {
			return fmt.Errorf("failed to get profile: %w", err)
		}

		c, err := client.NewClient(profile, client.WithVerbose(GetVerbose()))
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		containerService := client.NewContainerService(c)
		container, err := containerService.Inspect(endpointID, containerID)
		if err != nil {
			return err
		}

		format := output.ParseFormat(cmd.Flag("output").Value.String())

		switch format {
		case output.FormatJSON, output.FormatYAML:
			formatter := output.NewFormatter(output.Options{Format: format})
			return formatter.Format(container)

		default:
			fmt.Printf("ID:           %s\n", container.Id)
			fmt.Printf("Name:         %s\n", container.Name)
			fmt.Printf("Image:        %s\n", container.Image)
			fmt.Printf("Status:       %s\n", container.State.Status)
			fmt.Printf("Running:      %s\n", output.FormatBool(container.State.Running))
			fmt.Printf("Paused:       %s\n", output.FormatBool(container.State.Paused))
			fmt.Printf("Restarting:   %s\n", output.FormatBool(container.State.Restarting))
			fmt.Printf("Pid:          %d\n", container.State.Pid)
			fmt.Printf("Exit Code:    %d\n", container.State.ExitCode)
			
			if container.State.StartedAt != "" {
				fmt.Printf("Started At:   %s\n", container.State.StartedAt)
			}
			if container.State.FinishedAt != "" && container.State.FinishedAt != "0001-01-01T00:00:00Z" {
				fmt.Printf("Finished At:  %s\n", container.State.FinishedAt)
			}

			if len(container.Config.Env) > 0 {
				fmt.Printf("\nEnvironment:\n")
				for _, env := range container.Config.Env {
					fmt.Printf("  %s\n", env)
				}
			}

			if len(container.Mounts) > 0 {
				fmt.Printf("\nMounts:\n")
				for _, mount := range container.Mounts {
					fmt.Printf("  %s -> %s (%s)\n", mount.Source, mount.Destination, mount.Type)
				}
			}

			if len(container.NetworkSettings.Networks) > 0 {
				fmt.Printf("\nNetworks:\n")
				for name, network := range container.NetworkSettings.Networks {
					fmt.Printf("  %s: %s\n", name, network.IPAddress)
				}
			}

			return nil
		}
	},
}

var containersStartCmd = &cobra.Command{
	Use:   "start [container]",
	Short: "Start a container",
	Long:  `Start one or more stopped containers.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		endpointID, _ := cmd.Flags().GetInt("endpoint")
		if endpointID == 0 {
			return fmt.Errorf("--endpoint flag is required")
		}

		containerID := args[0]

		profile, err := config.GetProfileFromViper()
		if err != nil {
			return fmt.Errorf("failed to get profile: %w", err)
		}

		c, err := client.NewClient(profile, client.WithVerbose(GetVerbose()))
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		containerService := client.NewContainerService(c)
		if err := containerService.Start(endpointID, containerID); err != nil {
			return err
		}

		if !GetQuiet() {
			fmt.Printf("Container %s started\n", containerID)
		}

		return nil
	},
}

var containersStopCmd = &cobra.Command{
	Use:   "stop [container]",
	Short: "Stop a container",
	Long:  `Stop one or more running containers.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		endpointID, _ := cmd.Flags().GetInt("endpoint")
		if endpointID == 0 {
			return fmt.Errorf("--endpoint flag is required")
		}

		containerID := args[0]

		profile, err := config.GetProfileFromViper()
		if err != nil {
			return fmt.Errorf("failed to get profile: %w", err)
		}

		c, err := client.NewClient(profile, client.WithVerbose(GetVerbose()))
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		containerService := client.NewContainerService(c)
		if err := containerService.Stop(endpointID, containerID); err != nil {
			return err
		}

		if !GetQuiet() {
			fmt.Printf("Container %s stopped\n", containerID)
		}

		return nil
	},
}

var containersRestartCmd = &cobra.Command{
	Use:   "restart [container]",
	Short: "Restart a container",
	Long:  `Restart one or more containers.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		endpointID, _ := cmd.Flags().GetInt("endpoint")
		if endpointID == 0 {
			return fmt.Errorf("--endpoint flag is required")
		}

		containerID := args[0]

		profile, err := config.GetProfileFromViper()
		if err != nil {
			return fmt.Errorf("failed to get profile: %w", err)
		}

		c, err := client.NewClient(profile, client.WithVerbose(GetVerbose()))
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		containerService := client.NewContainerService(c)
		if err := containerService.Restart(endpointID, containerID); err != nil {
			return err
		}

		if !GetQuiet() {
			fmt.Printf("Container %s restarted\n", containerID)
		}

		return nil
	},
}

var containersRemoveCmd = &cobra.Command{
	Use:     "remove [container]",
	Aliases: []string{"rm"},
	Short:   "Remove a container",
	Long:    `Remove one or more containers.`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		endpointID, _ := cmd.Flags().GetInt("endpoint")
		if endpointID == 0 {
			return fmt.Errorf("--endpoint flag is required")
		}

		containerID := args[0]
		force, _ := cmd.Flags().GetBool("force")

		profile, err := config.GetProfileFromViper()
		if err != nil {
			return fmt.Errorf("failed to get profile: %w", err)
		}

		c, err := client.NewClient(profile, client.WithVerbose(GetVerbose()))
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		containerService := client.NewContainerService(c)
		if err := containerService.Remove(endpointID, containerID, force); err != nil {
			return err
		}

		if !GetQuiet() {
			fmt.Printf("Container %s removed\n", containerID)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(containersCmd)
	containersCmd.AddCommand(containersListCmd)
	containersCmd.AddCommand(containersLogsCmd)
	containersCmd.AddCommand(containersInspectCmd)
	containersCmd.AddCommand(containersStartCmd)
	containersCmd.AddCommand(containersStopCmd)
	containersCmd.AddCommand(containersRestartCmd)
	containersCmd.AddCommand(containersRemoveCmd)

	containersListCmd.Flags().Int("endpoint", 0, "Environment endpoint ID (required)")
	containersListCmd.Flags().BoolP("all", "a", false, "Show all containers (default shows just running)")
	containersListCmd.MarkFlagRequired("endpoint")

	containersLogsCmd.Flags().Int("endpoint", 0, "Environment endpoint ID (required)")
	containersLogsCmd.Flags().BoolP("follow", "f", false, "Follow log output")
	containersLogsCmd.Flags().IntP("tail", "n", 100, "Number of lines to show from the end")
	containersLogsCmd.MarkFlagRequired("endpoint")

	containersInspectCmd.Flags().Int("endpoint", 0, "Environment endpoint ID (required)")
	containersInspectCmd.MarkFlagRequired("endpoint")

	containersStartCmd.Flags().Int("endpoint", 0, "Environment endpoint ID (required)")
	containersStartCmd.MarkFlagRequired("endpoint")

	containersStopCmd.Flags().Int("endpoint", 0, "Environment endpoint ID (required)")
	containersStopCmd.MarkFlagRequired("endpoint")

	containersRestartCmd.Flags().Int("endpoint", 0, "Environment endpoint ID (required)")
	containersRestartCmd.MarkFlagRequired("endpoint")

	containersRemoveCmd.Flags().Int("endpoint", 0, "Environment endpoint ID (required)")
	containersRemoveCmd.Flags().BoolP("force", "f", false, "Force remove running container")
	containersRemoveCmd.MarkFlagRequired("endpoint")
}
