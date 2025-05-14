package auth

import (
	"fmt"
	"github.com/devplaninc/devplan-cli/internal/devplan"
	"github.com/spf13/cobra"
	"os"
)

var Cmd = create()

func create() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with Devplan service",
		Long:  `Authenticate with the Devplan service to retrieve and store API key for future communications.`,
		Run:   func(_ *cobra.Command, _ []string) { runAuth(force) },
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force reauthentication even if token exists")
	return cmd
}

func runAuth(force bool) {
	var err error
	if force {
		_, err = devplan.RequestAuth()
	} else {
		_, err = devplan.VerifyAuth()
	}
	if err != nil {
		fmt.Printf("Failed to authenticate: %v\n", err)
		os.Exit(1)
		return
	}
}
