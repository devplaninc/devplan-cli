package workspace

import (
	"github.com/devplaninc/devplan-cli/internal/out"
	"github.com/spf13/viper"
	"os"
	"path"
	"path/filepath"
)

var defaultWorkspace = path.Join("devplan", "workspace")

const (
	workspaceConfigKey = "workspace_dir"
)

func GetPath() string {
	workspaceDir := viper.GetString(workspaceConfigKey)
	if workspaceDir == "" {
		// Use default directory in user's home
		home, err := os.UserHomeDir()
		if err != nil {
			out.Pfailf("Failed to get user home directory: %v\n", err)
			os.Exit(1)
		}
		workspaceDir = filepath.Join(home, defaultWorkspace)

		// Save to config for future use
		viper.Set(workspaceConfigKey, workspaceDir)
		err = viper.WriteConfig()
		if err != nil {
			out.Pfailf("Failed to write config: %v\n", err)
			// Continue anyway, just log the error
		}
	}

	// Ensure directory exists
	if _, err := os.Stat(workspaceDir); os.IsNotExist(err) {
		err = os.MkdirAll(workspaceDir, 0755)
		if err != nil {
			out.Pfailf("Failed to create workplace directory: %v\n", err)
			os.Exit(1)
		}
	}

	return workspaceDir
}

func GetFeaturesPath() string {
	return filepath.Join(GetPath(), "features")
}

func ListClonedFeatures() ([]string, error) {
	featuresPath := GetFeaturesPath()
	if _, err := os.Stat(featuresPath); os.IsNotExist(err) {
		return nil, nil
	}

	entries, err := os.ReadDir(featuresPath)
	if err != nil {
		return nil, err
	}

	var features []string
	for _, entry := range entries {
		if entry.IsDir() {
			features = append(features, entry.Name())
		}
	}
	return features, nil
}

func GetFeaturePath(dirName string) string {
	return filepath.Join(GetFeaturesPath(), dirName)
}
