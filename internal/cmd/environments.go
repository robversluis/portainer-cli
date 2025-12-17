package cmd

import (
	"fmt"

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
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Environments list command - to be implemented in Task 6")
	},
}

var environmentsGetCmd = &cobra.Command{
	Use:   "get [id]",
	Short: "Get environment details",
	Long:  `Retrieve detailed information about a specific environment.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Environments get command for ID: %s - to be implemented in Task 6\n", args[0])
	},
}

func init() {
	rootCmd.AddCommand(environmentsCmd)
	environmentsCmd.AddCommand(environmentsListCmd)
	environmentsCmd.AddCommand(environmentsGetCmd)
}
