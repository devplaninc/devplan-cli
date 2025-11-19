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
			out.Pfailf("Failed to get user home directory: %v\n", err)
			os.Exit(1)
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
			out.Pfailf("Failed to create workplace directory: %v\n", err)
			os.Exit(1)
		}
	}

	return workspaceDir
}

func GetFeaturesPath() string {
	return filepath.Join(GetPath(), "features")
}

type ClonedFeature struct {
	DirName  string
	FullPath string
	Repos    []ClonedRepo
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

	entries, err := os.ReadDir(featuresPath)
	if err != nil {
		return nil, err
	}

	var result []ClonedFeature
	for _, entry := range entries {
		if entry.IsDir() {
			fullPath := filepath.Join(featuresPath, entry.Name())
			result = append(result, ClonedFeature{
				DirName:  entry.Name(),
				FullPath: fullPath,
				Repos:    GetFeatureRepositories(fullPath),
			})
		}
	}
	return result, nil
}

func GetFeatureRepositories(featurePath string) []ClonedRepo {
	entries, err := os.ReadDir(featurePath)
	if err != nil {
		return nil
	}
	var result []ClonedRepo
	for _, entry := range entries {
		if entry.IsDir() {
			subDirPath := filepath.Join(featurePath, entry.Name())
			repo, err := git.RepoAtPath(subDirPath)
			if err == nil && len(repo.FullNames) > 0 {
				result = append(result, ClonedRepo{
					DirName: entry.Name(),
					Repo:    repo,
				})
			}
		}
	}
	return result
}

func GetFeaturePath(dirName string) string {
	return filepath.Join(GetFeaturesPath(), dirName)
}
