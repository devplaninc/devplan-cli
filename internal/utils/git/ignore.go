package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// UpdateIgnore creates or updates the .gitignore file
func UpdateIgnore() error {
	if err := updateIgnorePattenIfMissing("**/devplan_current_feature*.md*"); err != nil {
		return err
	}
	if err := updateIgnorePattenIfMissing("!.windsurf/rules/devplan_*.md"); err != nil {
		return err
	}
	if err := updateIgnorePattenIfMissing("!.windsurf/rules/devplan_current_feature*.md"); err != nil {
		return err
	}
	return nil
}

func updateIgnorePattenIfMissing(pattern string) error {
	rootDir, err := GetRoot()
	if err != nil {
		return fmt.Errorf("failed to get repository root: %w", err)
	}

	gitignorePath := filepath.Join(rootDir, ".gitignore")

	_, err = os.Stat(gitignorePath)
	if os.IsNotExist(err) {
		return os.WriteFile(gitignorePath, []byte(pattern+"\n"), 0644)
	} else if err != nil {
		return fmt.Errorf("failed to check .gitignore: %w", err)
	}

	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		return fmt.Errorf("failed to read .gitignore: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if line == pattern {
			// Pattern already exists, nothing to do
			return nil
		}
	}

	file, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open .gitignore for writing: %w", err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	if len(content) > 0 && !strings.HasSuffix(string(content), "\n") {
		if _, err := file.WriteString("\n"); err != nil {
			return fmt.Errorf("failed to write to .gitignore: %w", err)
		}
	}

	if _, err := file.WriteString(pattern + "\n"); err != nil {
		return fmt.Errorf("failed to write to .gitignore: %w", err)
	}

	return nil
}
