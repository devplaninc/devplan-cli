package rules

import (
	"github.com/spf13/cobra"
)

var (
	Cmd = create()
)

func create() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rules",
		Short: "Rules related commands",
	}
	cmd.AddCommand(getCmd)
	return cmd
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
