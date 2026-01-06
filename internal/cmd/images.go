package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/robversluis/portainer-cli/internal/client"
	"github.com/robversluis/portainer-cli/internal/config"
	"github.com/robversluis/portainer-cli/internal/output"
	"github.com/robversluis/portainer-cli/internal/watch"
	"github.com/spf13/cobra"
)

var imagesCmd = &cobra.Command{
	Use:   "images",
	Short: "Manage Docker images",
	Long:  `List, pull, push, and manage Docker images across environments.`,
}

var imagesListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List images",
	Long:    `Display a list of Docker images in the specified environment.`,
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

		c, err := client.NewClient(profile, GetClientOptions()...)
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		imageService := client.NewImageService(c)
		format := output.ParseFormat(cmd.Flag("output").Value.String())

		listFunc := func() error {
			images, err := imageService.List(endpointID)
			if err != nil {
				return err
			}

			switch format {
			case output.FormatJSON, output.FormatYAML:
				formatter := output.NewFormatter(output.Options{Format: format})
				return formatter.Format(images)

			default:
				table := output.NewTableData([]string{"ID", "Repository", "Tag", "Size", "Created"})
				for _, image := range images {
					createdTime := time.Unix(image.Created, 0)
					table.AddRow([]string{
						image.GetShortID(),
						image.GetRepository(),
						image.GetTag(),
						output.FormatSize(image.Size),
						output.FormatDuration(int64(time.Since(createdTime).Seconds())),
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

			fmt.Println("Watching images... (Press Ctrl+C to exit)")
			return watch.Watch(ctx, opts, listFunc)
		}

		return listFunc()
	},
}

var imagesInspectCmd = &cobra.Command{
	Use:   "inspect [image]",
	Short: "Inspect an image",
	Long:  `Display detailed information about a specific image.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		endpointID, err := cmd.Flags().GetInt("endpoint")
		if err != nil {
			return err
		}
		if endpointID == 0 {
			return fmt.Errorf("--endpoint flag is required")
		}

		imageID := args[0]

		profile, err := config.GetProfileFromViper()
		if err != nil {
			return fmt.Errorf("failed to get profile: %w", err)
		}

		c, err := client.NewClient(profile, GetClientOptions()...)
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		imageService := client.NewImageService(c)
		image, err := imageService.Inspect(endpointID, imageID)
		if err != nil {
			return err
		}

		format := output.ParseFormat(cmd.Flag("output").Value.String())

		switch format {
		case output.FormatJSON, output.FormatYAML:
			formatter := output.NewFormatter(output.Options{Format: format})
			return formatter.Format(image)

		default:
			fmt.Printf("ID:           %s\n", image.Id)
			if len(image.RepoTags) > 0 {
				fmt.Printf("Tags:         %s\n", output.FormatList(image.RepoTags, ", "))
			}
			if len(image.RepoDigests) > 0 {
				fmt.Printf("Digests:      %s\n", output.FormatList(image.RepoDigests, ", "))
			}
			fmt.Printf("Created:      %s\n", image.Created)
			fmt.Printf("Size:         %s\n", output.FormatSize(image.Size))
			fmt.Printf("Virtual Size: %s\n", output.FormatSize(image.VirtualSize))
			fmt.Printf("Architecture: %s\n", image.Architecture)
			fmt.Printf("OS:           %s\n", image.Os)

			if image.Author != "" {
				fmt.Printf("Author:       %s\n", image.Author)
			}
			if image.DockerVersion != "" {
				fmt.Printf("Docker:       %s\n", image.DockerVersion)
			}

			if image.Config != nil && len(image.Config.Env) > 0 {
				fmt.Printf("\nEnvironment:\n")
				for _, env := range image.Config.Env {
					fmt.Printf("  %s\n", env)
				}
			}

			if len(image.RootFS.Layers) > 0 {
				fmt.Printf("\nLayers (%d):\n", len(image.RootFS.Layers))
				for i, layer := range image.RootFS.Layers {
					if i < 5 {
						fmt.Printf("  %s\n", layer)
					} else if i == 5 {
						fmt.Printf("  ... and %d more\n", len(image.RootFS.Layers)-5)
						break
					}
				}
			}

			return nil
		}
	},
}

var imagesPullCmd = &cobra.Command{
	Use:   "pull [image]",
	Short: "Pull an image",
	Long:  `Pull a Docker image from a registry.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		endpointID, err := cmd.Flags().GetInt("endpoint")
		if err != nil {
			return err
		}
		if endpointID == 0 {
			return fmt.Errorf("--endpoint flag is required")
		}

		imageName := args[0]
		registryID, err := cmd.Flags().GetInt("registry")
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

		imageService := client.NewImageService(c)
		if err := imageService.Pull(endpointID, imageName, registryID); err != nil {
			return err
		}

		if !GetQuiet() {
			fmt.Printf("Image '%s' pulled successfully\n", imageName)
		}

		return nil
	},
}

var imagesRemoveCmd = &cobra.Command{
	Use:     "remove [image]",
	Aliases: []string{"rm"},
	Short:   "Remove an image",
	Long:    `Remove a Docker image.`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		endpointID, err := cmd.Flags().GetInt("endpoint")
		if err != nil {
			return err
		}
		if endpointID == 0 {
			return fmt.Errorf("--endpoint flag is required")
		}

		imageID := args[0]
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

		imageService := client.NewImageService(c)
		if err := imageService.Remove(endpointID, imageID, force); err != nil {
			return err
		}

		if !GetQuiet() {
			fmt.Printf("Image '%s' removed successfully\n", imageID)
		}

		return nil
	},
}

var imagesPruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove unused images",
	Long:  `Remove all dangling or unused images.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		endpointID, err := cmd.Flags().GetInt("endpoint")
		if err != nil {
			return err
		}
		if endpointID == 0 {
			return fmt.Errorf("--endpoint flag is required")
		}

		dangling, err := cmd.Flags().GetBool("dangling")
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

		imageService := client.NewImageService(c)
		if err := imageService.Prune(endpointID, dangling); err != nil {
			return err
		}

		if !GetQuiet() {
			fmt.Println("Images pruned successfully")
		}

		return nil
	},
}

var imagesTagCmd = &cobra.Command{
	Use:   "tag [source] [target]",
	Short: "Tag an image",
	Long:  `Create a tag TARGET_IMAGE that refers to SOURCE_IMAGE.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		endpointID, err := cmd.Flags().GetInt("endpoint")
		if err != nil {
			return err
		}
		if endpointID == 0 {
			return fmt.Errorf("--endpoint flag is required")
		}

		sourceImage := args[0]
		targetImage := args[1]

		profile, err := config.GetProfileFromViper()
		if err != nil {
			return fmt.Errorf("failed to get profile: %w", err)
		}

		c, err := client.NewClient(profile, GetClientOptions()...)
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		parts := splitImageName(targetImage)
		imageService := client.NewImageService(c)
		if err := imageService.Tag(endpointID, sourceImage, parts[0], parts[1]); err != nil {
			return err
		}

		if !GetQuiet() {
			fmt.Printf("Tagged '%s' as '%s'\n", sourceImage, targetImage)
		}

		return nil
	},
}

func splitImageName(imageName string) [2]string {
	parts := [2]string{"", "latest"}

	lastColon := -1
	for i := len(imageName) - 1; i >= 0; i-- {
		if imageName[i] == ':' {
			lastColon = i
			break
		}
		if imageName[i] == '/' {
			break
		}
	}

	if lastColon > 0 {
		parts[0] = imageName[:lastColon]
		parts[1] = imageName[lastColon+1:]
	} else {
		parts[0] = imageName
	}

	return parts
}

func init() {
	rootCmd.AddCommand(imagesCmd)
	imagesCmd.AddCommand(imagesListCmd)
	imagesCmd.AddCommand(imagesInspectCmd)
	imagesCmd.AddCommand(imagesPullCmd)
	imagesCmd.AddCommand(imagesRemoveCmd)
	imagesCmd.AddCommand(imagesPruneCmd)
	imagesCmd.AddCommand(imagesTagCmd)

	imagesListCmd.Flags().Int("endpoint", 0, "Environment endpoint ID (required)")
	imagesListCmd.Flags().BoolP("watch", "w", false, "Watch for changes and continuously update")
	imagesListCmd.Flags().Int("interval", 2, "Refresh interval in seconds for watch mode")
	_ = imagesListCmd.MarkFlagRequired("endpoint")

	imagesInspectCmd.Flags().Int("endpoint", 0, "Environment endpoint ID (required)")
	_ = imagesInspectCmd.MarkFlagRequired("endpoint")

	imagesPullCmd.Flags().Int("endpoint", 0, "Environment endpoint ID (required)")
	imagesPullCmd.Flags().Int("registry", 0, "Registry ID for authentication")
	_ = imagesPullCmd.MarkFlagRequired("endpoint")

	imagesRemoveCmd.Flags().Int("endpoint", 0, "Environment endpoint ID (required)")
	imagesRemoveCmd.Flags().BoolP("force", "f", false, "Force removal of the image")
	_ = imagesRemoveCmd.MarkFlagRequired("endpoint")

	imagesPruneCmd.Flags().Int("endpoint", 0, "Environment endpoint ID (required)")
	imagesPruneCmd.Flags().Bool("dangling", true, "Remove only dangling images")
	_ = imagesPruneCmd.MarkFlagRequired("endpoint")

	imagesTagCmd.Flags().Int("endpoint", 0, "Environment endpoint ID (required)")
	_ = imagesTagCmd.MarkFlagRequired("endpoint")
}
