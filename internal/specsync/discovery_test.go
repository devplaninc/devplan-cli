package specsync

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscoverSpecs(t *testing.T) {
	t.Run("discovers markdown files", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create test files
		err := os.WriteFile(filepath.Join(tmpDir, "readme.md"), []byte("# Readme"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(tmpDir, "instructions.md"), []byte("# Instructions"), 0644)
		require.NoError(t, err)

		specs, err := DiscoverTaskSpecs(tmpDir)
		require.NoError(t, err)
		assert.Len(t, specs, 1)

		names := make(map[string]bool)
		for _, s := range specs {
			names[s.Name] = true
			assert.NotEmpty(t, s.Checksum)
			assert.NotEmpty(t, s.Path)
		}
		assert.Contains(t, names, "readme.md")
	})

	t.Run("discovers nested markdown files", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create nested structure
		storyDir := filepath.Join(tmpDir, "STORY-123")
		taskDir := filepath.Join(storyDir, "TASK-456")
		err := os.MkdirAll(taskDir, 0755)
		require.NoError(t, err)

		err = os.WriteFile(filepath.Join(tmpDir, "prd.md"), []byte("# PRD"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(storyDir, "requirements.md"), []byte("# Story"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(taskDir, "instructions.md"), []byte("# Task"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(taskDir, "plan.md"), []byte("# Task"), 0644)
		require.NoError(t, err)

		specs, err := DiscoverTaskSpecs(tmpDir)
		require.NoError(t, err)
		assert.Len(t, specs, 1)

		names := make(map[string]bool)
		for _, s := range specs {
			names[s.Name] = true
		}
		assert.True(t, names["plan.md"])
	})

	t.Run("ignores non-markdown files", func(t *testing.T) {
		tmpDir := t.TempDir()

		err := os.WriteFile(filepath.Join(tmpDir, "readme.md"), []byte("# Readme"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(tmpDir, "data.json"), []byte("{}"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(tmpDir, "script.py"), []byte("print('hi')"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("text"), 0644)
		require.NoError(t, err)

		specs, err := DiscoverTaskSpecs(tmpDir)
		require.NoError(t, err)
		assert.Len(t, specs, 1)
		assert.Equal(t, "readme.md", specs[0].Name)
	})

	t.Run("ignores hidden files", func(t *testing.T) {
		tmpDir := t.TempDir()

		err := os.WriteFile(filepath.Join(tmpDir, "readme.md"), []byte("# Readme"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(tmpDir, ".hidden.md"), []byte("# Hidden"), 0644)
		require.NoError(t, err)

		specs, err := DiscoverTaskSpecs(tmpDir)
		require.NoError(t, err)
		assert.Len(t, specs, 1)
		assert.Equal(t, "readme.md", specs[0].Name)
	})

	t.Run("ignores hidden directories", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create hidden directory with markdown file
		hiddenDir := filepath.Join(tmpDir, ".hidden")
		err := os.MkdirAll(hiddenDir, 0755)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(hiddenDir, "secret.md"), []byte("# Secret"), 0644)
		require.NoError(t, err)

		// Create visible markdown file
		err = os.WriteFile(filepath.Join(tmpDir, "readme.md"), []byte("# Readme"), 0644)
		require.NoError(t, err)

		specs, err := DiscoverTaskSpecs(tmpDir)
		require.NoError(t, err)
		assert.Len(t, specs, 1)
		assert.Equal(t, "readme.md", specs[0].Name)
	})

	t.Run("empty directory returns empty slice", func(t *testing.T) {
		tmpDir := t.TempDir()

		specs, err := DiscoverTaskSpecs(tmpDir)
		require.NoError(t, err)
		assert.Empty(t, specs)
	})

	t.Run("non-existent directory returns error", func(t *testing.T) {
		_, err := DiscoverTaskSpecs("/nonexistent/path")
		assert.Error(t, err)
	})

	t.Run("calculates correct checksums", func(t *testing.T) {
		tmpDir := t.TempDir()

		content := []byte("# Test Content\nWith some text")
		err := os.WriteFile(filepath.Join(tmpDir, "test.md"), content, 0644)
		require.NoError(t, err)

		expectedChecksum := CalculateChecksumBytes(content)

		specs, err := DiscoverTaskSpecs(tmpDir)
		require.NoError(t, err)
		require.Len(t, specs, 1)
		assert.Equal(t, expectedChecksum, specs[0].Checksum)
	})
}
