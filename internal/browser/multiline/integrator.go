package multiline

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SubmitMsg string

type model struct {
	textInput     Model
	err           error
	ready         bool
	keyboardFocus bool
}

func InitTextArea() *model {
	ti := New()
	ti.Placeholder = "sql query here"
	ti.Focus()
	return &model{
		textInput:     ti,
		err:           nil,
		keyboardFocus: true,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch tmsg := msg.(type) {
	case tea.KeyMsg:
		if tmsg.Type == tea.KeyTab {
			var tmp tea.Cmd
			if m.keyboardFocus {
				m.textInput.Blur()
			} else {
				tmp = m.textInput.Focus()
			}
			m.keyboardFocus = !m.keyboardFocus
			return m, tmp
		}
		if !m.keyboardFocus {
			return m, nil
		}
		switch tmsg.Type {
		case tea.KeyCtrlE:
			return m, func() tea.Msg {
				if statement := m.textInput.Value(); strings.Trim(statement, " ") != "" {
					return SubmitMsg(statement)
				}
				return nil
			}
		}
	case tea.WindowSizeMsg:
		m.textInput.Width = tmsg.Width - 5
		m.textInput.Height = tmsg.Height + 1
	}
	var cmd1 tea.Cmd
	m.textInput, cmd1 = m.textInput.Update(msg)
	return m, tea.Batch(cmd1)
}

func (m model) View() string {
	content := m.textInput.View()
	if !m.keyboardFocus {
		content = lipgloss.NewStyle().Faint(true).Render(content)
	}
	return content
}
