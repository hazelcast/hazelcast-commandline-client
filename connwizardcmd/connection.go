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
		Short: "Assist with connection configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			m := NewListModel()
			if err := tea.NewProgram(m).Start(); err != nil {
				return err
			}
			c := config.DefaultConfig()
			var im *InputModel
			switch m.Choice() {
			case ChoiceViridian:
				im = showInput(ViridianInput(&c))
			case ChoiceLocal:
				// pass
			case ChoiceRemote:
				im = showInput(StandaloneInput(&c))
			}
			if m.Choice() == "e" {
				return hzcerrors.ErrUserCancelled
			}
			if config.Exists() {
				im = showInput(ApprovalInput())
			}
			if im.Choice() == "y" {
				return writeConfig(cmd, &c)
			}
			return nil
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

func showInput(im *InputModel) *InputModel {
	if err := tea.NewProgram(im).Start(); err != nil {
		panic("Can not run input model during connection-wizard.")
	}
	return im
}
