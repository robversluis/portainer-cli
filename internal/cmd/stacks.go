package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/robversluis/portainer-cli/internal/client"
	"github.com/robversluis/portainer-cli/internal/config"
	"github.com/robversluis/portainer-cli/internal/output"
	"github.com/robversluis/portainer-cli/internal/watch"
	"github.com/spf13/cobra"
)

var stacksCmd = &cobra.Command{
	Use:   "stacks",
	Short: "Manage stacks",
	Long:  `Deploy, update, and manage Docker Compose and Kubernetes stacks.`,
}

var stacksListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List stacks",
	Long:    `Display a list of all deployed stacks.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		endpointID, err := cmd.Flags().GetInt("endpoint")
		if err != nil {
			return err
		}
		if endpointID == 0 {
			return fmt.Errorf("--endpoint flag is required")
		}

		watchMode, err := cmd.Flags().GetBool("watch")
		if err != nil {
			return err
		}

		interval, err := cmd.Flags().GetInt("interval")
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

		stackService := client.NewStackService(c)
		format := output.ParseFormat(cmd.Flag("output").Value.String())

		listFunc := func() error {
			stacks, err := stackService.List(endpointID)
			if err != nil {
				return err
			}

			switch format {
			case output.FormatJSON, output.FormatYAML:
				formatter := output.NewFormatter(output.Options{Format: format})
				return formatter.Format(stacks)

			default:
				table := output.NewTableData([]string{"ID", "Name", "Type", "Status"})
				for _, stack := range stacks {
					table.AddRow([]string{
						fmt.Sprintf("%d", stack.Id),
						stack.Name,
						stack.TypeString(),
						stack.StatusString(),
					})
				}
				return output.PrintTable(*table)
			}
		}

		if watchMode {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
			go func() {
				<-sigChan
				cancel()
			}()

			opts := watch.DefaultOptions()
			opts.Interval = time.Duration(interval) * time.Second

			fmt.Println("Watching stacks... (Press Ctrl+C to exit)")
			return watch.Watch(ctx, opts, listFunc)
		}

		return listFunc()
	},
}

var stacksDeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy a stack",
	Long:  `Deploy a new stack from a Docker Compose file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		endpointID, err := cmd.Flags().GetInt("endpoint")
		if err != nil {
			return err
		}
		if endpointID == 0 {
			return fmt.Errorf("--endpoint flag is required")
		}

		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}
		if name == "" {
			return fmt.Errorf("--name flag is required")
		}

		filePath, err := cmd.Flags().GetString("file")
		if err != nil {
			return err
		}
		if filePath == "" {
			return fmt.Errorf("--file flag is required")
		}

		envVars, err := cmd.Flags().GetStringArray("env")
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

		var env []client.StackEnv
		for _, e := range envVars {
			parts := strings.SplitN(e, "=", 2)
			if len(parts) == 2 {
				env = append(env, client.StackEnv{
					Name:  parts[0],
					Value: parts[1],
				})
			}
		}

		stackService := client.NewStackService(c)
		stack, err := stackService.DeployFromFile(endpointID, name, filePath, env)
		if err != nil {
			return err
		}

		if !GetQuiet() {
			fmt.Printf("Stack '%s' deployed successfully (ID: %d)\n", stack.Name, stack.Id)
		}

		return nil
	},
}

