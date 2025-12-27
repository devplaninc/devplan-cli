package metadata

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDevplanDir(t *testing.T) {
	tests := []struct {
		name     string
		repoPath string
		expected string
	}{
		{
			name:     "simple path",
			repoPath: "/path/to/repo",
			expected: filepath.Join("/path/to/repo", ".devplan"),
		},
		{
			name:     "path with trailing slash",
			repoPath: "/path/to/repo/",
			expected: filepath.Join("/path/to/repo/", ".devplan"),
		},
		{
			name:     "relative path",
			repoPath: "repo",
			expected: filepath.Join("repo", ".devplan"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetDevplanDir(tt.repoPath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetMetaFilePath(t *testing.T) {
	tests := []struct {
		name     string
		repoPath string
		expected string
	}{
		{
			name:     "simple path",
			repoPath: "/path/to/repo",
			expected: filepath.Join("/path/to/repo", ".devplan", "meta.json"),
		},
		{
			name:     "relative path",
			repoPath: "repo",
			expected: filepath.Join("repo", ".devplan", "meta.json"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetMetaFilePath(tt.repoPath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEnsureDevplanDir(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("creates directory if it doesn't exist", func(t *testing.T) {
		repoPath := filepath.Join(tempDir, "test-repo")
		err := EnsureDevplanDir(repoPath)
		require.NoError(t, err)

		devplanPath := GetDevplanDir(repoPath)
		info, err := os.Stat(devplanPath)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})

	t.Run("succeeds if directory already exists", func(t *testing.T) {
		repoPath := filepath.Join(tempDir, "existing-repo")
		devplanPath := GetDevplanDir(repoPath)
		err := os.MkdirAll(devplanPath, 0755)
		require.NoError(t, err)

		err = EnsureDevplanDir(repoPath)
		require.NoError(t, err)
	})
}

func TestWriteAndReadMetadata(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("write and read full metadata", func(t *testing.T) {
		repoPath := filepath.Join(tempDir, "repo1")
		meta := Metadata{
			TaskID:      "task-123",
			TaskName:    "Add feature",
			ProjectID:   "proj-456",
			ProjectName: "My Project",
			RepoURL:     "https://github.com/user/repo.git",
			RepoName:    "user/repo",
		}

		err := WriteMetadata(repoPath, meta)
		require.NoError(t, err)

		// Verify file exists
		metaPath := GetMetaFilePath(repoPath)
		_, err = os.Stat(metaPath)
		require.NoError(t, err)

		// Read back
		readMeta, err := ReadMetadata(repoPath)
		require.NoError(t, err)
		require.NotNil(t, readMeta)

		assert.Equal(t, meta.TaskID, readMeta.TaskID)
		assert.Equal(t, meta.TaskName, readMeta.TaskName)
		assert.Equal(t, meta.ProjectID, readMeta.ProjectID)
		assert.Equal(t, meta.ProjectName, readMeta.ProjectName)
		assert.Equal(t, meta.RepoURL, readMeta.RepoURL)
		assert.Equal(t, meta.RepoName, readMeta.RepoName)
	})

	t.Run("write and read partial metadata", func(t *testing.T) {
		repoPath := filepath.Join(tempDir, "repo2")
		meta := Metadata{
			ProjectID:   "proj-789",
			ProjectName: "Another Project",
		}

		err := WriteMetadata(repoPath, meta)
		require.NoError(t, err)

		readMeta, err := ReadMetadata(repoPath)
		require.NoError(t, err)
		require.NotNil(t, readMeta)

		assert.Equal(t, meta.ProjectID, readMeta.ProjectID)
		assert.Equal(t, meta.ProjectName, readMeta.ProjectName)
		assert.Empty(t, readMeta.TaskID)
		assert.Empty(t, readMeta.TaskName)
	})

	t.Run("read non-existent metadata returns nil", func(t *testing.T) {
		repoPath := filepath.Join(tempDir, "nonexistent")

		readMeta, err := ReadMetadata(repoPath)
		require.NoError(t, err)
		assert.Nil(t, readMeta)
	})
}

func TestEnsureGitignore(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("creates gitignore if it doesn't exist", func(t *testing.T) {
		repoPath := filepath.Join(tempDir, "repo1")

		err := EnsureGitignore(repoPath)
		require.NoError(t, err)

		gitignorePath := filepath.Join(GetDevplanDir(repoPath), ".gitignore")
		content, err := os.ReadFile(gitignorePath)
		require.NoError(t, err)
		assert.Equal(t, "*\n", string(content))
	})

	t.Run("does not overwrite correct gitignore", func(t *testing.T) {
		repoPath := filepath.Join(tempDir, "repo2")
		devplanPath := GetDevplanDir(repoPath)
		err := os.MkdirAll(devplanPath, 0755)
		require.NoError(t, err)

		gitignorePath := filepath.Join(devplanPath, ".gitignore")
		err = os.WriteFile(gitignorePath, []byte("*\n"), 0644)
		require.NoError(t, err)

		// Get file info before
		infoBefore, err := os.Stat(gitignorePath)
		require.NoError(t, err)

		err = EnsureGitignore(repoPath)
		require.NoError(t, err)

		// Content should still be correct
		content, err := os.ReadFile(gitignorePath)
		require.NoError(t, err)
		assert.Equal(t, "*\n", string(content))

		// File should not have been rewritten (same mod time)
		infoAfter, err := os.Stat(gitignorePath)
		require.NoError(t, err)
		assert.Equal(t, infoBefore.ModTime(), infoAfter.ModTime())
	})

	t.Run("updates incorrect gitignore", func(t *testing.T) {
		repoPath := filepath.Join(tempDir, "repo3")
		devplanPath := GetDevplanDir(repoPath)
		err := os.MkdirAll(devplanPath, 0755)
		require.NoError(t, err)

		gitignorePath := filepath.Join(devplanPath, ".gitignore")
		err = os.WriteFile(gitignorePath, []byte("wrong content\n"), 0644)
		require.NoError(t, err)

		err = EnsureGitignore(repoPath)
		require.NoError(t, err)

		content, err := os.ReadFile(gitignorePath)
		require.NoError(t, err)
		assert.Equal(t, "*\n", string(content))
	})
}

func TestEnsureMetadataSetup(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("creates both metadata and gitignore", func(t *testing.T) {
		repoPath := filepath.Join(tempDir, "repo1")
		meta := Metadata{
			ProjectID:   "proj-123",
			ProjectName: "Test Project",
		}

		err := EnsureMetadataSetup(repoPath, meta)
		require.NoError(t, err)

		// Check metadata file
		metaPath := GetMetaFilePath(repoPath)
		_, err = os.Stat(metaPath)
		require.NoError(t, err)

		readMeta, err := ReadMetadata(repoPath)
		require.NoError(t, err)
		require.NotNil(t, readMeta)
		assert.Equal(t, meta.ProjectID, readMeta.ProjectID)

		// Check gitignore
		gitignorePath := filepath.Join(GetDevplanDir(repoPath), ".gitignore")
		content, err := os.ReadFile(gitignorePath)
		require.NoError(t, err)
		assert.Equal(t, "*\n", string(content))
	})
}

func TestMetadataJSONFormat(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("full metadata matches expected JSON format", func(t *testing.T) {
		repoPath := filepath.Join(tempDir, "repo1")
		meta := Metadata{
			TaskID:      "task-123",
			TaskName:    "Add user authentication",
			ProjectID:   "proj-456",
			ProjectName: "My Project",
			RepoURL:     "https://github.com/devplaninc/webapp.git",
			RepoName:    "devplaninc/webapp",
		}

		err := WriteMetadata(repoPath, meta)
		require.NoError(t, err)

		// Read the actual written file
		actualContent, err := os.ReadFile(GetMetaFilePath(repoPath))
		require.NoError(t, err)

		// Read the expected content from testdata
		expectedContent, err := os.ReadFile("testdata/full_metadata.json")
		require.NoError(t, err)

		// Compare JSON content (ignoring whitespace differences)
		assert.JSONEq(t, string(expectedContent), string(actualContent))
	})

	t.Run("project-only metadata matches expected JSON format", func(t *testing.T) {
		repoPath := filepath.Join(tempDir, "repo2")
		meta := Metadata{
			ProjectID:   "proj-789",
			ProjectName: "Backend Service",
			RepoURL:     "git@github.com:devplaninc/backend.git",
			RepoName:    "devplaninc/backend",
		}

		err := WriteMetadata(repoPath, meta)
		require.NoError(t, err)

		// Read the actual written file
		actualContent, err := os.ReadFile(GetMetaFilePath(repoPath))
		require.NoError(t, err)

		// Read the expected content from testdata
		expectedContent, err := os.ReadFile("testdata/project_only_metadata.json")
		require.NoError(t, err)

		// Compare JSON content (ignoring whitespace differences)
		assert.JSONEq(t, string(expectedContent), string(actualContent))
	})

	t.Run("verifies JSON indentation format", func(t *testing.T) {
		repoPath := filepath.Join(tempDir, "repo3")
		meta := Metadata{
			ProjectID:   "test-project",
			ProjectName: "Test",
		}

		err := WriteMetadata(repoPath, meta)
		require.NoError(t, err)

		// Read the actual written file
		actualContent, err := os.ReadFile(GetMetaFilePath(repoPath))
		require.NoError(t, err)

		// Verify it's pretty-printed with 2-space indentation and camelCase keys
		assert.Contains(t, string(actualContent), "{\n  \"projectId\":")
	})
}
