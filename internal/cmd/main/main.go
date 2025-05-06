package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/devplaninc/devplan-cli/internal/cmd"
	"os"
)

func main() {
	if len(os.Getenv("DEVPLAN_DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer func(f *os.File) {
			_ = f.Close()
		}(f)
	}

	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
