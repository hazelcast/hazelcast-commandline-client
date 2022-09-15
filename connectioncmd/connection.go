package connectioncmd

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hazelcast/hazelcast-commandline-client/config"
	"github.com/hazelcast/hazelcast-commandline-client/internal/file"
	"github.com/spf13/cobra"
	"os"
)

func New() *cobra.Command {
	cmd := cobra.Command{
		Use:   "connection",
		Short: "Changes your cluster connection config",
		Long:  `Changes your cluster connection config and creates a new config file.`,
		Run: func(cmd *cobra.Command, args []string) {
			items := []list.Item{
				item("Hazelcast Viridian"),
				item("Hazelcast Cloud"),
				item("Local"),
				item("Remote"),
			}
			newConfig := config.DefaultConfig()
			l := list.New(items, itemDelegate{}, 20, 14)
			l.Title = "Where is your Hazelcast cluster (Press Ctrl+C to quit)?"
			l.Styles.Title = lipgloss.NewStyle().Background(lipgloss.Color("6"))
			l.Styles.TitleBar = lipgloss.NewStyle()
			l.SetShowHelp(false)
			l.SetShowStatusBar(false)
			l.SetFilteringEnabled(false)
			m := Model{list: l}
			if err := tea.NewProgram(m).Start(); err != nil {
				os.Exit(1)
			}
			var inputModel InputModel
			if choice == "Hazelcast Viridian" {
				inputModel = ViridianInput(&newConfig)
			} else if choice == "Hazelcast Cloud" {
				inputModel = CloudInput(&newConfig)
			} else if choice == "Local" {
				inputModel = LocalInput(&newConfig)
			} else if choice == "Remote" {
				inputModel = RemoteInput(&newConfig)
			} else if choice == "" {
				return
			}
			if err := tea.NewProgram(inputModel).Start(); err != nil {
				os.Exit(1)
			}
			exists, _ := file.Exists(config.DefaultConfigPath())
			if exists {
				choice = "n"
				inputModel = ApprovalInput()
				if err := tea.NewProgram(inputModel).Start(); err != nil {
					os.Exit(1)
				}
			} else {
				choice = "y"
			}
			if choice == "y" {
				err := config.WriteToFile(&newConfig, config.DefaultConfigPath())
				if err != nil {
					cmd.Println("There was an error during overwriting config file.")
				} else if exists {
					cmd.Println("Your config file has been changed. Please re-start CLC to apply new config.")
				}
			}
		},
	}
	return &cmd
}
