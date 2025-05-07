package project

import (
	"fmt"
	"github.com/devplaninc/devplan-cli/internal/cmd/project/focus"
	"github.com/spf13/cobra"
)

const (
	nowSection   = "now-projects"
	nextSection  = "next-projects"
	laterSection = "later-projects"
)

var Cmd = create()

func init() {
	Cmd.AddCommand(focus.Cmd)
}

func mainGroupID(companyID int32) string {
	return fmt.Sprintf("%v-projects", companyID)
}

func create() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "project",
		Short:   "Project management commands",
		Aliases: []string{"p"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("please specify a subcommand. Use '%s --help' for more information", cmd.Name())
		},
		Args: cobra.NoArgs,
	}
	cmd.SetHelpFunc(func(c *cobra.Command, args []string) {
		c.Printf("Available Commands:\n")
		for _, sub := range c.Commands() {
			c.Printf("  %s\t\t%s\n", sub.Name(), sub.Short)
		}
	})
	return cmd
}
