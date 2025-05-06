package cmd

import (
	"fmt"
	"github.com/devplaninc/devplan-cli/internal/devplan"
	"github.com/devplaninc/devplan-cli/internal/out"
	"github.com/spf13/cobra"
)

var (
	selfCmd = &cobra.Command{
		Use:   "self",
		Short: "Print information about the current user",
		Long:  "Print information about the current user",
		Run:   runSelf,
	}
)

func init() {
	rootCmd.AddCommand(selfCmd)
}

func runSelf(_ *cobra.Command, _ []string) {
	client := devplan.NewClient(devplan.ClientConfig{})
	self, err := client.GetSelf()
	if err != nil {
		fmt.Printf("Failed to get self: %v\n", err)
		return
	}
	fmt.Printf("Self: %s\n", out.Highlight(self.GetOwnInfo().GetUser().GetEmail()))
}
