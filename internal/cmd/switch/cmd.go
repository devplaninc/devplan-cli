package switch_cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/devplaninc/devplan-cli/internal/out"
	"github.com/devplaninc/devplan-cli/internal/utils/ide"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	Cmd = create()
)

func create() *cobra.Command {
	var ideName string
	cmd := &cobra.Command{
		Use:     "switch",
		Hidden:  true,
		Aliases: []string{"sw"},
		Short:   "List and switch between cloned features",
		Long:    `List all cloned features in the workspace and switch to one of them by opening it in your preferred IDE.`,
		Run: func(_ *cobra.Command, _ []string) {
			runSwitch(ideName)
		},
	}
	cmd.Flags().StringVarP(&ideName, "ide", "i", "", "IDE to use (e.g., vscode, intellij, cursor)")
	return cmd
}

func runSwitch(ideName string) {
	// Get workspace directory
	workspaceDir := getWorkspaceDir()
	featuresDir := filepath.Join(workspaceDir, "features")

	// Check if features directory exists
	if _, err := os.Stat(featuresDir); os.IsNotExist(err) {
		fmt.Println(out.Failf("No cloned features found. Use 'devplan clone' to clone a feature first."))
		os.Exit(1)
	}

	// Get list of cloned features
	features, err := listClonedFeatures(featuresDir)
	if err != nil {
		fmt.Println(out.Failf("Failed to list cloned features: %v", err))
		os.Exit(1)
	}

	if len(features) == 0 {
		fmt.Println(out.Failf("No cloned features found. Use 'devplan clone' to clone a feature first."))
		os.Exit(1)
	}

	// Prompt user to select a feature
	prompt := promptui.Select{
		Label: "Select a feature to switch to",
		Items: features,
	}
	idx, _, err := prompt.Run()
	if err != nil {
		fmt.Println(out.Failf("Feature selection failed: %v", err))
		os.Exit(1)
	}
	selectedFeature := features[idx]

	// Get the repository path
	repoPath := filepath.Join(featuresDir, selectedFeature)

	// If IDE name is provided, use it
	if ideName != "" {
		ideV := ide.IDE(strings.ToLower(ideName))
		launchSelectedIDE(ideV, repoPath)
		return
	}

	ides, err := ide.DetectInstalledIDEs()
	if err != nil {
		fmt.Println(out.Failf("Failed to detect installed IDEs: %v", err))
		os.Exit(1)
	}

	if len(ides) == 0 {
		fmt.Println(out.Failf("No supported IDEs found on your system."))
		os.Exit(1)
	}

	// Otherwise, prompt user to select an IDE
	ideNames := getIDENames(ides)
	idePrompt := promptui.Select{
		Label: "Select an IDE to open the feature",
		Items: ideNames,
	}
	ideIdx, _, err := idePrompt.Run()
	if err != nil {
		fmt.Println(out.Failf("IDE selection failed: %v", err))
		os.Exit(1)
	}
	selectedIDEName := ideNames[ideIdx]

	launchSelectedIDE(selectedIDEName, repoPath)
}

func launchSelectedIDE(ideName ide.IDE, repoPath string) {
	fmt.Printf("Opening %s in %s...\n", out.H(repoPath), out.Hf("%v", ideName))
	err := ide.LaunchIDE(ideName, repoPath)
	if err != nil {
		fmt.Println(out.Failf("Failed to launch IDE: %v", err))
		os.Exit(1)
	}
	fmt.Println(out.Successf("Successfully opened %s in %s", repoPath, ideName))
}

func getIDENames(ides map[ide.IDE]string) []ide.IDE {
	var names []ide.IDE
	for name := range ides {
		names = append(names, name)
	}
	return names
}

func listClonedFeatures(featuresDir string) ([]string, error) {
	entries, err := os.ReadDir(featuresDir)
	if err != nil {
		return nil, err
	}

	var features []string
	for _, entry := range entries {
		if entry.IsDir() {
			features = append(features, entry.Name())
		}
	}
	return features, nil
}

func getWorkspaceDir() string {
	workspaceDir := viper.GetString("workspace_dir")
	if workspaceDir == "" {
		// Use default directory in user's home
		home, err := os.UserHomeDir()
		if err != nil {
			out.Pfailf("Failed to get user home directory: %v\n", err)
			os.Exit(1)
		}
		workspaceDir = filepath.Join(home, "devplan", "workspace")

		// Save to config for future use
		viper.Set("workspace_dir", workspaceDir)
		err = viper.WriteConfig()
		if err != nil {
			out.Pfailf("Failed to write config: %v\n", err)
			// Continue anyway, just log the error
		}
	}

	return workspaceDir
}
