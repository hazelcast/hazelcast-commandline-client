package connectioncmd

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hazelcast/hazelcast-commandline-client/config"
	"strings"
)

var (
	selectedItemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	noStyle           = lipgloss.NewStyle()
	blurStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

const (
	uncompleted = 0
	ssl         = 1
	completed   = 2
)

type InputModel struct {
	focusIndex     int
	inputs         []textinput.Model
	cursorMode     textinput.CursorMode
	quitting       bool
	state          int
	config         *config.Config
	connectionType string
}

func ViridianInput(config *config.Config) InputModel {
	m := InputModel{
		inputs:         make([]textinput.Model, 6),
		config:         config,
		connectionType: "Viridian",
		state:          uncompleted,
		cursorMode:     textinput.CursorStatic,
	}
	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.CursorStyle = selectedItemStyle
		t.SetCursorMode(textinput.CursorStatic)
		t.PromptStyle = noStyle
		t.TextStyle = noStyle
		switch i {
		case 0:
			t.Prompt = "- Cluster Name: "
			t.PromptStyle = selectedItemStyle
			t.TextStyle = selectedItemStyle
			t.Focus()
		case 1:
			t.Prompt = "- Discovery Token: "
		case 2:
			t.Prompt = "- CA Certificate Path: "
		case 3:
			t.Prompt = "- SSL Certificate Path: "
		case 4:
			t.Prompt = "- SSL Key Path: "
		case 5:
			t.Prompt = "- SSL Password: "
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		}
		m.inputs[i] = t
	}
	return m
}

func CloudInput(config *config.Config) InputModel {
	m := InputModel{
		inputs:         make([]textinput.Model, 3),
		config:         config,
		connectionType: "Cloud",
		state:          uncompleted,
		cursorMode:     textinput.CursorStatic,
	}
	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.CursorStyle = selectedItemStyle
		t.SetCursorMode(textinput.CursorStatic)
		t.PromptStyle = noStyle
		t.TextStyle = noStyle
		switch i {
		case 0:
			t.Prompt = "- Cluster Name: "
			t.PromptStyle = selectedItemStyle
			t.TextStyle = selectedItemStyle
			t.Focus()
		case 1:
			t.Prompt = "- Discovery Token: "
		case 2:
			t.Prompt = "- Setup SSL? (y/n): "
		}
		m.inputs[i] = t
	}
	return m
}

func StandaloneInput(config *config.Config, connection string) InputModel {
	m := InputModel{
		inputs:         make([]textinput.Model, 3),
		config:         config,
		connectionType: connection,
		state:          uncompleted,
		cursorMode:     textinput.CursorStatic,
	}
	for i := range m.inputs {
		t := textinput.New()
		t.CursorStyle = selectedItemStyle
		t.PromptStyle = noStyle
		t.TextStyle = noStyle
		t.SetCursorMode(textinput.CursorStatic)
		switch i {
		case 0:
			t.Prompt = "- Cluster Name: "
			t.PromptStyle = selectedItemStyle
			t.TextStyle = selectedItemStyle
			if m.connectionType == "Local" {
				t.Placeholder = "dev"
			}
			t.Focus()
		case 1:
			t.Prompt = "- Member Addresses: "
			if m.connectionType == "Local" {
				t.Placeholder = "localhost:5701"
			}
		case 2:
			t.Prompt = "- Setup SSL? (y/n): "
			if m.connectionType == "Local" {
				t.Placeholder = "n"
			}
		}
		m.inputs[i] = t
	}
	return m
}

func SSLInput(config *config.Config) InputModel {
	m := InputModel{
		inputs:         make([]textinput.Model, 4),
		config:         config,
		state:          ssl,
		connectionType: "SSL",
		cursorMode:     textinput.CursorStatic,
	}
	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.CursorStyle = selectedItemStyle
		switch i {
		case 0:
			t.Prompt = "- CA certificate path: "
			t.Focus()
			t.PromptStyle = selectedItemStyle
			t.TextStyle = selectedItemStyle
			t.SetCursorMode(textinput.CursorStatic)
		case 1:
			t.Prompt = "- SSL certificate path: "
			t.PromptStyle = noStyle
			t.TextStyle = noStyle
			t.SetCursorMode(textinput.CursorStatic)
		case 2:
			t.Prompt = "- SSL key path: "
			t.PromptStyle = noStyle
			t.TextStyle = noStyle
			t.SetCursorMode(textinput.CursorStatic)
		case 3:
			t.Prompt = "- SSL Password: "
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
			t.PromptStyle = noStyle
			t.TextStyle = noStyle
			t.SetCursorMode(textinput.CursorStatic)
		}
		m.inputs[i] = t
	}
	return m
}

func ApprovalInput() InputModel {
	m := InputModel{
		inputs:         make([]textinput.Model, 1),
		connectionType: "Approval",
		state:          completed,
		cursorMode:     textinput.CursorStatic,
	}
	t := textinput.New()
	t.CursorStyle = selectedItemStyle
	t.Prompt = "Your config file will be overwritten, do you want to continue? (y/n): "
	t.Focus()
	t.PromptStyle = noStyle
	t.TextStyle = noStyle
	t.SetCursorMode(textinput.CursorStatic)
	m.inputs[0] = t
	return m
}

