package connwizardcmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/hazelcast/hazelcast-commandline-client/config"
)

var (
	selectedItemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	noStyle           = lipgloss.NewStyle()
)

const (
	viridian         = 0
	standalone       = 1
	ssl              = 2
	approveOverwrite = 3
)

const (
	clusterNameMsg    = "• Cluster Name: "
	addressesMsg      = "• Member Addresses: "
	setupSslMsg       = "• Setup SSL? (y/n): "
	discoveryTokenMsg = "• Discovery Token: "
	caPathMsg         = "• CA Certificate Path: "
	certPathMsg       = "• SSL Certificate Path: "
	keyPathMsg        = "• SSL Key Path: "
	passwordMsg       = "• SSL Password: "

	approveOverwriteMsg = "Your config file will be overwritten, do you want to continue? (y/n): "
	submitMsg           = "[ OK ]"

	viridianInfoMsg   = "Please provide your Hazelcast Viridian tokens below."
	standaloneInfoMsg = "Please provide cluster name and address of your standalone cluster."
	sslInfoMsg        = "Please provide paths to your SSL certificates below."
)

type InputModel struct {
	focusIndex int
	inputs     []textinput.Model
	quitting   bool
	config     *config.Config
	inputType  int
	choice     string
}

func ViridianInput(config *config.Config) *InputModel {
	inputs := []textinput.Model{
		textInputWithPrompt(clusterNameMsg),
		textInputWithPrompt(discoveryTokenMsg),
		textInputWithPrompt(caPathMsg),
		textInputWithPrompt(certPathMsg),
		textInputWithPrompt(keyPathMsg),
		passwordInput(),
	}
	updateSelectedIem(&inputs[0])
	return &InputModel{
		inputs:    inputs,
		config:    config,
		inputType: viridian,
	}
}

func StandaloneInput(config *config.Config) *InputModel {
	inputs := []textinput.Model{
		textInputWithPrompt(clusterNameMsg),
		textInputWithPrompt(addressesMsg),
		textInputWithPrompt(setupSslMsg),
	}
	updateSelectedIem(&inputs[0])
	return &InputModel{
		inputs:    inputs,
		config:    config,
		inputType: standalone,
	}
}

func SSLInput(config *config.Config) *InputModel {
	m := &InputModel{
		inputs:    make([]textinput.Model, 4),
		config:    config,
		inputType: ssl,
	}
	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		switch i {
		case 0:
			t.Prompt = caPathMsg
			updateSelectedIem(&t)
		case 1:
			t.Prompt = certPathMsg
		case 2:
			t.Prompt = keyPathMsg
		case 3:
			t.Prompt = passwordMsg
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		}
		m.inputs[i] = t
	}
	return m
}

func ApprovalInput() *InputModel {
	m := &InputModel{
		inputs:    make([]textinput.Model, 1),
		inputType: approveOverwrite,
	}
	t := textinput.New()
	t.Prompt = approveOverwriteMsg
	t.CharLimit = 1
	t.Focus()
	m.inputs[0] = t
	return m
}

func (m *InputModel) Choice() string {
	return m.choice
}

func (m *InputModel) updateConfig() {
	switch m.inputType {
	case viridian:
		m.config.Hazelcast.Cluster.Cloud.Enabled = true
		m.config.SSL.ServerName = "hazelcast.cloud"
		m.config.Hazelcast.Cluster.Name = m.inputs[0].Value()
		m.config.Hazelcast.Cluster.Cloud.Token = m.inputs[1].Value()
		m.config.SSL.Enabled = true
		m.config.SSL.CAPath = m.inputs[2].Value()
		m.config.SSL.CertPath = m.inputs[3].Value()
		m.config.SSL.KeyPath = m.inputs[4].Value()
		m.config.SSL.KeyPassword = m.inputs[5].Value()
	case standalone:
		m.config.Hazelcast.Cluster.Name = m.inputs[0].Value()
		addressString := strings.ReplaceAll(m.inputs[1].Value(), " ", "")
		m.config.Hazelcast.Cluster.Network.Addresses = strings.Split(addressString, ",")
		m.config.SSL.Enabled = m.inputs[2].Value() == "y"
	case ssl:
		m.config.SSL.CAPath = m.inputs[0].Value()
		m.config.SSL.CertPath = m.inputs[1].Value()
		m.config.SSL.KeyPath = m.inputs[2].Value()
		m.config.SSL.KeyPassword = m.inputs[3].Value()
	case approveOverwrite:
		m.choice = m.inputs[0].Value()
	}
}

func (m *InputModel) Init() tea.Cmd {
	return nil
}

func (m *InputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			if m.inputType == approveOverwrite {
				m.choice = "n"
			} else {
				m.choice = "e"
			}
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
			m.updateConfig()
			if m.focusIndex == len(m.inputs) || m.inputType == approveOverwrite {
				if m.inputType == standalone && m.config.SSL.Enabled {
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
		var msg string
		switch m.inputType {
		case viridian:
			msg = viridianInfoMsg
		case standalone:
			msg = standaloneInfoMsg
		case ssl:
			msg = sslInfoMsg
		}
		isApproval := m.inputType == approveOverwrite
		if !isApproval {
			fmt.Fprintf(&b, "%s\n", fmt.Sprintf("%s", noStyle.Render(msg)))
		}
		for i := range m.inputs {
			b.WriteString(m.inputs[i].View())
			if i < len(m.inputs)-1 {
				b.WriteRune('\n')
			}
		}
		if !isApproval {
			button := fmt.Sprintf("%s", noStyle.Render(submitMsg))
			if m.focusIndex == len(m.inputs) {
				button = fmt.Sprintf("%s", selectedItemStyle.Copy().Render(submitMsg))
			}
			fmt.Fprintf(&b, "\n\n%s", button)
		}
		return b.String()
	}
}

func textInputWithPrompt(prompt string) textinput.Model {
	t := textinput.New()
	t.Prompt = prompt
	return t
}

func passwordInput() textinput.Model {
	t := textinput.New()
	t.Prompt = passwordMsg
	t.EchoMode = textinput.EchoPassword
	t.EchoCharacter = '•'
	return t
}

func updateSelectedIem(t *textinput.Model) {
	t.Focus()
	t.PromptStyle = selectedItemStyle
	t.TextStyle = selectedItemStyle
}
