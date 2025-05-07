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
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Long:  "Print the version information of the CLI",
		Run:   runVersion,
	}
	// Flag to print latest available version
	showLatest bool
)

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolVarP(&showLatest, "latest", "l", false, "Print latest available version")
}

func runVersion(_ *cobra.Command, _ []string) {
	if showLatest {
		client := &updater.Client{}
		ver, err := client.GetVersionConfig()
		if err != nil {
			out.Failf("Failed to lget latest version: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Latest available version: %s\n", out.H(ver.GetProductionVersion()))
		return
	}
	fmt.Printf(`Version: %v
Commit: %v
Build Date: %v
`, out.H(version.GetVersion()), out.H(version.GetCommitHash()), out.H(version.GetBuildDate()))
}
