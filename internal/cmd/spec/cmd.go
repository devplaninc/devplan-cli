package spec

import (
	"fmt"
	"os"

	"github.com/devplaninc/devplan-cli/internal/out"
	"github.com/spf13/cobra"
)

var (
	Cmd = create()
)

func create() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "spec",
		Short:   "Specifications related commands",
		Aliases: []string{"specs"},
	}
	cmd.AddCommand(ideCmd)
	cmd.AddCommand(taskCmd)
	cmd.AddCommand(startCmd)
	return cmd
}

func check(err error) {
	if err != nil {
		fmt.Println(out.Failf("%v", err))
		os.Exit(1)
	}
}
