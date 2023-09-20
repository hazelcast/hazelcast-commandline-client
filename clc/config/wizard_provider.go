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
	if err == nil {
		// note that comparing err to nil
		return cfg, nil
	}
	var configName string
	if !errors.Is(err, clcerrors.ErrNoClusterConfig) {
		return hazelcast.Config{}, err
	}
	cs, err := FindAll(paths.Configs())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return hazelcast.Config{}, clcerrors.ErrNoClusterConfig
		}
	}
	if len(cs) == 0 {
		return hazelcast.Config{}, clcerrors.ErrNoClusterConfig
	}
	if len(cs) == 1 {
		configName = cs[0]
	}
	if configName == "" {
		if terminal.IsPipe(maybeUnwrapStdout(ec)) {
			return hazelcast.Config{}, fmt.Errorf(`no configuration was provided and cannot display the configuration wizard; use the --config flag`)
		}
		// ask the config to the user
		configName, err = p.runWizard(cs)
		if err != nil {
			return hazelcast.Config{}, err
		}
	}
	fp, err := NewFileProvider(configName)
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

func (p *WizardProvider) runWizard(configNames []string) (string, error) {
	m := wizard.InitializeList(configNames)
	model, err := tea.NewProgram(m).Run()
	if err != nil {
		return "", err
	}
	if model.View() == "" {
		return "", clcerrors.ErrNoClusterConfig
	}
	return model.View(), nil
}
