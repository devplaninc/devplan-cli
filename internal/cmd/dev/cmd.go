package dev

import (
	"github.com/devplaninc/devplan-cli/internal/cmd/dev/fetch"
	"github.com/devplaninc/devplan-cli/internal/cmd/dev/rules"
	"github.com/spf13/cobra"
)

var (
	Cmd = create()
)

func create() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "dev",
		Short:  "Developer utilities",
		Hidden: true,
	}
	cmd.AddCommand(rules.Cmd)
	cmd.AddCommand(fetch.Cmd)
	return cmd
}
