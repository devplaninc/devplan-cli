package ide

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// LaunchIDE launches the specified IDE at the given repository path
func LaunchIDE(ide IDE, repoPath string) (bool, error) {
	installed, err := DetectInstalledIDEs()
	if err != nil {
		return false, fmt.Errorf("failed to detect installed IDEs: %w", err)
	}
	return LaunchWithPath(installed[ide], repoPath)
}

// LaunchWithPath launches the specified IDE at the given repository path
func LaunchWithPath(idePath string, repoPath string) (bool, error) {
	if _, err := os.Stat(idePath); os.IsNotExist(err) {
		return false, fmt.Errorf("IDE executable not found at %s", idePath)
	}

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return false, fmt.Errorf("repository path not found at %s", repoPath)
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin", "linux", "windows":
		if isClaudeExecutable(idePath) {
			return launchClaudeInTerminal(idePath, repoPath)
		}
		cmd = exec.Command(idePath, repoPath)
	default:
		return false, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return true, cmd.Start()
}

func isClaudeExecutable(idePath string) bool {
	// Check if the executable name contains "claude"
	execName := filepath.Base(idePath)
	return strings.Contains(strings.ToLower(execName), "claude")
}

// launchClaudeInTerminal launches Claude in a new terminal session
func launchClaudeInTerminal(idePath string, repoPath string) (bool, error) {
	switch runtime.GOOS {
	case "darwin":
		// On macOS, use osascript to open a new Terminal window
		script := fmt.Sprintf(`tell application "Terminal" to do script "cd '%s' && '%s'"`, repoPath, idePath)
		cmd := exec.Command("osascript", "-e", script)
		err := cmd.Start()
		return true, err
	default:
		fmt.Printf("Only MacOS is supported for direct execution of claude. Please start manually:\n  cd \"%s\" && claude\n", repoPath)
		return false, nil
	}
}
