package selector

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/devplaninc/devplan-cli/internal/out"
	"io"
	"strings"
)

type Props struct {
	Optional bool
}

const listHeight = 12

var (
	titleStyle        = lipgloss.NewStyle().Bold(true)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(3)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(0).Underline(true).Bold(true)
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
)

type item struct {
	id    string
	title string
	extra string

	showID bool
}

func (i item) Title() string       { return i.title }
func (i item) FilterValue() string { return i.title }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}
	str := i.title
	if i.showID {
		str = fmt.Sprintf("%s [%s]", str, i.id)
	}
	if i.extra != "" {
		str = fmt.Sprintf("%s %s", str, i.extra)
	}
	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return " > " + selectedItemStyle.Render(strings.Join(s, " "))
		}
	}
	_, _ = fmt.Fprint(w, fn(str))
}

type model struct {
	list     list.Model
	choice   item
	quitting bool
	title    string
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = i
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.choice.id != "" {
		return out.Successf("Selected %s: %s.\n", m.title, out.H(m.choice.title))
	}
	if m.quitting {
		return out.Failf("No %s selected\n", m.title)
	}
	return "\n" + m.list.View()
}

func runSelector(title string, items []item, props Props, defaultIndex ...int) (string, error) {
	listItems := make([]list.Item, len(items))
	for i, it := range items {
		listItems[i] = it
	}
	const defaultWidth = 20
	l := list.New(listItems, itemDelegate{}, defaultWidth, listHeight)
	l.Title = fmt.Sprintf("Select %v", title)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	// Set default selection if provided
	if len(defaultIndex) > 0 && defaultIndex[0] >= 0 && defaultIndex[0] < len(items) {
		l.Select(defaultIndex[0])
	}

	m := model{list: l, title: title}

	res, err := tea.NewProgram(m).Run()
	if err != nil {
		return "", fmt.Errorf("failed to select %s: %w", title, err)
	}
	resModel := res.(model)
	selected := resModel.list.SelectedItem()
	if !props.Optional && (selected == nil || selected.(item).id == "" || resModel.quitting) {
		return "", fmt.Errorf("%v is required but not selected", title)
	}
	if selected == nil || resModel.quitting {
		return "", nil
	}
	return selected.(item).id, nil
}
