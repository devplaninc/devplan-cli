package ide

import (
	"fmt"
	"github.com/devplaninc/devplan-cli/internal/out"
	"github.com/devplaninc/devplan-cli/internal/utils/git"
	"os"
	"path/filepath"
	"strings"
)

// CleanupCurrentFeaturePrompts cleans up prompt files with "devplan_current_feature" in their name.
func CleanupCurrentFeaturePrompts(assistants []Assistant) error {
	for _, a := range assistants {
		dir, err := GetRulesPath(a)
		if err != nil {
			return err
		}
		if err := CleanupCurrentFeaturePromptsFromDir(dir); err != nil {
			return err
		}
	}
	return nil
}

func CleanupCurrentFeaturePromptsFromDir(directory string) error {
	root, err := git.GetRoot()
	if err != nil {
		return fmt.Errorf("failed to get repository root: %w", err)
	}

	dir := filepath.Join(root, directory)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil // Directory doesn't exist, skip it
	}
	// "devplan_current_feature"
	rulePref := fmt.Sprintf("%s%s", ruleFileNamePrefix, ruleFileNameCurrentFeaturePrefix)

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if strings.HasPrefix(info.Name(), rulePref) {
			if err := os.Remove(path); err != nil {
				out.Pfailf("Warning: Failed to delete prompt file %s: %v\n", path, err)
			}
			outPath := path
			if relativePath, err := filepath.Rel(root, path); err == nil {
				outPath = relativePath
			}
			out.Psuccessf("Deleted old prompt file %s\n", out.H(outPath))
		}

		return nil
	})

	if err != nil {
		out.Pfailf("Warning: Error while cleaning up prompt files: %v\n", err)
	}

	return nil
}
