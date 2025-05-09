package cmd

import (
	"fmt"
	"github.com/devplaninc/devplan-cli/internal/devplan"
	"github.com/spf13/cobra"
	"os"
)

var (
	authCmd = &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with Devplan service",
		Long:  `Authenticate with the Devplan service to retrieve and store API key for future communications.`,
		Run:   runAuth,
	}
)

func init() {
	rootCmd.AddCommand(authCmd)
}

func runAuth(_ *cobra.Command, _ []string) {
	if _, err := devplan.VerifyAuth(); err != nil {
		fmt.Printf("Failed to authenticate: %v\n", err)
		os.Exit(1)
		return
	}
}
