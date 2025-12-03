package ide

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/devplaninc/devplan-cli/internal/utils/prefs"
	"github.com/opensdd/osdd-core/core/executable"
)

// LaunchIDE launches the specified IDE at the given repository path
func LaunchIDE(ide IDE, repoPath string, start bool) (bool, error) {
	installed, err := DetectInstalledIDEs()
	if err != nil {
		return false, fmt.Errorf("failed to detect installed IDEs: %w", err)
	}
	return LaunchWithPath(installed[ide], repoPath, start)
}

// LaunchWithPath launches the specified IDE at the given repository path
func LaunchWithPath(idePath string, repoPath string, start bool) (bool, error) {
	if _, err := os.Stat(idePath); os.IsNotExist(err) {
		return false, fmt.Errorf("IDE executable not found at %s", idePath)
	}

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return false, fmt.Errorf("repository path not found at %s", repoPath)
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin", "linux", "windows":
		if isTerminalExecutable(idePath) {
			return launchInTerminal(idePath, repoPath, start)
		}
		cmd = exec.Command(idePath, repoPath)
	default:
		return false, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return true, cmd.Start()
}

func isTerminalExecutable(idePath string) bool {
	execName := strings.ToLower(filepath.Base(idePath))
	return strings.Contains(execName, "claude") || strings.Contains(execName, "cursor-agent")
}

// launchInTerminal launches CLI IDE in a new terminal session
func launchInTerminal(idePath string, repoPath string, start bool) (bool, error) {
	switch runtime.GOOS {
	case "darwin":
		// On macOS, use osascript to open a new Terminal window
		instructions := ""
		if start {
			instructions = ` \"Execute current feature\"`
		}
		extra := getExtraLaunchParams(idePath)
		script := fmt.Sprintf(
			`tell application "Terminal" to do script "cd '%s' && '%s'%v %s"`, repoPath, idePath, extra, instructions)
		cmd := exec.Command("osascript", "-e", script)
		err := cmd.Start()
		return true, err
	default:
		fmt.Printf("Only MacOS is supported for direct execution of CLIs. Please start manually:\n  cd \"%s\" && %v\n", repoPath, idePath)
		return false, nil
	}
}

func getExtraLaunchParams(idePath string) string {
	if strings.Contains(strings.ToLower(idePath), "cursor-agent") {
		return " -f"
	}
	return ""
}

func PathWithTilde(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	pref := fmt.Sprintf("%s/", home)
	if strings.HasPrefix(path, pref) {
		return "~/" + strings.TrimPrefix(path, pref)
	}
	return path
}

func WriteLaunchResult(res executable.LaunchResult) error {
	if cmd := res.ToExecute; cmd != "" {
		return WriteLaunchCmd(cmd)
	}
	return nil
}

func WriteLaunchCmd(cmd string) error {
	if cmd == "" || prefs.InstructionFile == "" {
		return nil
	}
	dir := filepath.Dir(prefs.InstructionFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	content := fmt.Sprintf("exec=%v\n", cmd)
	return os.WriteFile(prefs.InstructionFile, []byte(content), 0644)
}
