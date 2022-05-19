package multiline

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SubmitMsg string

// to create a distance between right most border of text box and terminal border
const multilineTextBoxRightPadding = 5

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
			var teaCmd tea.Cmd
			if m.keyboardFocus {
				m.textInput.Blur()
			} else {
				teaCmd = m.textInput.Focus()
			}
			m.keyboardFocus = !m.keyboardFocus
			return m, teaCmd
		}
		if !m.keyboardFocus {
			return m, nil
		}
		switch tmsg.Type {
		case tea.KeyCtrlE:
			submitQueryCmd := func() tea.Msg {
				if statement := m.textInput.Value(); strings.Trim(statement, " ") != "" {
					return SubmitMsg(statement)
				}
				return nil
			}
			return m, submitQueryCmd
		}
	case tea.WindowSizeMsg:
		m.textInput.Width = tmsg.Width - multilineTextBoxRightPadding
		m.textInput.Height = tmsg.Height
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
