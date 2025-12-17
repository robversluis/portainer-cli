package cmd

import (
	"fmt"
	"strconv"

	"github.com/robversluis/portainer-cli/internal/client"
	"github.com/robversluis/portainer-cli/internal/config"
	"github.com/robversluis/portainer-cli/internal/output"
	"github.com/spf13/cobra"
)

var environmentsCmd = &cobra.Command{
	Use:     "environments",
	Aliases: []string{"env"},
	Short:   "Manage Portainer environments",
	Long:    `List, create, update, and delete Portainer environments (endpoints).`,
}

var environmentsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all environments",
	Long:    `Display a list of all Portainer environments with their status and details.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		profile, err := config.GetProfileFromViper()
		if err != nil {
			return fmt.Errorf("failed to get profile: %w", err)
		}

		c, err := client.NewClient(profile, client.WithVerbose(GetVerbose()))
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		envService := client.NewEnvironmentService(c)
		environments, err := envService.List()
		if err != nil {
			return err
		}

		format := output.ParseFormat(cmd.Flag("output").Value.String())

		switch format {
		case output.FormatJSON, output.FormatYAML:
			formatter := output.NewFormatter(output.Options{Format: format})
			return formatter.Format(environments)

		default:
			table := output.NewTableData([]string{"ID", "Name", "Type", "URL", "Status"})
			for _, env := range environments {
				url := env.URL
				if len(url) > 40 {
					url = output.TruncateString(url, 40)
				}
				table.AddRow([]string{
					fmt.Sprintf("%d", env.Id),
					env.Name,
					env.TypeString(),
					url,
					env.StatusString(),
				})
			}
			return output.PrintTable(*table)
		}
	},
}

var environmentsGetCmd = &cobra.Command{
	Use:   "get [id or name]",
	Short: "Get environment details",
	Long:  `Retrieve detailed information about a specific environment by ID or name.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profile, err := config.GetProfileFromViper()
		if err != nil {
			return fmt.Errorf("failed to get profile: %w", err)
		}

		c, err := client.NewClient(profile, client.WithVerbose(GetVerbose()))
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		envService := client.NewEnvironmentService(c)
		
		var env *client.Environment
		
		if id, err := strconv.Atoi(args[0]); err == nil {
			env, err = envService.Get(id)
			if err != nil {
				return err
			}
		} else {
			env, err = envService.GetByName(args[0])
			if err != nil {
				return err
			}
		}

		format := output.ParseFormat(cmd.Flag("output").Value.String())

		switch format {
		case output.FormatJSON, output.FormatYAML:
			formatter := output.NewFormatter(output.Options{Format: format})
			return formatter.Format(env)

		default:
			fmt.Printf("ID:          %d\n", env.Id)
			fmt.Printf("Name:        %s\n", env.Name)
			fmt.Printf("Type:        %s\n", env.TypeString())
			fmt.Printf("URL:         %s\n", env.URL)
			if env.PublicURL != "" {
				fmt.Printf("Public URL:  %s\n", env.PublicURL)
			}
			fmt.Printf("Status:      %s\n", env.StatusString())
			fmt.Printf("Group ID:    %d\n", env.GroupId)

			if env.EdgeID != "" {
				fmt.Printf("\nEdge Configuration:\n")
				fmt.Printf("  Edge ID:   %s\n", env.EdgeID)
				if env.EdgeCheckinInterval > 0 {
					fmt.Printf("  Checkin:   %ds\n", env.EdgeCheckinInterval)
				}
			}

			if env.Agent.Version != "" {
				fmt.Printf("\nAgent:\n")
				fmt.Printf("  Version:   %s\n", env.Agent.Version)
			}

			snapshot := env.GetLatestSnapshot()
			if snapshot != nil {
				fmt.Printf("\nSnapshot:\n")
				fmt.Printf("  Containers:  %d running, %d stopped\n", 
					snapshot.RunningContainerCount, snapshot.StoppedContainerCount)
				if snapshot.HealthyContainerCount > 0 || snapshot.UnhealthyContainerCount > 0 {
					fmt.Printf("  Health:      %d healthy, %d unhealthy\n",
						snapshot.HealthyContainerCount, snapshot.UnhealthyContainerCount)
				}
				fmt.Printf("  Images:      %d\n", snapshot.ImageCount)
				fmt.Printf("  Volumes:     %d\n", snapshot.VolumeCount)
				if snapshot.ServiceCount > 0 {
					fmt.Printf("  Services:    %d\n", snapshot.ServiceCount)
				}
				if snapshot.StackCount > 0 {
					fmt.Printf("  Stacks:      %d\n", snapshot.StackCount)
				}
				if snapshot.TotalCPU > 0 {
					fmt.Printf("  CPU:         %d\n", snapshot.TotalCPU)
				}
				if snapshot.TotalMemory > 0 {
					fmt.Printf("  Memory:      %s\n", output.FormatSize(snapshot.TotalMemory))
				}
			}

			if len(env.TagIds) > 0 {
				fmt.Printf("\nTags:        %v\n", env.TagIds)
			}

			return nil
		}
	},
}

var environmentsInspectCmd = &cobra.Command{
	Use:   "inspect [id or name]",
	Short: "Inspect environment (alias for get)",
	Long:  `Inspect detailed information about a specific environment.`,
	Args:  cobra.ExactArgs(1),
	RunE:  environmentsGetCmd.RunE,
}

func init() {
	rootCmd.AddCommand(environmentsCmd)
	environmentsCmd.AddCommand(environmentsListCmd)
	environmentsCmd.AddCommand(environmentsGetCmd)
	environmentsCmd.AddCommand(environmentsInspectCmd)
}
