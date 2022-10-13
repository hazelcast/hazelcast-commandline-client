package connection

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const spinnerText = `Trying to connect cluster %s.
   Check the logs at %s.`

var (
	currSpinner = spinner.Dot
	// TODO: get the colors from the theme
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#40826D"))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render
)

type connectionSpinnerModel struct {
	clusterName string
	logFileName string
	spinner     spinner.Model
	escaped     *bool
}

func newConnectionSpinnerModel(clusterName, logfile string, escaped *bool) *connectionSpinnerModel {
	s := spinner.New()
	s.Style = spinnerStyle
	s.Spinner = currSpinner
	return &connectionSpinnerModel{
		spinner:     s,
		clusterName: clusterName,
		logFileName: logfile,
		escaped:     escaped,
	}
}

func (m connectionSpinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m connectionSpinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	e := &EmptyDisplay{}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			*m.escaped = true
			return e, nil
		default:
			return m, nil
		}
	case Quitting:
		return e, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	default:
		return m, nil
	}
}

func (m connectionSpinnerModel) View() (s string) {
	info := fmt.Sprintf(spinnerText, m.clusterName, m.logFileName)
	s = fmt.Sprintf("\n%s %s\n", m.spinner.View(), info)
	s += helpStyle("\nCTRL+C to exit.\n")
	return
}

type Quitting struct {
}

type EmptyDisplay struct {
	quit bool
}

func (e *EmptyDisplay) Init() tea.Cmd {
	return nil
}

func (e *EmptyDisplay) Update(_ tea.Msg) (tea.Model, tea.Cmd) {
	if e.quit {
		return e, tea.Quit
	}
	return e, nil
}

func (e *EmptyDisplay) View() string {
	e.quit = true
	return ""
}
