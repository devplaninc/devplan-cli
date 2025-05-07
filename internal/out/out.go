package out

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
)

var (
	highlightStyle = lipgloss.NewStyle().Bold(true)
	errorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	checkStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	failStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	Check          = checkStyle.Render("✓")
	Fail           = failStyle.Render("✗")
)

const (
	ColorGreen = "10"
)

func Highlight(s string) string {
	return highlightStyle.Render(s)
}

func Highlightf(format string, a ...any) string {
	return highlightStyle.Render(fmt.Sprintf(format, a...))
}

func Psuccessf(format string, a ...any) {
	fmt.Print(Successf(format, a...))
}

func Successf(format string, a ...any) string {
	return Check + lipgloss.NewStyle().MarginLeft(1).Render(fmt.Sprintf(format, a...))
}

func Failf(format string, a ...any) string {
	return Fail + lipgloss.NewStyle().MarginLeft(1).Render(fmt.Sprintf(format, a...))
}

func Error(s string) string {
	return Fail + " " + errorStyle.Render(s)
}
