//go:build base || wizard

package wizard

import (
	"context"
	"sync/atomic"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/spf13/pflag"

	"github.com/hazelcast/hazelcast-commandline-client/clc/config"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type Provider struct {
	fp  *atomic.Pointer[config.FileProvider]
	cfg hazelcast.Config
}

func NewProvider(path string) (*Provider, error) {
	fp, err := config.NewFileProvider(path)
	if err != nil {
		return nil, err
	}
	var fpp atomic.Pointer[config.FileProvider]
	fpp.Store(fp)
	return &Provider{fp: &fpp}, nil
}

func (p *Provider) GetString(key string) string {
	return p.fp.Load().GetString(key)
}

func (p *Provider) Set(key string, value any) {
	p.fp.Load().Set(key, value)
}

func (p *Provider) All() map[string]any {
	return p.fp.Load().All()
}

func (p *Provider) BindFlag(name string, flag *pflag.Flag) {
	p.fp.Load().BindFlag(name, flag)
}

func (p *Provider) ClientConfig(ec plug.ExecContext) (hazelcast.Config, error) {
	cfg, err := p.fp.Load().ClientConfig(ec)
	if err != nil {
		if !ec.Interactive() {
			return hazelcast.Config{}, err
		}
		// ask the config to the user
		name, err := p.runWizard(context.Background(), ec)
		if err != nil {
			return hazelcast.Config{}, err
		}
		fp, err := config.NewFileProvider(name)
		if err != nil {
			return cfg, err
		}
		p.fp.Store(fp)
		return fp.ClientConfig(ec)
	}
	return cfg, nil
}

func (p *Provider) runWizard(ctx context.Context, ec plug.ExecContext) (string, error) {
	dirs, err := config.FindAll(paths.Configs())
	if err != nil {
		return "", err
	}
	if len(dirs) == 0 {
		m := initialModel()
		model, err := tea.NewProgram(m).Run()
		if err != nil {
			return "", err
		}
		if model.View() == "esc" {
			return "", errors.ErrNoClusterConfig
		}
		args := m.GetInputs()
		_, err = config.ImportSource(ctx, ec, args[0], args[1])
		if err != nil {
			return "", err
		}
		return args[0], nil
	}
	m := initializeList(dirs)
	model, err := tea.NewProgram(m).Run()
	if err != nil {
		return "", err
	}
	if model.View() == "esc" {
		return "", errors.ErrNoClusterConfig
	}
	return model.View(), nil
}