var stacksGetCmd = &cobra.Command{
	Use:   "get [id or name]",
	Short: "Get stack details",
	Long:  `Retrieve detailed information about a specific stack.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		endpointID, err := cmd.Flags().GetInt("endpoint")
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

		stackService := client.NewStackService(c)
		
		var stack *client.Stack
		var stackID int
		if _, err := fmt.Sscanf(args[0], "%d", &stackID); err == nil {
			stack, err = stackService.Get(stackID)
			if err != nil {
				return err
			}
		} else {
			if endpointID == 0 {
				return fmt.Errorf("--endpoint flag is required when using stack name")
			}
			stack, err = stackService.GetByName(endpointID, args[0])
			if err != nil {
				return err
			}
		}

		format := output.ParseFormat(cmd.Flag("output").Value.String())

		switch format {
		case output.FormatJSON, output.FormatYAML:
			formatter := output.NewFormatter(output.Options{Format: format})
			return formatter.Format(stack)

		default:
			fmt.Printf("ID:          %d\n", stack.Id)
			fmt.Printf("Name:        %s\n", stack.Name)
			fmt.Printf("Type:        %s\n", stack.TypeString())
			fmt.Printf("Status:      %s\n", stack.StatusString())
			fmt.Printf("Endpoint ID: %d\n", stack.EndpointId)
			
			if stack.EntryPoint != "" {
				fmt.Printf("Entry Point: %s\n", stack.EntryPoint)
			}
			
			if len(stack.Env) > 0 {
				fmt.Printf("\nEnvironment Variables:\n")
				for _, env := range stack.Env {
					fmt.Printf("  %s=%s\n", env.Name, env.Value)
				}
			}

			return nil
		}
	},
}

var stacksRemoveCmd = &cobra.Command{
	Use:     "remove [id or name]",
	Aliases: []string{"rm"},
	Short:   "Remove a stack",
	Long:    `Remove a deployed stack.`,
	Args:    cobra.ExactArgs(1),
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

		stackService := client.NewStackService(c)
		
		var stackID int
		if _, err := fmt.Sscanf(args[0], "%d", &stackID); err == nil {
			if err := stackService.Remove(stackID, endpointID); err != nil {
				return err
			}
		} else {
			stack, err := stackService.GetByName(endpointID, args[0])
			if err != nil {
				return err
			}
			if err := stackService.Remove(stack.Id, endpointID); err != nil {
				return err
			}
		}

		if !GetQuiet() {
			fmt.Printf("Stack removed successfully\n")
		}

		return nil
	},
}

var stacksUpdateCmd = &cobra.Command{
	Use:   "update [stack-id]",
	Short: "Update a stack",
	Long:  `Update an existing stack with a new compose file.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var stackID int
		if _, err := fmt.Sscanf(args[0], "%d", &stackID); err != nil {
			return fmt.Errorf("invalid stack ID: %s", args[0])
		}

		endpointID, err := cmd.Flags().GetInt("endpoint")
		if err != nil {
			return err
		}
		if endpointID == 0 {
			return fmt.Errorf("--endpoint flag is required")
		}

		stackFile, err := cmd.Flags().GetString("file")
		if err != nil {
			return err
		}
		if stackFile == "" {
			return fmt.Errorf("--file flag is required")
		}

		envVars, err := cmd.Flags().GetStringArray("env")
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

		stackService := client.NewStackService(c)
		
		content, err := client.ParseStackFile(stackFile)
		if err != nil {
			return err
		}

		var env []client.StackEnv
		for _, envVar := range envVars {
			parts := strings.SplitN(envVar, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid env format: %s (expected KEY=VALUE)", envVar)
			}
			env = append(env, client.StackEnv{
				Name:  parts[0],
				Value: parts[1],
			})
		}

		if err := stackService.Update(stackID, endpointID, content, env); err != nil {
			return err
		}

		if !GetQuiet() {
			fmt.Printf("Stack %d updated successfully\n", stackID)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(stacksCmd)
	stacksCmd.AddCommand(stacksListCmd)
	stacksCmd.AddCommand(stacksDeployCmd)
	stacksCmd.AddCommand(stacksGetCmd)
	stacksCmd.AddCommand(stacksUpdateCmd)
	stacksCmd.AddCommand(stacksRemoveCmd)

	stacksListCmd.Flags().Int("endpoint", 0, "Environment endpoint ID (required)")
	stacksListCmd.Flags().BoolP("watch", "w", false, "Watch for changes and continuously update")
	stacksListCmd.Flags().Int("interval", 2, "Refresh interval in seconds for watch mode")
	_ = stacksListCmd.MarkFlagRequired("endpoint")

	stacksDeployCmd.Flags().String("file", "", "Path to stack file (required)")
	stacksDeployCmd.Flags().String("name", "", "Stack name (required)")
	stacksDeployCmd.Flags().Int("endpoint", 0, "Environment endpoint ID (required)")
	stacksDeployCmd.Flags().StringArray("env", []string{}, "Environment variables (KEY=VALUE)")
	_ = stacksDeployCmd.MarkFlagRequired("file")
	_ = stacksDeployCmd.MarkFlagRequired("name")
	_ = stacksDeployCmd.MarkFlagRequired("endpoint")

	stacksGetCmd.Flags().Int("endpoint", 0, "Environment endpoint ID (required for name lookup)")

	stacksRemoveCmd.Flags().Int("endpoint", 0, "Environment endpoint ID (required)")
	_ = stacksRemoveCmd.MarkFlagRequired("endpoint")

	stacksUpdateCmd.Flags().Int("endpoint", 0, "Environment endpoint ID (required)")
	stacksUpdateCmd.Flags().String("file", "", "Path to stack file (required)")
	stacksUpdateCmd.Flags().StringArray("env", []string{}, "Environment variables (KEY=VALUE)")
	_ = stacksUpdateCmd.MarkFlagRequired("endpoint")
	_ = stacksUpdateCmd.MarkFlagRequired("file")
}
