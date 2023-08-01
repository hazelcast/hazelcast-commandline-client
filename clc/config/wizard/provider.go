package wizard

import (
	"context"
	"errors"
	"os"
	"sync/atomic"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/spf13/pflag"

	"github.com/hazelcast/hazelcast-commandline-client/clc/config"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	clcerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
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

func (p *Provider) ClientConfig(ctx context.Context, ec plug.ExecContext) (hazelcast.Config, error) {
	cfg, err := p.fp.Load().ClientConfig(ctx, ec)
	if err != nil {
		// ask the config to the user
		name, err := p.runWizard(ctx, ec)
		if err != nil {
			return hazelcast.Config{}, err
		}
		fp, err := config.NewFileProvider(name)
		if err != nil {
			return cfg, err
		}
		p.fp.Store(fp)
		return fp.ClientConfig(ctx, ec)
	}
	return cfg, nil
}

func (p *Provider) runWizard(ctx context.Context, ec plug.ExecContext) (string, error) {
	cs, err := config.FindAll(paths.Configs())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = os.MkdirAll(paths.Configs(), 0700)
		}
		if err != nil {
			return "", err
		}
	}
	if len(cs) == 0 {
		m := initialModel()
		mv, err := tea.NewProgram(m).Run()
		if err != nil {
			return "", err
		}
		if mv.View() == "" {
			return "", clcerrors.ErrNoClusterConfig
		}
		args := m.GetInputs()
		_, err = config.ImportSource(ctx, ec, args[0], args[1])
		if err != nil {
			return "", err
		}
		return args[0], nil
	}
	m := initializeList(cs)
	model, err := tea.NewProgram(m).Run()
	if err != nil {
		return "", err
	}
	if model.View() == "" {
		return "", clcerrors.ErrNoClusterConfig
	}
	return model.View(), nil
}
