package cmd

import (
	"fmt"
	"github.com/devplaninc/devplan-cli/internal/out"
	"github.com/devplaninc/devplan-cli/internal/utils/updater"
	"github.com/devplaninc/devplan-cli/internal/version"
	"github.com/spf13/cobra"
	"os"
)

var (
	updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update the CLI to the latest version",
		Long: `Update the CLI to the latest production version or to a specific version.

By default, this command will update to the latest production version.
Use the --to flag to update to a specific version.`,
		Run: runUpdate,
	}

	// Flag to specify a specific version to update to
	toVersion string

	// Flag to list all available versions
	listVersions bool
)

func init() {
	updateCmd.Flags().StringVar(&toVersion, "to", "", "Update to a specific version")
	updateCmd.Flags().BoolVar(&listVersions, "list", false, "List all available versions")
	rootCmd.AddCommand(updateCmd)
}

func runUpdate(_ *cobra.Command, _ []string) {
	client := &updater.Client{}

	if listVersions {
		versions, err := client.ListAvailableVersions()
		if err != nil {
			fmt.Printf("Failed to list available versions: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Available versions:")
		for _, v := range versions {
			fmt.Printf("  %s\n", v)
		}
		return
	}

	if toVersion != "" {
		fmt.Printf("Updating to version %s...\n", out.Highlight(toVersion))
		if err := client.Update(toVersion); err != nil {
			fmt.Printf("Failed to update: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Successfully updated to version %s\n", out.Highlight(toVersion))
		return
	}

	// Otherwise, check for updates and update to the latest production version
	hasUpdate, latestVersion, err := client.CheckForUpdate()
	if err != nil {
		fmt.Printf("Failed to check for updates: %v\n", err)
		os.Exit(1)
	}

	if !hasUpdate {
		fmt.Printf("You are already using the latest version: %s\n", out.Highlight(version.GetVersion()))
		return
	}

	fmt.Printf("Updating from %s to %s...\n", out.Highlight(version.GetVersion()), out.Highlight(latestVersion))
	if err := client.Update(latestVersion); err != nil {
		fmt.Printf("Failed to update: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully updated to version %s\n", out.Highlight(latestVersion))
}
