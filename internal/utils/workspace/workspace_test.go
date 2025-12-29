package workspace

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/devplaninc/devplan-cli/internal/utils/git"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestClonedFeature_GetDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		feature  ClonedFeature
		expected string
	}{
		{
			name: "single repo",
			feature: ClonedFeature{
				DirName: "feature1",
				Repos: []ClonedRepo{
					{
						Repo: git.RepoInfo{
							FullNames: []string{"devplaninc/webapp"},
						},
					},
				},
			},
			expected: "feature1 (devplaninc/webapp)",
		},
		{
			name: "multiple repos",
			feature: ClonedFeature{
				DirName: "feature2",
				Repos: []ClonedRepo{
					{
						Repo: git.RepoInfo{
							FullNames: []string{"devplaninc/webapp"},
						},
					},
					{
						Repo: git.RepoInfo{
							FullNames: []string{"devplaninc/api"},
						},
					},
				},
			},
			expected: "feature2 (devplaninc/webapp, devplaninc/api)",
		},
		{
			name: "no repos",
			feature: ClonedFeature{
				DirName: "feature3",
				Repos:   []ClonedRepo{},
			},
			expected: "feature3",
		},
		{
			name: "repo with no full names",
			feature: ClonedFeature{
				DirName: "feature4",
				Repos: []ClonedRepo{
					{
						Repo: git.RepoInfo{
							FullNames: []string{},
						},
					},
				},
			},
			expected: "feature4",
		},
		{
			name: "mixed repos with and without names",
			feature: ClonedFeature{
				DirName: "feature5",
				Repos: []ClonedRepo{
					{
						Repo: git.RepoInfo{
							FullNames: []string{},
						},
					},
					{
						Repo: git.RepoInfo{
							FullNames: []string{"devplaninc/backend"},
						},
					},
				},
			},
			expected: "feature5 (devplaninc/backend)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.feature.GetDisplayName()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPathFunctions(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Set viper to use the temp directory
	viper.Set(workspaceConfigKey, tempDir)

	// Ensure the directory is created
	err := os.MkdirAll(tempDir, 0755)
	assert.NoError(t, err)

	t.Run("GetProjectFeaturesPath", func(t *testing.T) {
		tests := []struct {
			name        string
			projectName string
			expected    string
		}{
			{
				name:        "simple project name",
				projectName: "myproject",
				expected:    filepath.Join(tempDir, "features", "myproject"),
			},
			{
				name:        "project with underscores",
				projectName: "my_project",
				expected:    filepath.Join(tempDir, "features", "my_project"),
			},
			{
				name:        "project with dashes",
				projectName: "my-project",
				expected:    filepath.Join(tempDir, "features", "my-project"),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := GetProjectFeaturesPath(tt.projectName)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("GetMainRepoPath", func(t *testing.T) {
		tests := []struct {
			name        string
			projectName string
			repoName    string
			expected    string
		}{
			{
				name:        "simple paths",
				projectName: "myproject",
				repoName:    "webapp",
				expected:    filepath.Join(tempDir, "features", "myproject", "webapp"),
			},
			{
				name:        "project and repo with special chars",
				projectName: "my-project",
				repoName:    "web_app",
				expected:    filepath.Join(tempDir, "features", "my-project", "web_app"),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := GetMainRepoPath(tt.projectName, tt.repoName)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("GetWorktreePath", func(t *testing.T) {
		tests := []struct {
			name        string
			projectName string
			taskName    string
			expected    string
		}{
			{
				name:        "simple paths",
				projectName: "myproject",
				taskName:    "feature-123",
				expected:    filepath.Join(tempDir, "features", "myproject", "feature-123"),
			},
			{
				name:        "paths with underscores",
				projectName: "my_project",
				taskName:    "task_456",
				expected:    filepath.Join(tempDir, "features", "my_project", "task_456"),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := GetWorktreePath(tt.projectName, tt.taskName)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("GetFeaturePath", func(t *testing.T) {
		tests := []struct {
			name     string
			dirName  string
			expected string
		}{
			{
				name:     "simple directory name",
				dirName:  "feature1",
				expected: filepath.Join(tempDir, "features", "feature1"),
			},
			{
				name:     "directory with path separator",
				dirName:  "project/feature",
				expected: filepath.Join(tempDir, "features", "project", "feature"),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := GetFeaturePath(tt.dirName)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("ListClonedFeatures returns project folder if no repos in subfolders", func(t *testing.T) {
		// Create a temporary directory for testing
		tempDir := t.TempDir()
		viper.Set(workspaceConfigKey, tempDir)

		featuresDir := filepath.Join(tempDir, "features")
		projectDir := filepath.Join(featuresDir, "myproject")
		subDir1 := filepath.Join(projectDir, "sub1")
		subDir2 := filepath.Join(projectDir, "sub2")

		err := os.MkdirAll(subDir1, 0755)
		assert.NoError(t, err)
		err = os.MkdirAll(subDir2, 0755)
		assert.NoError(t, err)

		features, err := ListClonedFeatures()
		assert.NoError(t, err)
		assert.Len(t, features, 1)
		assert.Equal(t, "myproject", features[0].DirName)
		assert.Equal(t, projectDir, features[0].FullPath)
		assert.Empty(t, features[0].Repos)
	})

	t.Run("ListClonedFeatures returns subfolders if at least one repo exists", func(t *testing.T) {
		// Create a temporary directory for testing
		tempDir := t.TempDir()
		viper.Set(workspaceConfigKey, tempDir)

		featuresDir := filepath.Join(tempDir, "features")
		projectDir := filepath.Join(featuresDir, "myproject")
		repoDir := filepath.Join(projectDir, "repo")
		nonRepoDir := filepath.Join(projectDir, "junk")

		err := os.MkdirAll(repoDir, 0755)
		assert.NoError(t, err)
		err = os.MkdirAll(nonRepoDir, 0755)
		assert.NoError(t, err)

		// Make repoDir a git repo
		cmd := exec.Command("git", "-C", repoDir, "init")
		err = cmd.Run()
		assert.NoError(t, err)

		cmd = exec.Command("git", "-C", repoDir, "remote", "add", "origin", "https://github.com/user/repo.git")
		err = cmd.Run()
		assert.NoError(t, err)

		features, err := ListClonedFeatures()
		assert.NoError(t, err)
		assert.Len(t, features, 2)

		// They should be subfolders
		dirNames := []string{features[0].DirName, features[1].DirName}
		assert.Contains(t, dirNames, "myproject/repo")
		assert.Contains(t, dirNames, "myproject/junk")
	})

	t.Run("ListClonedRepos excludes non-git directories", func(t *testing.T) {
		// Create a temporary directory for testing
		tempDir := t.TempDir()
		viper.Set(workspaceConfigKey, tempDir)

		featuresDir := filepath.Join(tempDir, "features")
		projectDir := filepath.Join(featuresDir, "myproject")
		nonGitDir := filepath.Join(projectDir, "non-git-dir")

		err := os.MkdirAll(nonGitDir, 0755)
		assert.NoError(t, err)

		features, err := ListClonedRepos()
		assert.NoError(t, err)
		assert.Len(t, features, 0)
	})
}
