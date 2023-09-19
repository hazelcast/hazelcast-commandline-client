package wizard

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00E1E1"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	noStyle      = lipgloss.NewStyle()

	focusedButton = focusedStyle.Copy().Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
)

type textModel struct {
	focusIndex int
	quitting   bool
	choice     string
	inputs     []textinput.Model
}

func InitialModel() textModel {
	m := textModel{
		inputs:   make([]textinput.Model, 2),
		quitting: false,
		choice:   "",
	}
	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		switch i {
		case 0:
			t.Prompt = "Configuration Name: "
			t.Placeholder = "default"
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
			t.Focus()
		case 1:
			t.Prompt = "Source: "
			t.Placeholder = ""
		}
		m.inputs[i] = t
	}
	return m
}

func (m textModel) Init() tea.Cmd {
	return nil
}

func (m textModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			return m, tea.Quit
		case tea.KeyTab, tea.KeyShiftTab, tea.KeyEnter, tea.KeyUp, tea.KeyDown:
			s := msg.String()
			if s == "enter" && m.focusIndex == len(m.inputs) {
				m.quitting = true
				if m.inputs[0].Value() == "" {
					m.inputs[0].SetValue("default")
				}
				m.choice = "enter"
				return m, tea.Quit
			}
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}
			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}
			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = focusedStyle
					m.inputs[i].TextStyle = focusedStyle
					continue
				}
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = noStyle
				m.inputs[i].TextStyle = noStyle
			}
			return m, tea.Batch(cmds...)
		}
	}
	cmd := m.updateInputs(msg)
	return m, cmd
}

func (m *textModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return tea.Batch(cmds...)
}

func (m textModel) GetInputs() []string {
	return []string{m.inputs[0].Value(), m.inputs[1].Value()}
}

func (m textModel) View() string {
	if m.choice != "" {
		return m.choice
	}
	if m.quitting {
		return ""
	}
	var b strings.Builder
	b.WriteString(`There is no configuration detected.

This screen helps you create a new connection configuration.
Note that this screen supports only Viridian clusters.
For other clusters use the following command:
	
	clc config add --help

1. Enter the desired name in the "Configuration Name" field. 
2. On Viridian console, visit:
	
	Dashboard -> Connect Client -> CLI

3. Copy the URL in second box and pass it to "Source" field.
4. Navigate to the [Submit] button and press enter.
	
Alternatively, you can use the following command:
	
	clc config import --help

`)
	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}
	button := &blurredButton
	if m.focusIndex == len(m.inputs) {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n", *button)
	return b.String()
}
