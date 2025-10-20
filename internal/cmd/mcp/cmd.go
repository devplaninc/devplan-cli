package mcp

import (
	"fmt"
	"os"

	"github.com/devplaninc/devplan-cli/internal/mcp"
	"github.com/devplaninc/devplan-cli/internal/out"
	"github.com/spf13/cobra"
)

var (
	Cmd = create()
)

func create() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Devplan MCP server",
		Run: func(_ *cobra.Command, _ []string) {
			check(mcp.RunServer())
		},
	}
	return cmd
}

func check(err error) {
	if err != nil {
		fmt.Println(out.Failf("Error: %v", err))
		os.Exit(1)
	}
}
