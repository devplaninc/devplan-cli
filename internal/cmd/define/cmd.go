package define

import (
	"github.com/spf13/cobra"
)

var Cmd = create()

func create() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "define",
		Short: "Define feature commands",
	}
	cmd.AddCommand(createFeatureCmd())
	cmd.AddCommand(createStatusCmd())
	return cmd
}
