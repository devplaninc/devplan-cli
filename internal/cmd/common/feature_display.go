package common

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/devplaninc/devplan-cli/internal/utils/git"
	"github.com/devplaninc/devplan-cli/internal/utils/metadata"
	"github.com/devplaninc/devplan-cli/internal/utils/workspace"
)

// FeatureSelection identifies the selected feature and specific repo path
type FeatureSelection struct {
	FeatureIdx int
	RepoPath   string
}

// repoDetail holds per-repo display info
type repoDetail struct {
	name       string
	branch     string
	path       string
	hasChanges bool
}

// featureInfo holds enriched display data for a single feature
type featureInfo struct {
	name        string
	featureIdx  int
	featurePath string
	repos       []repoDetail
	hasChanges  bool
}

// BuildFeatureOptions creates selectable options for features.
// Each feature has a parent node (the feature name) and sub-items per repo.
func BuildFeatureOptions(features []workspace.ClonedFeature) ([]huh.Option[FeatureSelection], bool) {
	infos := collectFeatureInfos(features)

	hasAnyChanges := false
	var options []huh.Option[FeatureSelection]
	for _, info := range infos {
		if info.hasChanges {
			hasAnyChanges = true
		}

		if len(info.repos) == 0 {
			// No repos: flat item
			label := info.name
			sel := FeatureSelection{FeatureIdx: info.featureIdx, RepoPath: info.featurePath}
			options = append(options, huh.NewOption(label, sel))
		} else {
			// Parent node: feature name, selects feature directory
			parentLabel := info.name
			if info.hasChanges {
				parentLabel += " *"
			}
			parentSel := FeatureSelection{FeatureIdx: info.featureIdx, RepoPath: info.featurePath}
			options = append(options, huh.NewOption(parentLabel, parentSel))

			// Sub-items: one per repo
			for j, r := range info.repos {
				connector := "├"
				if j == len(info.repos)-1 {
					connector = "└"
				}
				label := fmt.Sprintf("    %s %s", connector, r.name)
				if r.hasChanges {
					label += " *"
				}
				if r.branch != "" {
					label += " · " + r.branch
				}
				sel := FeatureSelection{FeatureIdx: info.featureIdx, RepoPath: r.path}
				options = append(options, huh.NewOption(label, sel))
			}
		}
	}

	return options, hasAnyChanges
}

// ShowLegend displays the legend for uncommitted changes indicator
func ShowLegend() {
	legendStyle := lipgloss.NewStyle().Italic(true)
	fmt.Println(legendStyle.Render("Legend: * = uncommitted changes (excluding untracked files)"))
	fmt.Println()
}

// collectFeatureInfos gathers enriched display data for each feature
func collectFeatureInfos(features []workspace.ClonedFeature) []featureInfo {
	infos := make([]featureInfo, 0, len(features))
	for i, f := range features {
		info := featureInfo{
			name:        f.DirName,
			featureIdx:  i,
			featurePath: f.FullPath,
		}

		// Try to get a nicer name from metadata
		if meta := readFeatureMetadata(f); meta != nil {
			if meta.TaskName != "" {
				info.name = meta.TaskName
			} else if meta.StoryName != "" {
				info.name = meta.StoryName
			}
		}

		// Collect per-repo details
		repoPaths := f.GetRepoPaths()
		for j, repo := range f.Repos {
			rd := repoDetail{}
			if len(repo.Repo.FullNames) > 0 {
				rd.name = filepath.Base(repo.Repo.FullNames[0])
			} else {
				rd.name = repo.DirName
			}

			rd.path = repoPaths[j]
			if branch, err := git.GetCurrentBranch(rd.path); err == nil {
				rd.branch = branch
			}
			if changed, err := git.HasUncommittedChanges(rd.path); err == nil && changed {
				rd.hasChanges = true
				info.hasChanges = true
			}

			info.repos = append(info.repos, rd)
		}

		// Remove repo suffix from name if DirName ends with /repoName (single-repo case)
		if len(info.repos) == 1 {
			suffix := "/" + info.repos[0].name
			info.name = strings.TrimSuffix(info.name, suffix)
		}

		infos = append(infos, info)
	}
	return infos
}

// readFeatureMetadata tries to read metadata from the feature path and its repo paths
func readFeatureMetadata(f workspace.ClonedFeature) *metadata.Metadata {
	// Try feature path first
	if meta, err := metadata.ReadMetadata(f.FullPath); err == nil && meta != nil {
		return meta
	}
	// For feature workspaces, try each repo path
	if f.IsFeatureWorkspace {
		for _, repoPath := range f.GetRepoPaths() {
			if meta, err := metadata.ReadMetadata(repoPath); err == nil && meta != nil {
				return meta
			}
		}
	}
	return nil
}
