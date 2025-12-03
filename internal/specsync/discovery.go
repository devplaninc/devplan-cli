package specsync

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/exp/slices"
)

var inputSpecs = []string{"prd.md", "tech_brief.md", "requirements.md", "instructions.md", "dependencies.md"}

// DiscoverTaskSpecs walks the specs directory and finds all artifact files
func DiscoverTaskSpecs(taskDir string) ([]Spec, error) {
	var specs []Spec

	err := filepath.Walk(taskDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and hidden files
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip hidden files
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}
		if slices.Contains(inputSpecs, info.Name()) {
			return nil
		}

		// Only include markdown files as specs
		if !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}

		// Calculate checksum
		checksum, data, err := calculateChecksum(path)
		if err != nil {
			return fmt.Errorf("failed to calculate checksum for %s: %w", path, err)
		}

		specs = append(specs, Spec{
			Name:     info.Name(),
			Path:     path,
			Checksum: checksum,
			Content:  data,
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk specs directory %s: %w", taskDir, err)
	}

	return specs, nil
}
