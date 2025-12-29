package clean

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/devplaninc/devplan-cli/internal/cmd/common"
	"github.com/devplaninc/devplan-cli/internal/out"
	"github.com/devplaninc/devplan-cli/internal/utils/git"
	"github.com/devplaninc/devplan-cli/internal/utils/ide"
	"github.com/devplaninc/devplan-cli/internal/utils/workspace"
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

	// Build options for selection with full paths
	options, hasAnyChanges := common.BuildFeatureOptions(clonedFeatures)

	// Show legend if there are any uncommitted changes
	if hasAnyChanges {
		common.ShowLegend()
	}

	var selectedIdx int
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("Select a feature to clean up").
				Options(options...).
				Value(&selectedIdx),
		),
	)

	err = form.Run()
	check(err)

	selectedFeature := clonedFeatures[selectedIdx]
	featurePath := selectedFeature.FullPath
	displayPath := ide.PathWithTilde(featurePath)

	// Check for uncommitted changes
	hasChanges := false
	if len(selectedFeature.Repos) > 0 {
		var err error
		hasChanges, err = git.HasUncommittedChanges(featurePath)
		if err != nil {
			fmt.Println(out.Warnf("Could not check for uncommitted changes: %v", err))
		}
	}

	// Build confirmation message
	confirmMsg := fmt.Sprintf("Worktree %s will be permanently deleted.", displayPath)
	if hasChanges {
		confirmMsg = fmt.Sprintf("⚠️  WARNING: %s has uncommitted changes!\n\nWorktree %s will be permanently deleted.", displayPath, displayPath)
	}

	var confirm bool
	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(confirmMsg).
				Affirmative("Yes, delete it").
				Negative("No, keep it").
				Value(&confirm),
		),
	)

	err = confirmForm.Run()
	check(err)

	if !confirm {
		out.Pfailf("Deletion aborted")
		return
	}

	// Check if this is a worktree
	isWorktree := false
	if len(selectedFeature.Repos) > 0 {
		var err error
		isWorktree, err = git.IsWorktree(featurePath)
		if err != nil {
			// Could not determine, assume it's not a worktree
			isWorktree = false
		}
	}

	// Store the parent directory before deletion
	parentDir := filepath.Dir(featurePath)

	if isWorktree {
		// Get the main repo path
		mainRepoPath, err := git.GetMainRepoPath(featurePath)
		if err == nil {
			// Try to remove the worktree using git
			err = git.RemoveWorktree(mainRepoPath, featurePath)
			if err == nil {
				// Successfully removed as worktree, prune administrative files
				_ = git.PruneWorktrees(mainRepoPath)
			} else {
				// Failed to remove as worktree (might be a base branch), fall back to simple removal
				err = os.RemoveAll(featurePath)
				check(err)
			}
		} else {
			// Could not get main repo path, fall back to simple removal
			err = os.RemoveAll(featurePath)
			check(err)
		}
	} else {
		// Not a worktree, just remove the directory
		err = os.RemoveAll(featurePath)
		check(err)
	}

	// Check if parent directory is empty and remove it
	if entries, err := os.ReadDir(parentDir); err == nil && len(entries) == 0 {
		if err := os.Remove(parentDir); err == nil {
			parentDisplayPath := ide.PathWithTilde(parentDir)
			out.Psuccessf("Successfully deleted %s and empty parent directory %s\n", out.H(displayPath), out.H(parentDisplayPath))
		} else {
			out.Psuccessf("Successfully deleted %s\n", out.H(displayPath))
		}
	} else {
		out.Psuccessf("Successfully deleted %s\n", out.H(displayPath))
	}
}

func check(err error) {
	if err != nil {
		fmt.Println(out.Failf("Error: %v", err))
		os.Exit(1)
	}
}
