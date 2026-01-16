package list_cmd

import (
	"fmt"
	"os"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/huh"
	"github.com/devplaninc/devplan-cli/internal/cmd/common"
	"github.com/devplaninc/devplan-cli/internal/out"
	"github.com/devplaninc/devplan-cli/internal/utils/recentactivity"
	"github.com/devplaninc/devplan-cli/internal/utils/workspace"
	"github.com/spf13/cobra"
)

var (
	Cmd = create()
)

func create() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List cloned features",
		Long:    `List all cloned features in the workspace`,
		Run: func(_ *cobra.Command, _ []string) {
			runList()
		},
	}
	return cmd
}

func runList() {
	features, err := workspace.ListClonedRepos()
	check(err)

	if len(features) == 0 {
		fmt.Println(out.Failf("No cloned features found. Use 'devplan clone' to clone a feature first."))
		os.Exit(0)
	}

	features = recentactivity.SortClonedFeatures(features)

	options, hasAnyChanges := common.BuildFeatureOptions(features)

	// Show legend if there are any uncommitted changes
	if hasAnyChanges {
		common.ShowLegend()
	}

	var selectedIdx int
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("Choose a feature to copy its path to the clipboard:").
				Options(options...).
				Value(&selectedIdx),
		),
	)
	check(form.Run())

	selectedFeature := features[selectedIdx]
	featurePath := selectedFeature.FullPath

	if err := clipboard.WriteAll(featurePath); err != nil {
		out.Pfailf("Failed to copy %s to clipboard: %v", out.H(featurePath), err)
		os.Exit(1)
	}
	fmt.Println(out.Successf("Copied %s to clipboard", out.H(featurePath)))
}

func check(err error) {
	if err != nil {
		fmt.Println(out.Failf("Error: %v", err))
		os.Exit(1)
	}
}
