package vertical

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	components  []tea.Model
	margins     []int
	total       int
	lastSizeMsg *tea.WindowSizeMsg
}

func InitialModel(components []tea.Model, margins []int) model {
	var count int
	for _, m := range margins {
		if m > 0 {
			count += m
		}
	}
	return model{components: components, margins: margins, total: count}
}

func (m model) Init() tea.Cmd {
	var cmds []tea.Cmd
	for _, c := range m.components {
		cmds = append(cmds, c.Init())
	}
	return tea.Batch(cmds...)
}

func (l model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		cmds := l.updateComponentsWithNewSize(m)
		return l, tea.Batch(cmds...)
	case tea.KeyMsg:
		if k := m.String(); k == "ctrl+q" {
			return l, tea.Quit
		}
	}
	var cmds []tea.Cmd
	for i, c := range l.components {
		var cmd tea.Cmd
		l.components[i], cmd = c.Update(msg)
		cmds = append(cmds, cmd)
	}
	return l, tea.Batch(cmds...)
}

func (l *model) updateComponentsWithNewSize(wsm tea.WindowSizeMsg) []tea.Cmd {
	var staticHeight int
	for i, c := range l.components {
		if l.margins[i] <= 0 {
			staticHeight += lipgloss.Height(c.View())
		}
	}
	screenShareUnit := float32(wsm.Height-staticHeight) / float32(l.total)
	var cmds []tea.Cmd
	for i, c := range l.components {
		var cmd tea.Cmd
		l.components[i], cmd = c.Update(tea.WindowSizeMsg{
			Width:  wsm.Width,
			Height: int(screenShareUnit * float32(l.margins[i])),
		})
		cmds = append(cmds, cmd)
	}
	return cmds
}

func (m model) View() string {
	var content []string
	for _, c := range m.components {
		content = append(content, c.View())
	}
	return lipgloss.JoinVertical(lipgloss.Left, content...)
}
