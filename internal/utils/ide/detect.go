package ide

import (
	"os"
	"path/filepath"

	"github.com/devplaninc/devplan-cli/internal/utils/git"
)

// IDE type constants
const (
	Cursor = "cursor"
	Junie  = "junie"
)

var pathMap = map[string]string{
	Cursor: ".cursor",
	Junie:  ".junie",
}

// Detect returns a list of available IDEs based on the presence of IDE-specific files in the repository root
func Detect() ([]string, error) {
	root, err := git.GetRoot()
	if err != nil {
		if git.IsNotInRepoErr(err) {
			return nil, nil
		}
		return nil, err
	}

	return DetectOnPath(root)
}

// DetectOnPath returns a list of available IDEs based on the presence of IDE-specific files on the path
func DetectOnPath(path string) ([]string, error) {
	var availableIDEs []string
	for ide, pathToCheck := range pathMap {
		if _, err := os.Stat(filepath.Join(path, pathToCheck)); err == nil {
			availableIDEs = append(availableIDEs, ide)
		}
	}

	return availableIDEs, nil
}
