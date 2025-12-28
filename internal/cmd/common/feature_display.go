package common

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/devplaninc/devplan-cli/internal/utils/git"
	"github.com/devplaninc/devplan-cli/internal/utils/ide"
	"github.com/devplaninc/devplan-cli/internal/utils/workspace"
)

// BuildFeatureOptions creates formatted display options for features with uncommitted change indicators
// Returns the options list and whether any features have uncommitted changes (for legend display)
func BuildFeatureOptions(features []workspace.ClonedFeature) ([]huh.Option[int], bool) {
	var options []huh.Option[int]
	boldStyle := lipgloss.NewStyle().Bold(true)
	highlightStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("12")) // Light blue color
	warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11"))      // Yellow for warning

	hasAnyChanges := false
	for i, f := range features {
		displayPath := ide.PathWithTilde(f.FullPath)

		// Check for uncommitted changes
		hasChanges, err := git.HasUncommittedChanges(f.FullPath)
		if err != nil {
			hasChanges = false // Ignore errors, just don't show indicator
		}
		if hasChanges {
			hasAnyChanges = true
		}

		// Build repo names part
		var repoNames []string
		for _, repo := range f.Repos {
			if len(repo.Repo.FullNames) > 0 {
				repoNames = append(repoNames, repo.Repo.FullNames[0])
			}
		}

		// Add uncommitted changes indicator after the branch name
		nameWithIndicator := boldStyle.Render(f.DirName)
		if hasChanges {
			nameWithIndicator += " " + warnStyle.Render("*")
		}

		// Format: {DirName bold} * ({RepoNames highlighted}) ({FullPath plain})
		var label string
		if len(repoNames) > 0 {
			label = fmt.Sprintf("%s %s (%s)",
				nameWithIndicator,
				highlightStyle.Render("("+strings.Join(repoNames, ", ")+")"),
				displayPath)
		} else {
			label = fmt.Sprintf("%s (%s)",
				nameWithIndicator,
				displayPath)
		}

		options = append(options, huh.NewOption(label, i))
	}

	return options, hasAnyChanges
}

// ShowLegend displays the legend for uncommitted changes indicator
func ShowLegend() {
	legendStyle := lipgloss.NewStyle().Italic(true)
	fmt.Println(legendStyle.Render("Legend: * = uncommitted changes (excluding untracked files)"))
	fmt.Println()
}
