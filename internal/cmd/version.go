package cmd

import (
	"fmt"
	"github.com/devplaninc/devplan-cli/internal/version"
	"github.com/spf13/cobra"
)

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Long:  "Print the version information of the CLI",
		Run:   runVersion,
	}
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

func runVersion(_ *cobra.Command, _ []string) {
	fmt.Println(version.GetVersionInfo())
}
