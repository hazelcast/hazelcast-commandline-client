package viewer

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/hazelcast/hazelcast-commandline-client/internal/tuiutil"
)

var (
	HeaderHeight   = 3
	FooterHeight   = 1
	MaxInputLength int
	HeaderStyle    lipgloss.Style
	FooterStyle    lipgloss.Style
)

func (m *TuiModel) Data() *UIData {
	if m.QueryData != nil {
		return m.QueryData
	}

	return &m.DefaultData
}

func (m *TuiModel) Table() *TableState {
	if m.QueryResult != nil {
		return m.QueryResult
	}

	return &m.DefaultTable
}

func SetStyles() {
	HeaderStyle = lipgloss.NewStyle()
	FooterStyle = lipgloss.NewStyle()
	if !tuiutil.Ascii {
		HeaderStyle = HeaderStyle.
			Foreground(tuiutil.HeaderForeground()).Background(tuiutil.HeaderBackground())
		FooterStyle = FooterStyle.
			Foreground(tuiutil.FooterForeground())
	}
}

// INIT UPDATE AND RENDER

// Init currently doesn't do anything but necessary for interface adherence
func (m TuiModel) Init() tea.Cmd {
	SetStyles()
	return nil
}

// Update is where all commands and whatnot get processed
func (m TuiModel) Update(message tea.Msg) (TuiModel, tea.Cmd) {
	var (
		command  tea.Cmd
		commands []tea.Cmd
	)
	if !m.UI.FormatModeEnabled {
		m.Viewport, _ = m.Viewport.Update(message)
	}
	switch msg := message.(type) {
	case tea.MouseMsg:
		HandleMouseEvents(&m, &msg)
		m.SetViewSlices()
		break
	case tea.WindowSizeMsg:
		event := HandleWindowSizeEvents(&m, &msg)
		if event != nil {
			commands = append(commands, event)
		}
		break
	case tea.KeyMsg:
		// when fullscreen selection viewing is in session, don't allow UI manipulation other than quit or exit
		s := msg.String()
		invalidRenderCommand := m.UI.RenderSelection &&
			s != "esc" &&
			s != "ctrl+c" &&
			s != "q" &&
			s != "p" &&
			s != "m" &&
			s != "n"
		if invalidRenderCommand {
			break
		}
		if s == "ctrl+c" || (s == "q" && (!m.UI.EditModeEnabled && !m.UI.FormatModeEnabled)) {
			return m, tea.Quit
		}
		event := HandleKeyboardEvents(&m, &msg)
		if event != nil {
			commands = append(commands, event)
		}
		if !m.UI.EditModeEnabled && m.Ready {
			m.SetViewSlices()
			if m.UI.FormatModeEnabled {
				MoveCursorWithinBounds(&m)
			}
		}
		break
	case error:
		return m, nil
	}
	if m.Viewport.HighPerformanceRendering {
		commands = append(commands, command)
	}
	return m, tea.Batch(commands...)
}
