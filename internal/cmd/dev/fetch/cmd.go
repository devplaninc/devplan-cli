package fetch

import (
	"github.com/spf13/cobra"
)

var (
	Cmd = create()
)

func create() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fetch",
		Short: "Commands for fetching content",
	}
	cmd.AddCommand(projectCmd)
	return cmd
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
