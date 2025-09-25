package cmd

import (
	"fmt"
	"os"

	"github.com/devplaninc/devplan-cli/internal/cmd/auth"
	"github.com/devplaninc/devplan-cli/internal/cmd/clean"
	"github.com/devplaninc/devplan-cli/internal/cmd/clone"
	"github.com/devplaninc/devplan-cli/internal/cmd/focus"
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
	if err := rootCmd.PersistentFlags().MarkHidden("domain"); err != nil {
		fmt.Printf("Failed to initialize CLI (domain flag): %v\n)", err)
		os.Exit(1)
	}
	rootCmd.AddCommand(auth.Cmd)
	rootCmd.AddCommand(focus.Cmd)
	rootCmd.AddCommand(clone.Cmd)
	rootCmd.AddCommand(switch_cmd.Cmd)
	rootCmd.AddCommand(clean.Cmd)
}
