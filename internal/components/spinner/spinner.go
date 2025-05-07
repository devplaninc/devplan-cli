package spinner

import (
	"context"
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/devplaninc/devplan-cli/internal/out"
)

var (
	waitingStyle = lipgloss.NewStyle().Margin(1, 0, 1, 1)
	successStyle = lipgloss.NewStyle().Faint(true)
)

func Run(ctx context.Context, waitingMessage, successMessage string) error {
	s := spinner.New()
	s.Spinner = spinner.Points
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(out.ColorGreen))
	p := tea.NewProgram(model{
		spinner:        s,
		waitingMessage: waitingMessage,
		successMessage: successMessage,
	})
	go func() {
		<-ctx.Done()
		p.Send(done{})
	}()

	m, err := p.Run()

	if err != nil {
		return err
	}
	if m.(model).quitting {
		return fmt.Errorf("cancelled")
	}
	return nil
}

type Spinner struct {
	model model
}

type errMsg error

type done struct{}

type model struct {
	spinner  spinner.Model
	quitting bool
	err      error

	done bool

	waitingMessage string
	successMessage string
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		default:
			return m, nil
		}

	case errMsg:
		m.err = msg
		return m, nil

	case done:
		m.done = true
		return m, tea.Quit

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m model) View() string {
	if m.err != nil {
		return out.Fail(m.err.Error())
	}
	if m.done {
		return out.Successf("%s\n", successStyle.Render(m.successMessage))
	}
	if m.quitting {
		return out.Failf("cancelled")
	}
	return waitingStyle.Render(fmt.Sprintf("%s %v", m.spinner.View(), m.waitingMessage))
}
