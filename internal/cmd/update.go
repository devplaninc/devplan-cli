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
		Long: `Update the CLI to the latest available version or to a specific version.

By default, this command will update to the latest available version.
Use the --to flag to update to a specific version.`,
		Run: runUpdate,
		Example: `# Update to the latest version
  devplan update

# Update to a specific version (e.g. 0.1.0)
  devplan update --to 0.1.0
`,
	}

	// Flag to specify a specific version to update to
	toVersion string
)

func init() {
	updateCmd.Flags().StringVar(&toVersion, "to", "", "Update to a specific version")

	rootCmd.AddCommand(updateCmd)
}

func runUpdate(_ *cobra.Command, _ []string) {
	client := &updater.Client{}

	if toVersion != "" {
		fmt.Printf("Updating to version %s...\n", out.H(toVersion))
		if err := client.Update(toVersion); err != nil {
			fmt.Printf("Failed to update: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Successfully updated to version %s\n", out.H(toVersion))
		return
	}

	// Otherwise, check for updates and update to the latest production version
	hasUpdate, latestVersion, err := client.CheckForUpdate()
	if err != nil {
		fmt.Printf("Failed to check for updates: %v\n", err)
		os.Exit(1)
	}

	if !hasUpdate {
		out.Psuccessf("Version is up to date!")
		return
	}

	fmt.Printf("Updating from %s to %s...\n", out.H(version.GetVersion()), out.H(latestVersion))
	if err := client.Update(latestVersion); err != nil {
		fmt.Printf("Failed to update: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully updated to version %s\n", out.H(latestVersion))
}
