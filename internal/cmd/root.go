package cmd

import (
	"fmt"
	"github.com/devplaninc/devplan-cli/internal/cmd/focus"
	"github.com/devplaninc/devplan-cli/internal/utils/globals"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&globals.Domain, "domain", "", "domain to use (app, beta, local)")
	if err := rootCmd.PersistentFlags().MarkHidden("domain"); err != nil {
		fmt.Printf("Failed to initialize CLI (domain flag): %v\n)", err)
		os.Exit(1)
	}
	rootCmd.AddCommand(focus.Cmd)
}

func initConfig() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	configDir := filepath.Join(home, ".devplan")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		_ = os.MkdirAll(configDir, 0755)
	}

	viper.AddConfigPath(configDir)
	viper.SetConfigType("json")
	viper.SetConfigName("config")
	viper.SetEnvPrefix("devplan")

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		slog.Debug("Using config file: %s", viper.ConfigFileUsed())
	}
}
