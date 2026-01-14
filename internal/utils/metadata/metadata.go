package metadata

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	devplanDir       = ".devplan_meta"
	metaFile         = "meta.json"
	gitignore        = ".gitignore"
	gitignoreContent = "*\n"
)

// Metadata contains information stored in .devplan_meta/meta.json
type Metadata struct {
	TaskID      string `json:"taskId,omitempty"`
	StoryID     string `json:"storyId,omitempty"`
	TaskName    string `json:"taskName,omitempty"`
	StoryName   string `json:"storyName,omitempty"`
	ProjectID   string `json:"projectId,omitempty"`
	ProjectName string `json:"projectName,omitempty"`
	RepoURL     string `json:"repoUrl,omitempty"`
	RepoName    string `json:"repoName,omitempty"`

	ProjectNumericID string `json:"projectNumericId,omitempty"`
	StoryNumericID   string `json:"storyNumericId,omitempty"`
	TaskNumericID    string `json:"taskNumericId,omitempty"`
}

// GetDevplanDir returns the path to the .devplan_meta directory
func GetDevplanDir(repoPath string) string {
	return filepath.Join(repoPath, devplanDir)
}

// GetMetaFilePath returns the path to the meta.json file
func GetMetaFilePath(repoPath string) string {
	return filepath.Join(GetDevplanDir(repoPath), metaFile)
}

// EnsureDevplanDir creates the .devplan_meta directory if it doesn't exist
func EnsureDevplanDir(repoPath string) error {
	devplanPath := GetDevplanDir(repoPath)
	if err := os.MkdirAll(devplanPath, 0755); err != nil {
		return fmt.Errorf("failed to create %v directory: %w", devplanDir, err)
	}
	return nil
}

// WriteMetadata writes metadata to .devplan_meta/meta.json
func WriteMetadata(repoPath string, meta Metadata) error {
	if err := EnsureDevplanDir(repoPath); err != nil {
		return err
	}

	metaPath := GetMetaFilePath(repoPath)
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metaPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	return nil
}

// ReadMetadata reads metadata from .devplan_meta/meta.json
func ReadMetadata(repoPath string) (*Metadata, error) {
	metaPath := GetMetaFilePath(repoPath)
	data, err := os.ReadFile(metaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No metadata file, return nil
		}
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var meta Metadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &meta, nil
}

// EnsureGitignore ensures .devplan_meta/.gitignore exists and contains the correct content
func EnsureGitignore(repoPath string) error {
	if err := EnsureDevplanDir(repoPath); err != nil {
		return err
	}

	gitignorePath := filepath.Join(GetDevplanDir(repoPath), gitignore)

	// Check if gitignore already exists
	if _, err := os.Stat(gitignorePath); err == nil {
		// File exists, check content
		content, err := os.ReadFile(gitignorePath)
		if err != nil {
			return fmt.Errorf("failed to read %v/.gitignore: %w", devplanDir, err)
		}
		// If content is correct, nothing to do
		if string(content) == gitignoreContent {
			return nil
		}
	}

	// Write gitignore
	if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
		return fmt.Errorf("failed to write %v/.gitignore: %w", devplanDir, err)
	}

	return nil
}

// EnsureMetadataSetup ensures both the metadata file and gitignore are set up
func EnsureMetadataSetup(repoPath string, meta Metadata) error {
	if err := WriteMetadata(repoPath, meta); err != nil {
		return err
	}
	if err := EnsureGitignore(repoPath); err != nil {
		return err
	}
	return nil
}
