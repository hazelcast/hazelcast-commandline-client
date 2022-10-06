package connwizardcmd

import (
	"errors"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hazelcast/hazelcast-commandline-client/config"
	hzcerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	cmd := cobra.Command{
		Use:   "connection-wizard",
		Short: "Assist with connection configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			exists := config.ConfigExists()
			items := []list.Item{
				item("Hazelcast Viridian"),
				item("Local (Default)"),
				item("Standalone (Remote or Local)"),
			}
			l := list.New(items, itemDelegate{}, 20, 14)
			l.Title = "Where is your Hazelcast cluster (Press Ctrl+C to quit)?"
			l.Styles.Title = noStyle
			l.Styles.TitleBar = noStyle
			l.SetShowStatusBar(false)
			l.SetShowHelp(false)
			l.SetFilteringEnabled(false)
			m := ListModel{list: l}
			if err := tea.NewProgram(m).Start(); err != nil {
				return err
			}
			if err := handleExit(cmd); err != nil {
				return nil
			}
			c := config.DefaultConfig()
			var im InputModel
			switch choice {
			case "Hazelcast Viridian":
				im = ViridianInput(&c)
			case "Standalone (Remote or Local)":
				im = StandaloneInput(&c)
			case "Local (Default)":
				return handleWrite(cmd, &c, "y")
			}
			if err := tea.NewProgram(im).Start(); err != nil {
				return hzcerrors.NewLoggableError(err, "Can not run list model during connection-wizard.")
			}
			if err := handleExit(cmd); err != nil {
				return nil
			}
			choice = "y"
			if exists {
				im = ApprovalInput()
				if err := tea.NewProgram(im).Start(); err != nil {
					return hzcerrors.NewLoggableError(err, "Can not run input model during connection-wizard.")
				}
			}
			return handleWrite(cmd, &c, choice)
		},
	}
	return &cmd
}

func handleExit(cmd *cobra.Command) error {
	if choice == "e" {
		if !config.ConfigExists() {
			cmd.Println("You did not make any selection. CLC will connect to the default cluster running at localhost:5701.")
		}
		return errors.New("")
	}
	return nil
}

func handleWrite(cmd *cobra.Command, c *config.Config, choice string) error {
	if choice == "y" {
		exists := config.ConfigExists()
		err := config.WriteToFile(c, config.DefaultConfigPath())
		if err != nil {
			return hzcerrors.NewLoggableError(err, "There was an error during overwriting config file.")
		} else if exists {
			cmd.Println("Your config file has been changed. Please re-start CLC to apply new config.")
		}
	}
	return nil
}
