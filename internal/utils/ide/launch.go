package ide

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// LaunchIDE launches the specified IDE at the given repository path
func LaunchIDE(ide IDE, repoPath string) error {
	installed, err := DetectInstalledIDEs()
	if err != nil {
		return fmt.Errorf("failed to detect installed IDEs: %w", err)
	}
	return LaunchWithPath(installed[ide], repoPath)
}

// LaunchWithPath launches the specified IDE at the given repository path
func LaunchWithPath(idePath string, repoPath string) error {
	if _, err := os.Stat(idePath); os.IsNotExist(err) {
		return fmt.Errorf("IDE executable not found at %s", idePath)
	}

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return fmt.Errorf("repository path not found at %s", repoPath)
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin", "linux", "windows":
		cmd = exec.Command(idePath, repoPath)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Start()
}
