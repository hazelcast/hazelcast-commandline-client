package connectioncmd

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hazelcast/hazelcast-commandline-client/config"
	"os"
	"strings"
)

var (
	selectedItemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	noStyle           = lipgloss.NewStyle()

	focusedButton = selectedItemStyle.Copy().Render("[ Submit ]")
	blurredButton = fmt.Sprintf("%s", noStyle.Render("[ Submit ]"))
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
		connectionType: "Hazelcast Viridian",
		state:          0,
		cursorMode:     textinput.CursorStatic,
	}
	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.CursorStyle = selectedItemStyle
		switch i {
		case 0:
			t.Prompt = "- Cluster name: "
			t.Focus()
			t.PromptStyle = selectedItemStyle
			t.TextStyle = selectedItemStyle
			t.SetCursorMode(textinput.CursorStatic)
		case 1:
			t.Prompt = "- Discovery token: "
			t.PromptStyle = noStyle
			t.TextStyle = noStyle
			t.SetCursorMode(textinput.CursorStatic)
		case 2:
			t.Prompt = "- CA certificate path: "
			t.PromptStyle = noStyle
			t.TextStyle = noStyle
			t.SetCursorMode(textinput.CursorStatic)
		case 3:
			t.Prompt = "- SSL certificate path: "
			t.PromptStyle = noStyle
			t.TextStyle = noStyle
			t.SetCursorMode(textinput.CursorStatic)
		case 4:
			t.Prompt = "- SSL key path: "
			t.PromptStyle = noStyle
			t.TextStyle = noStyle
			t.SetCursorMode(textinput.CursorStatic)
		case 5:
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

func CloudInput(config *config.Config) InputModel {
	m := InputModel{
		inputs:         make([]textinput.Model, 3),
		config:         config,
		connectionType: "Hazelcast Cloud",
		state:          0,
		cursorMode:     textinput.CursorStatic,
	}
	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.CursorStyle = selectedItemStyle
		switch i {
		case 0:
			t.Prompt = "- Cluster name: "
			t.Focus()
			t.PromptStyle = selectedItemStyle
			t.TextStyle = selectedItemStyle
			t.SetCursorMode(textinput.CursorStatic)
		case 1:
			t.Prompt = "- Discovery token: "
			t.PromptStyle = noStyle
			t.TextStyle = noStyle
			t.SetCursorMode(textinput.CursorStatic)
		case 2:
			t.Prompt = "- SSL enabled: "
			t.PromptStyle = noStyle
			t.TextStyle = noStyle
		}
		m.inputs[i] = t
	}
	return m
}

func LocalInput(config *config.Config) InputModel {
	m := InputModel{
		inputs:         make([]textinput.Model, 4),
		config:         config,
		connectionType: "Local",
		state:          0,
		cursorMode:     textinput.CursorStatic,
	}
	for i := range m.inputs {
		t := textinput.New()
		t.CursorStyle = selectedItemStyle
		t.SetCursorMode(textinput.CursorStatic)
		t.CharLimit = 64
		switch i {
		case 0:
			t.Prompt = "- Cluster name: "
			t.Focus()
			t.Placeholder = "dev"
			t.PromptStyle = selectedItemStyle
			t.TextStyle = selectedItemStyle
		case 1:
			t.Prompt = "- IP address: "
			t.PromptStyle = noStyle
			t.Placeholder = "127.0.0.1"
			t.TextStyle = noStyle
		case 2:
			t.Prompt = "- Port range: "
			t.Placeholder = "5700-5800"
			t.PromptStyle = noStyle
			t.TextStyle = noStyle
		case 3:
			t.Prompt = "- SSL enabled: "
			t.Placeholder = "false"
			t.PromptStyle = noStyle
			t.TextStyle = noStyle
		}
		m.inputs[i] = t
	}
	return m
}

func RemoteInput(config *config.Config) InputModel {
	m := InputModel{
		inputs:         make([]textinput.Model, 4),
		connectionType: "Remote",
		config:         config,
		state:          0,
		cursorMode:     textinput.CursorStatic,
	}
	for i := range m.inputs {
		t := textinput.New()
		t.CursorStyle = selectedItemStyle
		t.SetCursorMode(textinput.CursorStatic)
		switch i {
		case 0:
			t.Prompt = "- Cluster name: "
			t.Focus()
			t.PromptStyle = selectedItemStyle
			t.TextStyle = selectedItemStyle
		case 1:
			t.Prompt = "- IP address: "
			t.PromptStyle = noStyle
			t.TextStyle = noStyle
		case 2:
			t.Prompt = "- Port number: "
			t.PromptStyle = noStyle
			t.TextStyle = noStyle
		case 3:
			t.Prompt = "- SSL enabled: "
			t.Placeholder = "false"
			t.PromptStyle = noStyle
			t.TextStyle = noStyle
		}
		m.inputs[i] = t
	}
	return m
}

