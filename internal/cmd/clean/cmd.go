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
	"github.com/devplaninc/devplan-cli/internal/utils/recentactivity"
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

	clonedFeatures = recentactivity.SortClonedFeatures(clonedFeatures)

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

	// Check for uncommitted changes across all repo paths
	hasChanges := false
	if len(selectedFeature.Repos) > 0 {
		for _, repoPath := range selectedFeature.GetRepoPaths() {
			changed, err := git.HasUncommittedChanges(repoPath)
			if err != nil {
				fmt.Println(out.Warnf("Could not check for uncommitted changes: %v", err))
			} else if changed {
				hasChanges = true
				break
			}
		}
	}

	// Build confirmation message
	entityLabel := "Worktree"
	if selectedFeature.IsFeatureWorkspace {
		entityLabel = "Directory"
	}
	confirmMsg := fmt.Sprintf("%s %s will be permanently deleted.", entityLabel, displayPath)
	if hasChanges {
		confirmMsg = fmt.Sprintf("⚠️  WARNING: %s has uncommitted changes!\n\n%s %s will be permanently deleted.", displayPath, entityLabel, displayPath)
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
		out.Pfailf("Deletion aborted\n")
		return
	}

	fmt.Printf("Cleaning up %s...\n", out.H(displayPath))

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
			fmt.Printf("Removing worktree...\n")
			err = git.RemoveWorktree(mainRepoPath, featurePath)
			if err == nil {
				// Successfully removed as worktree, prune administrative files
				fmt.Printf("Pruning worktrees in %s...\n", out.H(ide.PathWithTilde(mainRepoPath)))
				_ = git.PruneWorktrees(mainRepoPath)
			} else {
				// Failed to remove as worktree (might be a base branch), fall back to simple removal
				fmt.Printf("Failed to remove worktree, deleting directory %s...\n", out.H(displayPath))
				err = os.RemoveAll(featurePath)
				check(err)
			}
		} else {
			// Could not get main repo path, fall back to simple removal
			fmt.Printf("Deleting directory %s...\n", out.H(displayPath))
			err = os.RemoveAll(featurePath)
			check(err)
		}
	} else {
		// Not a worktree, just remove the directory
		fmt.Printf("Deleting directory %s...\n", out.H(displayPath))
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
