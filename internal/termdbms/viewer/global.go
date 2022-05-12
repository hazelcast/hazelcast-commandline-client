package viewer

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/hazelcast/hazelcast-commandline-client/internal/termdbms/tuiutil"
)

type Command func(m *TuiModel) tea.Cmd

var (
	GlobalCommands = make(map[string]Command)
)

func init() {
	// GLOBAL COMMANDS
	GlobalCommands["t"] = func(m *TuiModel) tea.Cmd {
		tuiutil.SelectedTheme = (tuiutil.SelectedTheme + 1) % len(tuiutil.ValidThemes)
		SetStyles()
		return nil
	}
	GlobalCommands["pgdown"] = func(m *TuiModel) tea.Cmd {
		for i := 0; i < m.Viewport.Height; i++ {
			ScrollDown(m)
		}

		return nil
	}
	GlobalCommands["pgup"] = func(m *TuiModel) tea.Cmd {
		for i := 0; i < m.Viewport.Height; i++ {
			ScrollUp(m)
		}

		return nil
	}
	GlobalCommands["c"] = func(m *TuiModel) tea.Cmd {
		ToggleColumn(m)

		return nil
	}
	GlobalCommands["b"] = func(m *TuiModel) tea.Cmd {
		m.UI.BorderToggle = !m.UI.BorderToggle

		return nil
	}
	GlobalCommands["up"] = func(m *TuiModel) tea.Cmd {
		if m.UI.CurrentTable == len(m.Data().TableIndexMap) {
			m.UI.CurrentTable = 1
		} else {
			m.UI.CurrentTable++
		}

		// fix spacing and whatnot
		m.TableStyle = m.TableStyle.Width(m.CellWidth())
		m.MouseData.Y = HeaderHeight
		m.MouseData.X = 0
		m.Viewport.YOffset = 0
		m.Scroll.ScrollXOffset = 0

		return nil
	}
	GlobalCommands["down"] = func(m *TuiModel) tea.Cmd {
		if m.UI.CurrentTable == 1 {
			m.UI.CurrentTable = len(m.Data().TableIndexMap)
		} else {
			m.UI.CurrentTable--
		}

		// fix spacing and whatnot
		m.TableStyle = m.TableStyle.Width(m.CellWidth())
		m.MouseData.Y = HeaderHeight
		m.MouseData.X = 0
		m.Viewport.YOffset = 0
		m.Scroll.ScrollXOffset = 0

		return nil
	}
	GlobalCommands["right"] = func(m *TuiModel) tea.Cmd {
		headers := m.GetHeaders()
		headersLen := len(headers)
		if headersLen > maxHeaders && m.Scroll.ScrollXOffset <= headersLen-maxHeaders {
			m.Scroll.ScrollXOffset++
		}

		return nil
	}
	GlobalCommands["left"] = func(m *TuiModel) tea.Cmd {
		if m.Scroll.ScrollXOffset > 0 {
			m.Scroll.ScrollXOffset--
		}

		return nil
	}
	GlobalCommands["s"] = func(m *TuiModel) tea.Cmd {
		max := len(m.GetSchemaData()[m.GetHeaders()[m.GetColumn()]])

		if m.MouseData.Y-HeaderHeight+m.Viewport.YOffset < max-1 {
			m.MouseData.Y++
			ceiling := m.Viewport.Height + HeaderHeight - 1
			tuiutil.Clamp(m.MouseData.Y, m.MouseData.Y+1, ceiling)
			if m.MouseData.Y > ceiling {
				ScrollDown(m)
				m.MouseData.Y = ceiling
			}
		}

		return nil
	}
	GlobalCommands["w"] = func(m *TuiModel) tea.Cmd {
		pre := m.MouseData.Y
		if m.Viewport.YOffset > 0 && m.MouseData.Y == HeaderHeight {
			ScrollUp(m)
			m.MouseData.Y = pre
		} else if m.MouseData.Y > HeaderHeight {
			m.MouseData.Y--
		}

		return nil
	}
	GlobalCommands["d"] = func(m *TuiModel) tea.Cmd {
		cw := m.CellWidth()
		col := m.GetColumn()
		cols := len(m.Data().TableHeadersSlice) - 1
		if (m.MouseData.X-m.Viewport.Width) <= cw && m.GetColumn() < cols { // within tolerances
			m.MouseData.X += cw
		} else if col == cols {
			return func() tea.Msg {
				return tea.KeyMsg{
					Type: tea.KeyRight,
					Alt:  false,
				}
			}
		}

		return nil
	}
	GlobalCommands["a"] = func(m *TuiModel) tea.Cmd {
		cw := m.CellWidth()
		if m.MouseData.X-cw >= 0 {
			m.MouseData.X -= cw
		} else if m.GetColumn() == 0 {
			return func() tea.Msg {
				return tea.KeyMsg{
					Type: tea.KeyLeft,
					Alt:  false,
				}
			}
		}
		return nil
	}
	GlobalCommands["k"] = GlobalCommands["up"]    // dual bind of up/k
	GlobalCommands["j"] = GlobalCommands["down"]  // dual bind of down/j
	GlobalCommands["l"] = GlobalCommands["right"] // dual bind of right/l
	GlobalCommands["h"] = GlobalCommands["left"]  // dual bind of left/h
	GlobalCommands["m"] = func(m *TuiModel) tea.Cmd {
		ScrollUp(m)
		return nil
	}
	GlobalCommands["n"] = func(m *TuiModel) tea.Cmd {
		ScrollDown(m)
		return nil
	}
}
