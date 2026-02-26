package workspace

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/devplaninc/devplan-cli/internal/out"
	"github.com/devplaninc/devplan-cli/internal/utils/git"
	"github.com/spf13/viper"
)

var defaultWorkspace = path.Join("devplan", "workspace")

const (
	workspaceConfigKey = "workspace_dir"
)

func GetPath() string {
	workspaceDir := viper.GetString(workspaceConfigKey)
	if workspaceDir == "" {
		// Use default directory in user's home
		home, err := os.UserHomeDir()
		if err != nil {
			panic(fmt.Sprintf("Failed to get user home directory: %v", err))
		}
		workspaceDir = filepath.Join(home, defaultWorkspace)

		// Save to config for future use
		viper.Set(workspaceConfigKey, workspaceDir)
		err = viper.WriteConfig()
		if err != nil {
			out.Pfailf("Failed to write config: %v\n", err)
			// Continue anyway, just log the error
		}
	}

	// Ensure directory exists
	if _, err := os.Stat(workspaceDir); os.IsNotExist(err) {
		err = os.MkdirAll(workspaceDir, 0755)
		if err != nil {
			panic(fmt.Sprintf("Failed to create workplace directory: %v", err))
		}
	}

	return workspaceDir
}

func GetFeaturesPath() string {
	return filepath.Join(GetPath(), "features")
}

// GetProjectFeaturesPath returns the path to the features directory for a specific project
func GetProjectFeaturesPath(projectName string) string {
	return filepath.Join(GetFeaturesPath(), projectName)
}

type ClonedFeature struct {
	DirName            string
	FullPath           string
	Repos              []ClonedRepo
	IsFeatureWorkspace bool
}

// GetRepoPaths returns the paths to all git repositories in this feature.
// For task worktrees (single repo), returns []string{f.FullPath}.
// For feature workspaces (multiple repos), returns a path for each repo.
func (f ClonedFeature) GetRepoPaths() []string {
	if f.IsFeatureWorkspace {
		var paths []string
		for _, r := range f.Repos {
			paths = append(paths, filepath.Join(f.FullPath, r.DirName))
		}
		return paths
	}
	return []string{f.FullPath}
}

func (f ClonedFeature) GetDisplayName() string {
	var repoNames []string
	for _, repo := range f.Repos {
		if len(repo.Repo.FullNames) > 0 {
			repoNames = append(repoNames, repo.Repo.FullNames[0])
		}
	}
	if len(repoNames) == 0 {
		return f.DirName
	}
	return fmt.Sprintf("%s (%s)", f.DirName, strings.Join(repoNames, ", "))
}

type ClonedRepo struct {
	DirName string
	Repo    git.RepoInfo
}

func ListClonedFeatures() ([]ClonedFeature, error) {
	featuresPath := GetFeaturesPath()
	if _, err := os.Stat(featuresPath); os.IsNotExist(err) {
		return nil, nil
	}

	// Read project directories
	projectEntries, err := os.ReadDir(featuresPath)
	if err != nil {
		return nil, err
	}

	var result []ClonedFeature
	// Iterate through each project directory
	for _, projectEntry := range projectEntries {
		if !projectEntry.IsDir() {
			continue
		}

		projectPath := filepath.Join(featuresPath, projectEntry.Name())

		// Read subdirectories (repos and worktrees)
		subEntries, err := os.ReadDir(projectPath)
		if err != nil {
			continue
		}

		var projectFeatures []ClonedFeature
		hasRepo := false
		for _, subEntry := range subEntries {
			if !subEntry.IsDir() {
				continue
			}

			fullPath := filepath.Join(projectPath, subEntry.Name())

			// Create a display name that includes the project
			displayName := filepath.Join(projectEntry.Name(), subEntry.Name())

			feature := ClonedFeature{
				DirName:  displayName,
				FullPath: fullPath,
			}

			// Check if it's a git repository (task worktree pattern)
			repo, err := git.RepoAtPath(fullPath)
			if err == nil {
				feature.Repos = []ClonedRepo{
					{
						DirName: subEntry.Name(),
						Repo:    repo,
					},
				}
				hasRepo = true
			} else {
				// Not a git repo — check if it's a feature workspace with child repos
				childEntries, childErr := os.ReadDir(fullPath)
				if childErr == nil {
					var childRepos []ClonedRepo
					for _, childEntry := range childEntries {
						if !childEntry.IsDir() {
							continue
						}
						childPath := filepath.Join(fullPath, childEntry.Name())
						childRepo, childRepoErr := git.RepoAtPath(childPath)
						if childRepoErr == nil {
							childRepos = append(childRepos, ClonedRepo{
								DirName: childEntry.Name(),
								Repo:    childRepo,
							})
						}
					}
					if len(childRepos) > 0 {
						feature.Repos = childRepos
						feature.IsFeatureWorkspace = true
						hasRepo = true
					}
				}
			}

			projectFeatures = append(projectFeatures, feature)
		}

		if hasRepo {
			result = append(result, projectFeatures...)
		} else {
			// If no repos found in subfolders, just show the project directory
			result = append(result, ClonedFeature{
				DirName:  projectEntry.Name(),
				FullPath: projectPath,
			})
		}
	}
	return result, nil
}

// ListClonedRepos returns only those features that are valid git repositories
func ListClonedRepos() ([]ClonedFeature, error) {
	all, err := ListClonedFeatures()
	if err != nil {
		return nil, err
	}
	var result []ClonedFeature
	for _, f := range all {
		if len(f.Repos) > 0 {
			result = append(result, f)
		}
	}
	return result, nil
}

func GetFeaturePath(dirName string) string {
	return filepath.Join(GetFeaturesPath(), dirName)
}

// GetMainRepoPath returns the path to the main repository for a project
// Main repo is located at features/{projectName}/{repoName}/
func GetMainRepoPath(projectName, repoName string) string {
	return filepath.Join(GetProjectFeaturesPath(projectName), repoName)
}

// GetWorktreePath returns the path to a worktree for a specific project and task
// Worktree is located at features/{projectName}/{taskName}/
func GetWorktreePath(projectName, taskName string) string {
	return filepath.Join(GetProjectFeaturesPath(projectName), taskName)
}

// GetFeatureWorkspacePath returns the path to a feature workspace containing multiple repos.
// Feature workspace is located at features/{projectName}/{featureName}/
// Note: shares the same parent directory as GetWorktreePath (task worktrees). In practice,
// feature names and task names are sanitized differently, so collisions are unlikely.
func GetFeatureWorkspacePath(projectName, featureName string) string {
	return filepath.Join(GetProjectFeaturesPath(projectName), featureName)
}

// MainRepoExists checks if the main repository for a project exists
func MainRepoExists(projectName, repoName string) bool {
	repoPath := GetMainRepoPath(projectName, repoName)
	if _, err := os.Stat(repoPath); err == nil {
		// Check if it's actually a git repository
		_, err := git.RepoAtPath(repoPath)
		return err == nil
	}
	return false
}
