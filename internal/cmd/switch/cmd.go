package switch_cmd

import (
	"fmt"
	"github.com/devplaninc/devplan-cli/internal/utils/workspace"
	"os"
	"strings"

	"github.com/devplaninc/devplan-cli/internal/out"
	"github.com/devplaninc/devplan-cli/internal/utils/ide"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
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
	features, err := workspace.ListClonedFeatures()
	check(err)
	if len(features) == 0 {
		fmt.Println(out.Failf("No cloned features found. Use 'devplan clone' to clone a feature first."))
		os.Exit(1)
	}

	var displayItems []string
	for _, f := range features {
		displayItems = append(displayItems, f.GetDisplayName())
	}

	prompt := promptui.Select{
		Label: "Select a feature to switch to",
		Items: displayItems,
	}
	idx, _, err := prompt.Run()
	if err != nil {
		fmt.Println(out.Failf("Feature selection failed: %v", err))
		os.Exit(1)
	}
	selectedFeature := features[idx]

	repoPath := selectedFeature.FullPath

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
	launched, err := ide.LaunchIDE(ideName, repoPath)
	if err != nil {
		fmt.Println(out.Failf("Failed to launch IDE: %v", err))
		os.Exit(1)
	}
	if launched {
		fmt.Println(out.Successf("Successfully opened %s in %s", repoPath, ideName))
	}
}

func getIDENames(ides map[ide.IDE]string) []ide.IDE {
	var names []ide.IDE
	for name := range ides {
		names = append(names, name)
	}
	return names
}

func check(err error) {
	if err != nil {
		fmt.Println(out.Failf("Error: %v", err))
		os.Exit(1)
	}
}
