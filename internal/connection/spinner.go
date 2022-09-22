package connection

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const spinnerText = `Connecting to the cluster %s at %s.
   Check the logs at %s.`

var (
	currSpinner  = spinner.Dot
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#40826D"))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render
)

type connectionSpinnerModel struct {
	clusterName string
	address     string
	logFileName string
	escapedCh   chan<- bool
	spinner     spinner.Model
}

func newConnectionSpinnerModel(clusterName, address, filename string, escapedCh chan<- bool) *connectionSpinnerModel {
	m := &connectionSpinnerModel{}
	m.spinner = spinner.New()
	m.spinner.Style = spinnerStyle
	m.spinner.Spinner = currSpinner
	m.clusterName = clusterName
	m.address = address
	m.logFileName = filename
	m.escapedCh = escapedCh
	return m
}

func (m connectionSpinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m connectionSpinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.escapedCh <- true
			return m, tea.Quit
		default:
			return m, nil
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	default:
		return m, nil
	}
}

func (m connectionSpinnerModel) View() (s string) {
	info := fmt.Sprintf(spinnerText, m.clusterName, m.address, m.logFileName)
	s = fmt.Sprintf("\n%s %s\n", m.spinner.View(), info)
	s += helpStyle("\nCTRL+C to exit.\n")
	return
}