func updateConfig(m *InputModel) {
	switch m.state {
	case uncompleted:
		switch m.connectionType {
		case "Viridian":
			m.config.Hazelcast.Cluster.Cloud.Enabled = true
			m.config.SSL.ServerName = "hazelcast.cloud"
			m.config.Hazelcast.Cluster.Name = m.inputs[0].Value()
			m.config.Hazelcast.Cluster.Cloud.Token = m.inputs[1].Value()
			m.config.SSL.Enabled = true
			m.config.SSL.CAPath = m.inputs[2].Value()
			m.config.SSL.CertPath = m.inputs[3].Value()
			m.config.SSL.KeyPath = m.inputs[4].Value()
			m.config.SSL.KeyPassword = m.inputs[5].Value()
			m.state = completed
		case "Cloud":
			m.config.Hazelcast.Cluster.Cloud.Enabled = true
			m.config.SSL.ServerName = "hazelcast.cloud"
			m.config.Hazelcast.Cluster.Name = m.inputs[0].Value()
			m.config.Hazelcast.Cluster.Cloud.Token = m.inputs[1].Value()
			m.config.SSL.Enabled = m.inputs[2].Value() == "y"
		case "Local", "Remote":
			m.config.Hazelcast.Cluster.Name = m.inputs[0].Value()
			addressString := strings.ReplaceAll(m.inputs[1].Value(), " ", "")
			m.config.Hazelcast.Cluster.Network.Addresses = strings.Split(addressString, ",")
			m.config.SSL.Enabled = m.inputs[2].Value() == "y"
		}
	case ssl:
		m.config.SSL.CAPath = m.inputs[0].Value()
		m.config.SSL.CertPath = m.inputs[1].Value()
		m.config.SSL.KeyPath = m.inputs[2].Value()
		m.config.SSL.KeyPassword = m.inputs[3].Value()
		m.state = completed
	case completed:
		choice = m.inputs[0].Value()
	}
}

func (m InputModel) Init() tea.Cmd {
	return nil
}

func (m InputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			return m, tea.Quit
		case tea.KeyUp:
			if m.focusIndex--; m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}
			cmd := m.updateStyles()
			return m, cmd
		case tea.KeyDown:
			if m.focusIndex++; m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			}
			cmd := m.updateStyles()
			return m, cmd
		case tea.KeyEnter:
			if m.focusIndex == len(m.inputs) || m.connectionType == "Approval" {
				updateConfig(&m)
				if m.state == completed {
					m.quitting = true
					return m, tea.Quit
				} else if m.config.SSL.Enabled {
					m = SSLInput(m.config)
				} else {
					m.quitting = true
					return m, tea.Quit
				}
			}
		}
	}
	cmd := m.updateInputs(msg)
	return m, cmd
}

func (m *InputModel) updateInputs(msg tea.Msg) tea.Cmd {
	var cmds = make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return tea.Batch(cmds...)
}

func (m *InputModel) updateStyles() tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := 0; i <= len(m.inputs)-1; i++ {
		if i == m.focusIndex {
			cmds[i] = m.inputs[i].Focus()
			m.inputs[i].PromptStyle = selectedItemStyle
			m.inputs[i].TextStyle = selectedItemStyle
			continue
		}
		m.inputs[i].Blur()
		m.inputs[i].PromptStyle = noStyle
		m.inputs[i].TextStyle = noStyle
	}
	return tea.Batch(cmds...)
}

func (m InputModel) View() string {
	if m.quitting {
		return ""
	} else {
		var b strings.Builder
		switch m.connectionType {
		case "Viridian":
			fmt.Fprintf(&b, "%s\n", fmt.Sprintf("%s",
				noStyle.Render("Please provide your Hazelcast Viridian tokens below.")))
		case "Local":
			fmt.Fprintf(&b, "%s\n", fmt.Sprintf("%s",
				noStyle.Render("Please provide IP address and port number of your local cluster.")))
		case "Remote":
			fmt.Fprintf(&b, "%s\n", fmt.Sprintf("%s",
				noStyle.Render("Please provide IP address and port number of your remote cluster.")))
		case "Cloud":
			fmt.Fprintf(&b, "%s\n", fmt.Sprintf("%s",
				noStyle.Render("Please provide your Hazelcast Cloud tokens below.")))
		case "SSL":
			fmt.Fprintf(&b, "%s\n", fmt.Sprintf("%s",
				noStyle.Render("Please provide paths to your SSL certificates below.")))
		}
		for i := range m.inputs {
			b.WriteString(m.inputs[i].View())
			if i < len(m.inputs)-1 {
				b.WriteRune('\n')
			}
		}
		if m.connectionType != "Approval" {
			button := fmt.Sprintf("%s", noStyle.Render("[ Submit ]"))
			if m.focusIndex == len(m.inputs) {
				button = fmt.Sprintf("%s", selectedItemStyle.Copy().Render("[ Submit ]"))
			}
			fmt.Fprintf(&b, "\n\n%s", button)
		}
		return b.String()
	}
}
