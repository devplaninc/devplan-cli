package list_cmd

import (
	"fmt"
	"os"

	"github.com/atotto/clipboard"
	osc52 "github.com/aymanbagabas/go-osc52/v2"
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

	var selected common.FeatureSelection
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[common.FeatureSelection]().
				Title("Choose a feature to copy its path to the clipboard:").
				Options(options...).
				Value(&selected),
		),
	)
	check(form.Run())

	featurePath := selected.RepoPath
	if featurePath == "" {
		featurePath = features[selected.FeatureIdx].FullPath
	}

	if err := clipboard.WriteAll(featurePath); err != nil {
		// Fall back to OSC 52 escape sequence for SSH/tmux sessions
		// where no display server is available
		seq := osc52.New(featurePath)
		if os.Getenv("TMUX") != "" {
			seq = seq.Tmux()
		}
		if _, err := seq.WriteTo(os.Stderr); err != nil {
			out.Pfailf("Failed to copy %s to clipboard: %v", out.H(featurePath), err)
			os.Exit(1)
		}
		fmt.Println(out.Successf("Copied %s to clipboard (via terminal)", out.H(featurePath)))
		return
	}
	fmt.Println(out.Successf("Copied %s to clipboard", out.H(featurePath)))
}

func check(err error) {
	if err != nil {
		fmt.Println(out.Failf("Error: %v", err))
		os.Exit(1)
	}
}
