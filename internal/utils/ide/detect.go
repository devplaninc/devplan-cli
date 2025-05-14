package ide

import (
	"os"
	"path/filepath"

	"github.com/devplaninc/devplan-cli/internal/utils/git"
)

// DetectAssistant returns a list of available IDEs based on the presence of IDE-specific files in the repository root
func DetectAssistant() ([]Assistant, error) {
	root, err := git.GetRoot()
	if err != nil {
		if git.IsNotInRepoErr(err) {
			return nil, nil
		}
		return nil, err
	}

	return DetectAssistantOnPath(root)
}

// DetectAssistantOnPath returns a list of available IDEs based on the presence of IDE-specific files on the path
func DetectAssistantOnPath(path string) ([]Assistant, error) {
	var result []Assistant
	for assistant, pathToCheck := range pathMap {
		if _, err := os.Stat(filepath.Join(path, pathToCheck)); err == nil {
			result = append(result, assistant)
		}
	}

	return result, nil
}
