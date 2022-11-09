package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/logger"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	puberrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

const (
	viridianCoordinatorURL       = "https://api.viridian.hazelcast.com"
	envHzCloudCoordinatorBaseURL = "HZ_CLOUD_COORDINATOR_BASE_URL"
)

type Main struct {
	root          *cobra.Command
	cmds          map[string]*cobra.Command
	vpr           *viper.Viper
	client        *hazelcast.Client
	lg            *logger.Logger
	stdout        io.WriteCloser
	stderr        io.WriteCloser
	isInteractive bool
	outputFormat  string
	configLoaded  bool
	props         *plug.Properties
	ec            *ExecContext
	cc            *CommandContext
}

func NewMain(cfgPath, logPath, logLevel string) (*Main, error) {
	rc := &cobra.Command{
		Use:   "clc",
		Short: "Hazelcast CLC",
		Long:  "Hazelcast CLC",
	}
	m := &Main{
		root:   rc,
		cmds:   map[string]*cobra.Command{},
		vpr:    viper.New(),
		stdout: nopWriteCloser{W: os.Stdout},
		stderr: nopWriteCloser{W: os.Stderr},
		props:  plug.NewProperties(),
	}
	cfgPath = paths.ResolveConfigPath(cfgPath)
	if _, err := m.loadConfig(cfgPath); err != nil {
		return nil, err
	}
	if logPath == "" {
		logPath = m.vpr.GetString(clc.PropertyLogPath)
	}
	logPath = paths.ResolveLogPath(logPath)
	if logLevel == "" {
		logLevel = m.vpr.GetString(clc.PropertyLogLevel)
	}
	if err := m.createLogger(logPath, logLevel); err != nil {
		return nil, err
	}
	for k, v := range m.vpr.AllSettings() {
		m.setConfigProps(m.props, k, v)
	}
	// these properties are managed manually
	m.props.Set(clc.PropertyConfig, cfgPath)
	m.props.Set(clc.PropertyLogPath, logPath)
	m.props.Set(clc.PropertyLogLevel, logLevel)
	m.cc = NewCommandContext(rc, m.vpr, m.isInteractive)
	if err := m.runInitializers(m.cc); err != nil {
		return nil, err
	}
	cf := func(ctx context.Context) (*hazelcast.Client, error) {
		return m.ensureClient(ctx, m.props)
	}
	m.ec = NewExecContext(m.lg, m.stdout, m.stderr, m.props, cf, m.isInteractive)
	m.ec.SetMain(m)
	if err := m.createCommands(); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *Main) CloneForInteractiveMode() (*Main, error) {
	mc := *m
	mc.isInteractive = true
	return &mc, nil
	/*
		rc := &cobra.Command{
			Use:   "",
			Short: "Hazelcast CLC",
			Long:  "Hazelcast CLC",
		}
		mc := &Main{
			root:          rc,
			cmds:          map[string]*cobra.Command{},
			vpr:           m.vpr,
			client:        m.client,
			lg:            m.lg,
			stdout:        m.stdout,
			stderr:        m.stderr,
			isInteractive: true,
			outputFormat:  m.outputFormat,
			configLoaded:  m.configLoaded,
			props:         m.props,
		}
		cf := func(ctx context.Context) (*hazelcast.Client, error) {
			return mc.ensureClient(ctx, mc.props)
		}
		mc.ec = NewExecContext(m.lg, m.stderr, m.stderr, m.props, cf, true)
		cc := NewCommandContext(rc, mc.vpr, mc.isInteractive)
		if err := mc.runInitializers(cc); err != nil {
			return nil, err
		}
		if err := mc.createCommands(cc); err != nil {
			return nil, err
		}
		return mc, nil
	*/
}

func (m *Main) Root() *cobra.Command {
	return m.root
}

func (m *Main) Execute(args []string) error {
	m.root.SetArgs(args)
	return m.root.Execute()
}

func (m *Main) Exit() error {
	// ignore file close error
	m.lg.Close()
	return nil
}

func (m *Main) createLogger(path, level string) error {
	weight, err := logger.WeightForLevel(level)
	if err != nil {
		return err
	}
	var f io.WriteCloser
	if path == "stderr" {
		f = os.Stderr
	} else {
		f, err = m.createGetLogFile(path)
		if err != nil {
			// failed to open the log file, use stderr
			f = os.Stderr
		}
	}
	m.lg, err = logger.New(f, weight)
	return nil
}

