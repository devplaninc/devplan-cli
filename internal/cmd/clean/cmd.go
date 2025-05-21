package clean

import (
	"fmt"
	"github.com/devplaninc/devplan-cli/internal/out"
	"github.com/devplaninc/devplan-cli/internal/utils/workspace"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"os"
)

var (
	Cmd = create()
)

func create() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Allows to clean up individual repositories from the workspace",
		Long:  "Lists all cloned repositories in the workspace and allows to delete them from the local machine.",
		Run: func(_ *cobra.Command, _ []string) {
			runClean()
		},
	}
	return cmd
}

func runClean() {
	clonedFeatures, err := workspace.ListClonedFeatures()
	check(err)
	if len(clonedFeatures) == 0 {
		out.Psuccessf("Nothing to clean!")
		return
	}

	var displayItems []string
	for _, f := range clonedFeatures {
		displayItems = append(displayItems, f.GetDisplayName())
	}

	prompt := promptui.Select{
		Label: "Select a feature to clean up",
		Items: displayItems,
	}
	idx, _, err := prompt.Run()
	check(err)
	selectedFeature := clonedFeatures[idx].DirName

	featurePath := workspace.GetFeaturePath(selectedFeature)

	confirm := promptui.Prompt{
		Label:     fmt.Sprintf("Directory %s will be permanently deleted. Are you sure", featurePath),
		IsConfirm: true,
	}
	resp, err := confirm.Run()
	check(err)
	if resp != "y" {
		out.Pfailf("Deletion aborted")
		return
	}

	err = os.RemoveAll(featurePath)
	check(err)
	out.Psuccessf("Successfully deleted %s\n", out.H(featurePath))
}

func check(err error) {
	if err != nil {
		fmt.Println(out.Failf("Error: %v", err))
		os.Exit(1)
	}
}
