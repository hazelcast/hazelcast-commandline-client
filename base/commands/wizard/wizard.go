//go:build base || wizard

package wizard

import (
	"context"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type WizardCommand struct {
}

func (wc WizardCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("wizard")
	help := "Connection wizard"
	cc.SetCommandHelp(help, help)
	return nil
}

func (wc WizardCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	dirs, err := wc.findConfigs()
	if err != nil {
		return err
	}
	if len(dirs) == 0 {
		_, err := tea.NewProgram(initialModel()).Run()
		if err != nil {
			fmt.Printf("could not start program: %s\n", err)
			return err
		}
	} else {
		model, err := tea.NewProgram(initializeList(dirs)).Run()
		if err != nil {
			return err
		}
		if model.View() != "" {
			ec.ChangeConfig(ctx, dirs[model.View()])
		}
	}
	return nil
}

func (wc *WizardCommand) findConfigs() (map[string]string, error) {
	dirs := make(map[string]string)
	configDir := paths.Configs()
	es, err := os.ReadDir(configDir)
	if err != nil {
		return nil, err
	}
	for _, e := range es {
		if !e.IsDir() {
			continue
		}
		if strings.HasPrefix(e.Name(), ".") || strings.HasPrefix(e.Name(), "_") {
			continue
		}
		if paths.Exists(paths.Join(configDir, e.Name(), "config.yaml")) {
			dirs[e.Name()] = paths.Join(configDir, e.Name(), "config.yaml")
		}
	}
	return dirs, nil
}
