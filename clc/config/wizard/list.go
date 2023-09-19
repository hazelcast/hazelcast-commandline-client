package wizard

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
)

const listHeight = 14

var (
	itemStyle         = lipgloss.NewStyle()
	selectedItemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00E1E1"))
)

type item string

func (i item) FilterValue() string {
	return ""
}

type itemDelegate struct{}

func (d itemDelegate) Height() int {
	return 1
}

func (d itemDelegate) Spacing() int {
	return 0
}

func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	v, ok := listItem.(item)
	if !ok {
		return
	}
	var text string
	if index == m.Index() {
		text = selectedItemStyle.Render(string(v))
	} else {
		text = itemStyle.Render(string(v))
	}
	check.I2(fmt.Fprint(w, "  "+text))
}

type model struct {
	list   list.Model
	choice string
	quit   bool
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
		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i)
			}
			return m, tea.Quit
		case "ctrl+c", "esc":
			m.quit = true
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.choice != "" {
		return m.choice
	}
	if m.quit {
		return ""
	}
	return m.list.View()
}

func InitializeList(dirs []string) model {
	var items []list.Item
	for _, k := range dirs {
		items = append(items, item(k))
	}
	l := list.New(items, itemDelegate{}, 20, listHeight)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	return model{list: l}
}
