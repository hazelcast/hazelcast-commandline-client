package connwizardcmd

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/hazelcast/hazelcast-commandline-client/config"
	hzcerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
)

func New() *cobra.Command {
	cmd := cobra.Command{
		Use:   "connection-wizard",
		Short: "Show the connection configuration wizard",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true
			if config.Exists() {
				response := showInput(ApprovalInput())
				if response != "y" {
					return hzcerrors.ErrUserCancelled
				}
			}
			m := NewListModel()
			m.Show()
			c := config.DefaultConfig()
			switch m.Choice() {
			case ChoiceViridian:
				showInput(ViridianInput(&c))
			case ChoiceLocal:
				// pass
			case ChoiceRemote:
				showInput(StandaloneInput(&c))
			case "e":
				return hzcerrors.ErrUserCancelled
			}
			return writeConfig(cmd, &c)
		},
	}
	return &cmd
}

func writeConfig(cmd *cobra.Command, c *config.Config) error {
	err := config.WriteToFile(c, config.DefaultConfigPath())
	if err != nil {
		return hzcerrors.NewLoggableError(err, "There was an error while overwriting config file.")
	}
	cmd.Printf("Configuration was saved to: %s\n", config.DefaultConfigPath())
	return nil
}

func showInput(im *InputModel) string {
	if err := tea.NewProgram(im).Start(); err != nil {
		panic("Can not run input model during connection-wizard.")
	}
	return im.Choice()
}
