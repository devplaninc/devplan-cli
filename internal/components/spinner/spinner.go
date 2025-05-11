package spinner

import (
	"bufio"
	"context"
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/devplaninc/devplan-cli/internal/out"
	"io"
	"strings"
)

var (
	waitingStyle  = lipgloss.NewStyle().Margin(1, 0, 1, 1)
	successStyle  = lipgloss.NewStyle().Faint(true)
	progressStyle = lipgloss.NewStyle().MarginLeft(3).Faint(true)
)

type Spinner struct {
	model model

	p *tea.Program

	progressWriter *io.PipeWriter
	progressReader *io.PipeReader
}

func New(waitingMessage, successMessage string) *Spinner {
	sp := spinner.New()
	sp.Spinner = spinner.Points
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(out.ColorGreen))
	pr, pw := io.Pipe()
	m := model{
		spinner:        sp,
		waitingMessage: waitingMessage,
		successMessage: successMessage,
	}
	return &Spinner{
		model:          m,
		p:              tea.NewProgram(m),
		progressReader: pr,
		progressWriter: pw,
	}
}

func (s *Spinner) SendProgress(msg string) {
	s.p.Send(cloneProgressMsg(msg))
}

func (s *Spinner) GetProgressWriter() io.Writer {
	return s.progressWriter
}

func (s *Spinner) Run(ctx context.Context) error {
	go func() {
		scanner := bufio.NewScanner(s.progressReader)
		for scanner.Scan() {
			line := scanner.Text()
			s.p.Send(cloneProgressMsg(line))
		}
	}()
	go func() {
		<-ctx.Done()
		s.p.Send(done{})
		if err := s.progressWriter.Close(); err != nil {
			fmt.Printf("failed to close progress writer: %v\n", err)
		}
	}()
	m, err := s.p.Run()
	if err != nil {
		return err
	}
	if m.(model).quitting {
		return fmt.Errorf("cancelled")
	}
	return nil
}

func Run(ctx context.Context, waitingMessage, successMessage string) error {
	return New(waitingMessage, successMessage).Run(ctx)
}

type errMsg error

type done struct{}
type cloneProgressMsg string

type model struct {
	spinner  spinner.Model
	quitting bool
	err      error
	progress []string

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
	case cloneProgressMsg:
		m.progress = append(m.progress, string(msg))
		return m, nil
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
	progressMsg := ""
	if len(m.progress) > 0 {
		progressMsg = progressStyle.Render(fmt.Sprintf("%v\n", strings.Join(m.progress, "\n")))
	}
	if m.done {
		return out.Successf("%s\n%v", successStyle.Render(m.successMessage), progressMsg)
	}
	if m.quitting {
		return out.Failf("cancelled")
	}
	return waitingStyle.Render(fmt.Sprintf("%s %v\n%v", m.spinner.View(), m.waitingMessage, progressMsg))
}
