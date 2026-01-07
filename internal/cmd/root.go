package cmd

import (
	"fmt"
	"os"

	"github.com/devplaninc/devplan-cli/internal/cmd/auth"
	"github.com/devplaninc/devplan-cli/internal/cmd/clean"
	"github.com/devplaninc/devplan-cli/internal/cmd/clone"
	"github.com/devplaninc/devplan-cli/internal/cmd/dev"
	"github.com/devplaninc/devplan-cli/internal/cmd/focus"
	list_cmd "github.com/devplaninc/devplan-cli/internal/cmd/list"
	"github.com/devplaninc/devplan-cli/internal/cmd/mcp"
	"github.com/devplaninc/devplan-cli/internal/cmd/spec"
	switch_cmd "github.com/devplaninc/devplan-cli/internal/cmd/switch"
	prefs_utils "github.com/devplaninc/devplan-cli/internal/utils/prefs"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "devplan",
		Short: "Official cli for https://devplan.com",
		Long: `Official cli for https://devplan.com.
Integrates Devplan project management with local AI-powered IDEs.`,
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&prefs_utils.Domain, "domain", "", "domain to use (app, beta, local)")
	rootCmd.PersistentFlags().StringVar(&prefs_utils.InstructionFile, "instructions-file", "", "Instructions file to output instructions instead of executing commands directly.")
	rootCmd.PersistentFlags().BoolVarP(&prefs_utils.Verbose, "verbose", "v", false, "verbose output")
	if err := rootCmd.PersistentFlags().MarkHidden("domain"); err != nil {
		fmt.Printf("Failed to initialize CLI (domain flag): %v\n)", err)
		os.Exit(1)
	}
	if err := rootCmd.PersistentFlags().MarkHidden("instructions-file"); err != nil {
		fmt.Printf("Failed to initialize CLI (instructions-file flag): %v\n)", err)
		os.Exit(1)
	}
	rootCmd.AddCommand(auth.Cmd)
	rootCmd.AddCommand(focus.Cmd)
	rootCmd.AddCommand(clone.Cmd)
	rootCmd.AddCommand(switch_cmd.Cmd)
	rootCmd.AddCommand(list_cmd.Cmd)
	rootCmd.AddCommand(clean.Cmd)
	rootCmd.AddCommand(dev.Cmd)
	rootCmd.AddCommand(mcp.Cmd)
	rootCmd.AddCommand(spec.Cmd)
}