func SSLInput(config *config.Config, connectionType string) InputModel {
	m := InputModel{
		inputs:         make([]textinput.Model, 4),
		config:         config,
		connectionType: connectionType,
		state:          2,
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
		state:          1,
		cursorMode:     textinput.CursorStatic,
	}
	t := textinput.New()
	t.CursorStyle = selectedItemStyle
	t.Prompt = "Your config file will be overwritten, are you want to continue? (y/n): "
	t.Focus()
	t.PromptStyle = noStyle
	t.TextStyle = noStyle
	t.SetCursorMode(textinput.CursorStatic)
	m.inputs[0] = t
	return m
}

func updateValues(m *InputModel) {
	if m.state == 2 {
		m.config.SSL.CAPath = m.inputs[0].Value()
		m.config.SSL.CertPath = m.inputs[1].Value()
		m.config.SSL.KeyPath = m.inputs[2].Value()
		m.config.SSL.KeyPassword = m.inputs[3].Value()
		m.state = 1
	} else {
		if m.connectionType == "Hazelcast Viridian" {
			_ = os.Setenv("HZ_CLOUD_COORDINATOR_BASE_URL", "https://api.viridian.hazelcast.com")
			m.config.Hazelcast.Cluster.Cloud.Enabled = true
			m.config.Hazelcast.Cluster.Name = m.inputs[0].Value()
			m.config.Hazelcast.Cluster.Cloud.Token = m.inputs[1].Value()
			m.config.SSL.Enabled = true
			m.config.SSL.ServerName = "hazelcast.cloud"
			m.config.SSL.CAPath = m.inputs[2].Value()
			m.config.SSL.CertPath = m.inputs[3].Value()
			m.config.SSL.KeyPath = m.inputs[4].Value()
			m.config.SSL.KeyPassword = m.inputs[5].Value()
		} else if m.connectionType == "Hazelcast Cloud" {
			_ = os.Setenv("HZ_CLOUD_COORDINATOR_BASE_URL", "https://coordinator.hazelcast.cloud")
			m.config.Hazelcast.Cluster.Cloud.Enabled = true
			m.config.Hazelcast.Cluster.Name = m.inputs[0].Value()
			m.config.Hazelcast.Cluster.Cloud.Token = m.inputs[1].Value()
			if m.inputs[2].Value() == "true" {
				m.state = 2
				m.config.SSL.Enabled = true
			} else {
				m.config.SSL.Enabled = false
			}
		} else if m.connectionType == "Local" {
			m.config.Hazelcast.Cluster.Name = m.inputs[0].Value()
			m.config.Hazelcast.Cluster.Network.Addresses = append(make([]string, 0), m.inputs[1].Value())
			if m.inputs[3].Value() == "true" {
				m.state = 2
				m.config.SSL.Enabled = true
			} else {
				m.config.SSL.Enabled = false
			}
		} else if m.connectionType == "Remote" {
			m.config.Hazelcast.Cluster.Name = m.inputs[0].Value()
			m.config.Hazelcast.Cluster.Network.Addresses = append(make([]string, 0), m.inputs[1].Value())
			if m.inputs[3].Value() == "true" {
				m.state = 2
				m.config.SSL.Enabled = true
			} else {
				m.config.SSL.Enabled = false
			}
		} else if m.connectionType == "Approval" {
			choice = m.inputs[0].Value()
		}
	}
}

func (m InputModel) Init() tea.Cmd {
	return nil
}

func (m InputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "tab", "enter", "up", "down":
			s := msg.String()
			if s == "enter" && (m.focusIndex == len(m.inputs) || m.connectionType == "Approval") {
				updateValues(&m)
				if m.state == 2 {
					m = SSLInput(m.config, m.connectionType)
					cmd := m.updateInputs(msg)
					return m, cmd
				} else {
					m.quitting = true
					return m, tea.Quit
				}
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
					m.inputs[i].PromptStyle = selectedItemStyle
					m.inputs[i].TextStyle = selectedItemStyle
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

func (m *InputModel) updateInputs(msg tea.Msg) tea.Cmd {
	var cmds = make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return tea.Batch(cmds...)
}

func (m InputModel) View() string {
	if m.quitting {
		return ""
	} else {
		var b strings.Builder
		if m.connectionType == "Hazelcast Viridian" {
			fmt.Fprintf(&b, "%s\n", fmt.Sprintf("%s",
				noStyle.Render("Please provide your Hazelcast Viridian tokens below.")))
		} else if m.connectionType == "Local" {
			fmt.Fprintf(&b, "%s\n", fmt.Sprintf("%s",
				noStyle.Render("Please provide IP address and port number.")))
		} else if m.connectionType == "Remote" {
			fmt.Fprintf(&b, "%s\n", fmt.Sprintf("%s",
				noStyle.Render("Please provide your remote connection address.")))
		}
		for i := range m.inputs {
			b.WriteString(m.inputs[i].View())
			if i < len(m.inputs)-1 {
				b.WriteRune('\n')
			}
		}
		if m.connectionType != "Approval" {
			button := &blurredButton
			if m.focusIndex == len(m.inputs) {
				button = &focusedButton
			}
			fmt.Fprintf(&b, "\n\n%s\n\n", *button)
		}
		return b.String()
	}
}
