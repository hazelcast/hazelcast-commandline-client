package viewer

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/hazelcast/hazelcast-commandline-client/internal/termdbms/tuiutil"
)

type TableAssembly func(m *TuiModel, s *string, c *chan bool)

var (
	HeaderAssembly func(m *TuiModel, s *string)
	FooterAssembly TableAssembly
	Message        string
	mid            *string
	MIP            bool
)

func init() {
	tmp := ""
	MIP = false
	mid = &tmp
	HeaderAssembly = func(m *TuiModel, s *string) {
		if m.UI.ShowClipboard {
			return
		}

		var (
			builder []string
		)

		bs := m.GetBaseStyle()
		style := bs.UnsetForeground().UnsetFaint().Underline(true).Bold(true)
		headers := m.Data().TableHeadersSlice
		for _, d := range headers {
			// write all headers
			text := " " + TruncateIfApplicable(m, d)
			builder = append(builder, style.
				Render(text))
		}
		{
			// schema name
			var headerTop string
			if m.UI.EditModeEnabled || m.UI.FormatModeEnabled {
				headerTop = m.TextInput.Model.View()
				if !m.TextInput.Model.Focused() {
					headerTop = HeaderStyle.Copy().Faint(true).Render(headerTop)
				}
			} else {
				headerTop = fmt.Sprintf(" %s (%d/%d) - %d record(s) + %d column(s)",
					m.GetSchemaName(),
					m.UI.CurrentTable,
					len(m.Data().TableHeaders), // look at how headers get rendered to get accurate record number
					len(m.GetColumnData()),
					len(m.GetHeaders())) // this will need to be refactored when filters get added
				navigationArrowL := lipgloss.Width("  <<<")
				titleWidth := m.Viewport.Width - navigationArrowL*2
				headerTop = "  <<<" + fmt.Sprintf("%*s", -titleWidth, fmt.Sprintf("%*s", (titleWidth+len(headerTop))/2, headerTop)) + ">>>  "
				headerStyle := HeaderStyle.Copy().Foreground(lipgloss.Color(tuiutil.Highlight())).Reverse(true)
				headerTop = headerStyle.Copy().Render(headerTop)
			}
			headerMid := lipgloss.JoinHorizontal(lipgloss.Left, builder...)
			if m.UI.RenderSelection {
				headerMid = ""
			}
			x := HeaderStyle.Copy().Foreground(lipgloss.Color(tuiutil.Highlight())).Reverse(true).Width(m.Viewport.Width).Render(" ")
			*s = lipgloss.JoinVertical(lipgloss.Left, x, headerTop, x, headerMid)
		}
	}
	FooterAssembly = func(m *TuiModel, s *string, done *chan bool) {
		if m.UI.ShowClipboard {
			*done <- true
			return
		}
		var (
			row int
			col int
		)
		if !m.UI.FormatModeEnabled { // reason we flip is because it makes more sense to store things by column for data
			row = m.GetRow() + m.Viewport.YOffset
			col = m.GetColumn() + m.Scroll.ScrollXOffset
		} else { // but for format mode thats just a regular row/col situation
			row = m.Format.CursorX
			col = m.Format.CursorY + m.Viewport.YOffset
		}
		footer := fmt.Sprintf(" %d, %d ", row, col)
		if m.UI.RenderSelection {
			footer = ""
		}
		undoRedoInfo := ""
		gapSize := m.Viewport.Width - lipgloss.Width(footer) - lipgloss.Width(undoRedoInfo) - 2
		if MIP {
			MIP = false
			if !tuiutil.Ascii {
				Message = FooterStyle.Render(Message)
			}
			go func() {
				newSize := gapSize - lipgloss.Width(Message)
				if newSize < 1 {
					newSize = 1
				}
				half := strings.Repeat("-", newSize/2)
				if lipgloss.Width(Message) > gapSize {
					Message = Message[0:gapSize-3] + "..."
				}
				*mid = half + Message + half
				time.Sleep(time.Second * 5)
				Message = ""
				go Program.Send(tea.KeyMsg{})
			}()
		} else if Message == "" {
			*mid = strings.Repeat("-", gapSize)
		}
		queryResultsFlag := "├"
		if m.QueryData != nil || m.QueryResult != nil {
			queryResultsFlag = "*"
		}
		footer = FooterStyle.Render(undoRedoInfo) + queryResultsFlag + *mid + "┤" + FooterStyle.Render(footer)
		*s = footer

		*done <- true
	}
}