func (m *Main) createGetLogFile(path string) (io.WriteCloser, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (m *Main) runAugmentors(ec plug.ExecContext, props *plug.Properties) error {
	for _, a := range plug.Registry.Augmentors() {
		if err := a.Item.Augment(ec, props); err != nil {
			return fmt.Errorf("augmenting %s: %w", a.Name, err)
		}
	}
	return nil
}

func (m *Main) runInitializers(cc *CommandContext) error {
	for _, ita := range plug.Registry.GlobalInitializers() {
		if err := ita.Item.Init(cc); err != nil {
			return err
		}
	}
	m.root.AddGroup(cc.Groups()...)
	return nil
}

func (m *Main) createCommands() error {
	for _, c := range plug.Registry.Commands() {
		c := c
		// skip interactive commands in interactive mode
		if m.isInteractive {
			if _, ok := c.Item.(plug.InteractiveCommander); ok {
				continue
			}
		}
		// create the command hierarchy
		ps := strings.Split(c.Name, ":")
		if len(ps) == 0 {
			continue
		}
		// parents of the current command
		parent := m.root
		for i := 1; i < len(ps); i++ {
			name := strings.Join(ps[:i], ":")
			p, ok := m.cmds[name]
			if !ok {
				p = &cobra.Command{
					Use: fmt.Sprintf("%s [command] [flags]", ps[i-1]),
				}
				p.SetUsageTemplate(usageTemplate)
				//p = &cobra.Command{}
				m.cmds[name] = p
				parent.AddCommand(p)
			}
			parent = p
		}
		// current command
		cmd := &cobra.Command{
			Use:          ps[len(ps)-1],
			SilenceUsage: true,
		}
		cmd.SetUsageTemplate(usageTemplate)
		cc := NewCommandContext(cmd, m.vpr, m.isInteractive)
		if ci, ok := c.Item.(plug.Initializer); ok {
			if err := ci.Init(cc); err != nil {
				return fmt.Errorf("initializing command: %w", err)
			}
		}
		parent.AddGroup(cc.Groups()...)
		if !cc.TopLevel() {
			cmd.RunE = func(cmd *cobra.Command, args []string) error {
				// resetting the flag values, so they are not persistent between runs.
				// resetting at the end of the function, after the command execution is complete.
				defer m.cc.Reset()
				cfs := cmd.Flags()
				props := m.props
				cfs.VisitAll(func(f *pflag.Flag) {
					// skip managed flags
					if f.Name == clc.PropertyConfig || f.Name == clc.PropertyLogPath || f.Name == clc.PropertyLogLevel {
						return
					}
					props.Set(f.Name, convertFlagValue(cfs, f.Name, f.Value))
				})
				ec := m.ec
				ec.SetArgs(args)
				ec.SetCmd(cmd)
				if err := m.runAugmentors(ec, props); err != nil {
					return err
				}
				if err := c.Item.Exec(cmd.Context(), ec); err != nil {
					return err
				}
				if ic, ok := c.Item.(plug.InteractiveCommander); ok {
					ec.SetInteractive(true)
					err := ic.ExecInteractive(cmd.Context(), ec)
					if errors.Is(err, puberrors.ErrNotAvailable) {
						return nil
					}
					return err
				}
				return ec.FlushOutput()
			}
		}
		parent.AddCommand(cmd)
		m.cmds[c.Name] = cmd
	}
	return nil
}

func (m *Main) ensureClient(ctx context.Context, props plug.ReadOnlyProperties) (*hazelcast.Client, error) {
	if m.client == nil {
		cfg, err := makeConfiguration(props, m.lg)
		if err != nil {
			return nil, err
		}
		m.client, err = hazelcast.StartNewClientWithConfig(ctx, cfg)
		if err != nil {
			return nil, err
		}
	}
	return m.client, nil
}

func (m *Main) loadConfig(path string) (bool, error) {
	if m.configLoaded {
		return true, nil
	}
	m.configLoaded = true
	defaultPath := paths.DefaultConfigPath()
	m.vpr.SetConfigFile(path)
	if err := m.vpr.ReadInConfig(); err != nil {
		// ignore the errors if the path is the default path, it is possible that it does not exist.
		var pe *fs.PathError
		if errors.As(err, &pe) {
			if path == defaultPath {
				return false, nil
			}
		}
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if path == defaultPath {
				return false, nil
			}
		}
		return false, err
	}
	return true, nil
}

func (m *Main) setConfigProps(props *plug.Properties, key string, value any) {
	switch vv := value.(type) {
	case map[string]any:
		for k, v := range vv {
			m.setConfigProps(props, fmt.Sprintf("%s.%s", key, k), v)
		}
	default:
		props.Set(key, value)
	}
}

type nopWriteCloser struct {
	W io.Writer
}

func (nc nopWriteCloser) Write(p []byte) (n int, err error) {
	return nc.W.Write(p)
}

func (nc nopWriteCloser) Close() error {
	return nil
}

func convertFlagValue(fs *pflag.FlagSet, name string, v pflag.Value) any {
	switch v.Type() {
	case "string":
		return MustValue(fs.GetString(name))
	case "bool":
		return MustValue(fs.GetBool(name))
	case "int":
		return MustValue(fs.GetInt(name))
	}
	return v
}
