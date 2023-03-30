//go:build base || wizard

package wizard

import (
	"context"
	"errors"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/hazelcast/hazelcast-commandline-client/clc/config"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

func RunWizard(ctx context.Context, ec plug.ExecContext) error {
	dirs, err := findConfigs()
	if err != nil {
		return err
	}
	if len(dirs) == 0 {
		m := initialModel()
		model, err := tea.NewProgram(m).Run()
		if err != nil {
			return err
		}
		if model.View() == "esc" {
			return errors.New("No configuration is specified.")
		}
		args := m.GetInputs()
		_, err = config.ImportSource(ctx, ec, args[0], args[1])
		if err != nil {
			return err
		}
		ec.ChangeConfig(ctx, paths.Join(paths.Configs(), args[0], "config.yaml"))
	} else {
		m := initializeList(dirs)
		model, err := tea.NewProgram(m).Run()
		if err != nil {
			return err
		}
		if model.View() == "esc" {
			return errors.New("No configuration is specified.")
		}
		ec.ChangeConfig(ctx, dirs[model.View()])
	}
	return nil
}

func findConfigs() (map[string]string, error) {
	dirs := make(map[string]string)
	configDir := paths.Configs()
	es, err := os.ReadDir(configDir)
	if err != nil {
		return dirs, nil
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
