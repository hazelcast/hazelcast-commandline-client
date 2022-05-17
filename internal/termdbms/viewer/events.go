package viewer

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// HandleMouseEvents does that
func HandleMouseEvents(m *TuiModel, msg *tea.MouseMsg) {
	switch msg.Type {
	case tea.MouseWheelDown:
		if !m.UI.EditModeEnabled {
			ScrollDown(m)
		}
		break
	case tea.MouseWheelUp:
		if !m.UI.EditModeEnabled {
			ScrollUp(m)
		}
		break
	case tea.MouseLeft:
		if !m.UI.EditModeEnabled && !m.UI.FormatModeEnabled && m.GetRow() < len(m.GetColumnData()) {
			SelectOption(m)
		}
		break
	default:
		if !m.UI.RenderSelection && !m.UI.EditModeEnabled && !m.UI.FormatModeEnabled {
			m.MouseData = tea.MouseEvent(*msg)
		}
		break
	}
}

// HandleWindowSizeEvents does that
func HandleWindowSizeEvents(m *TuiModel, msg *tea.WindowSizeMsg) tea.Cmd {
	verticalMargins := HeaderHeight + FooterHeight

	if !m.Ready {
		width := msg.Width
		height := msg.Height
		m.Viewport = viewport.Model{
			Width:  width,
			Height: height - verticalMargins}

		m.ClipboardList.SetWidth(width)
		m.ClipboardList.SetHeight(height)
		TUIWidth = width
		TUIHeight = height
		m.Viewport.YPosition = HeaderHeight
		m.Viewport.HighPerformanceRendering = true
		m.Ready = true
		m.MouseData.Y = HeaderHeight

		MaxInputLength = m.Viewport.Width

		m.TableStyle = m.GetBaseStyle()
		m.SetViewSlices()
	} else {
		m.Viewport.Width = msg.Width
		m.Viewport.Height = msg.Height - verticalMargins
	}

	if m.Viewport.HighPerformanceRendering {
		return viewport.Sync(m.Viewport)
	}

	return nil
}

// HandleKeyboardEvents does that
func HandleKeyboardEvents(m *TuiModel, msg *tea.KeyMsg) tea.Cmd {
	str := msg.String()
	for k := range GlobalCommands {
		if str == k {
			return GlobalCommands[str](m)
		}
	}
	return nil
}
