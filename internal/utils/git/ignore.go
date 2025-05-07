package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// UpdateIgnore creates or updates the .gitignore file
// to ignore files named "devplan_current_feature" with extension .md* anywhere in the repository
func UpdateIgnore() error {
	// Get the repository root directory
	rootDir, err := GetRoot()
	if err != nil {
		return fmt.Errorf("failed to get repository root: %w", err)
	}

	// Path to .gitignore file
	gitignorePath := filepath.Join(rootDir, ".gitignore")

	// Pattern to add to .gitignore - using **/ to match files anywhere in the repository
	pattern := "**/devplan_current_feature.md*"

	// Check if .gitignore exists
	_, err = os.Stat(gitignorePath)
	if os.IsNotExist(err) {
		// Create .gitignore with the pattern
		return os.WriteFile(gitignorePath, []byte(pattern+"\n"), 0644)
	} else if err != nil {
		return fmt.Errorf("failed to check .gitignore: %w", err)
	}

	// Read existing .gitignore
	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		return fmt.Errorf("failed to read .gitignore: %w", err)
	}

	// Check if pattern already exists
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if line == pattern {
			// Pattern already exists, nothing to do
			return nil
		}
	}

	// Append pattern to .gitignore
	file, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open .gitignore for writing: %w", err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	// Add a newline if the file doesn't end with one
	if len(content) > 0 && !strings.HasSuffix(string(content), "\n") {
		if _, err := file.WriteString("\n"); err != nil {
			return fmt.Errorf("failed to write to .gitignore: %w", err)
		}
	}

	// Write the pattern
	if _, err := file.WriteString(pattern + "\n"); err != nil {
		return fmt.Errorf("failed to write to .gitignore: %w", err)
	}

	return nil
}
