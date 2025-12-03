package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/devplaninc/devplan-cli/internal/cmd"
	"github.com/devplaninc/devplan-cli/internal/utils/logging"
)

func main() {
	if len(os.Getenv("DEVPLAN_DEBUG")) > 0 {
		fileName, err := logging.GetLogFile()
		if err == nil {
			f, err := tea.LogToFile(fileName, "debug")
			if err != nil {
				fmt.Println("fatal:", err)
				os.Exit(1)
			}
			defer func(f *os.File) {
				_ = f.Close()
			}(f)
		}

	}
	_ = logging.Setup()
	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
