package connectioncmd

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hazelcast/hazelcast-commandline-client/config"
	hzcerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal/file"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	cmd := cobra.Command{
		Use:   "connection-wizard",
		Short: "Assist with connection configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			items := []list.Item{
				item("Hazelcast Viridian"),
				item("Hazelcast Cloud"),
				item("Local"),
				item("Remote"),
			}
			l := list.New(items, itemDelegate{}, 20, 14)
			l.Title = "Where is your Hazelcast cluster (Press Ctrl+C to quit)?"
			l.Styles.Title = lipgloss.NewStyle().Background(lipgloss.Color("6"))
			l.Styles.TitleBar = lipgloss.NewStyle()
			l.SetShowStatusBar(false)
			l.SetShowHelp(false)
			l.SetFilteringEnabled(false)
			m := Model{list: l}
			if err := tea.NewProgram(m).Start(); err != nil {
				return err
			}
			c := config.DefaultConfig()
			var im InputModel
			switch choice {
			case "Hazelcast Viridian":
				im = ViridianInput(&c)
			case "Local", "Remote":
				im = StandaloneInput(&c, choice)
			case "Hazelcast Cloud":
				im = CloudInput(&c)
			default:
				return hzcerrors.NewLoggableError(nil, "No selection during connection-wizard.")
			}
			if err := tea.NewProgram(im).Start(); err != nil {
				return hzcerrors.NewLoggableError(err, "Can not run list model during connection-wizard.")
			}
			exists, err := file.Exists(config.DefaultConfigPath())
			if err != nil {
				return hzcerrors.NewLoggableError(err, "Can not check default config path during connection-wizard.")
			}
			if exists {
				choice = "n"
				im = ApprovalInput()
				if err := tea.NewProgram(im).Start(); err != nil {
					return hzcerrors.NewLoggableError(err, "Can not run input model during connection-wizard.")
				}
			} else {
				choice = "y"
			}
			if choice == "y" {
				err := config.WriteToFile(&c, config.DefaultConfigPath())
				if err != nil {
					return hzcerrors.NewLoggableError(err, "There was an error during overwriting config file.")
				} else if exists {
					cmd.Println("Your config file has been changed. Please re-start CLC to apply new config.")
				}
			}
			return nil
		},
	}
	return &cmd
}
