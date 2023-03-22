//go:build base || wizard

package wizard

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
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
	cd := paths.Configs()
	cs, _ := wc.findConfigs(cd)
	items := []list.Item{}
	for _, c := range cs {
		items = append(items, item(c))
	}
	l := list.New(items, itemDelegate{}, 20, listHeight)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	m := model{list: l}

	model, err := tea.NewProgram(m).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	} else {
		if model.View() != "" {
			I2(fmt.Fprintln(ec.Stdout(), model.View()))
		}

	}
	return nil
}

func (wc *WizardCommand) findConfigs(cd string) ([]string, error) {
	var cs []string
	es, err := os.ReadDir(cd)
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
		if paths.Exists(paths.Join(cd, e.Name(), "config.yaml")) {
			cs = append(cs, e.Name())
		}
	}
	return cs, nil
}

func init() {
	Must(plug.Registry.RegisterCommand("wizard", &WizardCommand{}))
}
