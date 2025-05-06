package out

import "github.com/charmbracelet/lipgloss"

var (
	highlightStyle = lipgloss.NewStyle().Bold(true)
	errorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	checkStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	Check          = checkStyle.Render("✓")
)

func Highlight(s string) string {
	return highlightStyle.Render(s)
}

func Error(s string) string {
	return errorStyle.Render(s)
}
