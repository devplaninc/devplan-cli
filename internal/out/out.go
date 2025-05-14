package out

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/manifoldco/promptui"
)

var (
	highlightStyle = lipgloss.NewStyle().Bold(true)
	Check          = lipgloss.NewStyle().Bold(true).Render(promptui.IconGood)
	Cross          = lipgloss.NewStyle().Bold(true).Render(promptui.IconBad)
	WarnIcon       = lipgloss.NewStyle().Bold(true).Render(promptui.IconWarn)
)

const (
	ColorGreen = "#009900"
)

// H stands for highlight
func H(v any) string {
	return Hf("%v", v)
}

// Hf highlights with formatting
func Hf(format string, a ...any) string {
	return highlightStyle.Render(fmt.Sprintf(format, a...))
}

func Psuccessf(format string, a ...any) {
	fmt.Print(Successf(format, a...))
}

func Successf(format string, a ...any) string {
	return Check + " " + lipgloss.NewStyle().Render(fmt.Sprintf(format, a...))
}

func Failf(format string, a ...any) string {
	return Cross + " " + lipgloss.NewStyle().Render(fmt.Sprintf(format, a...))
}

func Pfailf(format string, a ...any) {
	fmt.Print(Failf(format, a...))
}

func Warnf(format string, a ...interface{}) string {
	return WarnIcon + " " + fmt.Sprintf(format, a...)
}

func Fail(v any) string {
	return Failf("%v", v)
}

func Pfail(s string) {
	fmt.Print(Fail(s))
}
