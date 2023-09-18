package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync/atomic"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/spf13/pflag"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
	"github.com/hazelcast/hazelcast-commandline-client/internal/terminal"

	"github.com/hazelcast/hazelcast-commandline-client/clc/config/wizard"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	clcerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type WizardProvider struct {
	fp  *atomic.Pointer[FileProvider]
	cfg hazelcast.Config
}

func NewWizardProvider(path string) (*WizardProvider, error) {
	fp, err := NewFileProvider(path)
	if err != nil {
		return nil, err
	}
	var fpp atomic.Pointer[FileProvider]
	fpp.Store(fp)
	return &WizardProvider{fp: &fpp}, nil
}

func (p *WizardProvider) GetString(key string) string {
	return p.fp.Load().GetString(key)
}

func (p *WizardProvider) Set(key string, value any) {
	p.fp.Load().Set(key, value)
}

func (p *WizardProvider) All() map[string]any {
	return p.fp.Load().All()
}

func (p *WizardProvider) BindFlag(name string, flag *pflag.Flag) {
	p.fp.Load().BindFlag(name, flag)
}

func maybeUnwrapStdout(ec plug.ExecContext) any {
	if v, ok := ec.Stdout().(clc.NopWriteCloser); ok {
		return v.W
	}
	return ec.Stdout()
}

func (p *WizardProvider) ClientConfig(ctx context.Context, ec plug.ExecContext) (hazelcast.Config, error) {
	cfg, err := p.fp.Load().ClientConfig(ctx, ec)
	if err != nil {
		if terminal.IsPipe(maybeUnwrapStdout(ec)) {
			return hazelcast.Config{}, fmt.Errorf(`no configuration was provided and cannot display the configuration wizard; use the --config flag`)
		}
		// ask the config to the user
		name, err := p.runWizard(ctx, ec)
		if err != nil {
			return hazelcast.Config{}, err
		}
		fp, err := NewFileProvider(name)
		if err != nil {
			return cfg, err
		}
		config, err := fp.ClientConfig(ctx, ec)
		if err != nil {
			return hazelcast.Config{}, err
		}
		p.fp.Store(fp)
		return config, nil
	}
	return cfg, nil
}

func (p *WizardProvider) runWizard(ctx context.Context, ec plug.ExecContext) (string, error) {
	cs, err := FindAll(paths.Configs())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = os.MkdirAll(paths.Configs(), 0700)
		}
		if err != nil {
			return "", err
		}
	}
	if len(cs) == 0 {
		m := wizard.InitialModel()
		mv, err := tea.NewProgram(m).Run()
		if err != nil {
			return "", err
		}
		if mv.View() == "" {
			return "", clcerrors.ErrNoClusterConfig
		}
		args := m.GetInputs()
		stages := MakeImportStages(ec, args[0])
		_, err = stage.Execute(ctx, ec, args[1], stage.NewFixedProvider(stages...))
		if err != nil {
			return "", err
		}
		return args[0], nil
	}
	m := wizard.InitializeList(cs)
	model, err := tea.NewProgram(m).Run()
	if err != nil {
		return "", err
	}
	if model.View() == "" {
		return "", clcerrors.ErrNoClusterConfig
	}
	return model.View(), nil
}
