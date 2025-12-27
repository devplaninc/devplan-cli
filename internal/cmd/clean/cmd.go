package clean

import (
	"fmt"
	"os"

	"github.com/devplaninc/devplan-cli/internal/out"
	"github.com/devplaninc/devplan-cli/internal/utils/git"
	"github.com/devplaninc/devplan-cli/internal/utils/workspace"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
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
		Label:     fmt.Sprintf("Worktree %s will be permanently deleted. Are you sure", featurePath),
		IsConfirm: true,
	}
	resp, err := confirm.Run()
	check(err)
	if resp != "y" {
		out.Pfailf("Deletion aborted")
		return
	}

	// Check if this is a worktree
	isWorktree, err := git.IsWorktree(featurePath)
	if err != nil {
		fmt.Println(out.Warnf("Could not determine if path is a worktree: %v", err))
		isWorktree = false
	}

	if isWorktree {
		// Get the main repo path
		mainRepoPath, err := git.GetMainRepoPath(featurePath)
		if err != nil {
			fmt.Println(out.Warnf("Could not get main repo path: %v", err))
			// Fall back to simple removal
			err = os.RemoveAll(featurePath)
			check(err)
		} else {
			// Remove the worktree using git
			err = git.RemoveWorktree(mainRepoPath, featurePath)
			if err != nil {
				fmt.Println(out.Warnf("Failed to remove worktree via git: %v", err))
				// Fall back to simple removal
				err = os.RemoveAll(featurePath)
				check(err)
			}

			// Prune worktrees to clean up administrative files
			_ = git.PruneWorktrees(mainRepoPath)
		}
	} else {
		// Not a worktree, just remove the directory
		err = os.RemoveAll(featurePath)
		check(err)
	}

	out.Psuccessf("Successfully deleted %s\n", out.H(featurePath))
}

func check(err error) {
	if err != nil {
		fmt.Println(out.Failf("Error: %v", err))
		os.Exit(1)
	}
}
