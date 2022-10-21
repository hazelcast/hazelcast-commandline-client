package clc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/hazelcast/hazelcast-commandline-client/clc/internal"
	"github.com/hazelcast/hazelcast-commandline-client/clc/internal/logger"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/clc/property"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

type Main struct {
	root          *cobra.Command
	cmds          map[string]*cobra.Command
	vpr           *viper.Viper
	client        *hazelcast.Client
	lg            *logger.Logger
	stdout        io.WriteCloser
	stderr        io.WriteCloser
	ec            plug.ExecContext
	isInteractive bool
	outputType    string
	configLoaded  bool
}

func NewMain(interactive bool) *Main {
	rc := &cobra.Command{
		Use:   "clc",
		Short: "Hazelcast CLC",
		Long:  "Hazelcast Command Line Client",
	}
	m := &Main{
		root:          rc,
		cmds:          map[string]*cobra.Command{},
		vpr:           viper.New(),
		stdout:        nopWriteCloser{W: os.Stdout},
		stderr:        nopWriteCloser{W: os.Stderr},
		isInteractive: interactive,
	}
	m.createLogger()
	cc := internal.NewCommandContext(rc, m.vpr, m.isInteractive)
	if err := m.runInitializers(cc); err != nil {
		// TODO:
		panic(err)
	}
	if err := m.createCommands(); err != nil {
		// TODO:
		panic(err)
	}
	return m
}

func (m *Main) Execute() error {
	return m.root.Execute()
}

func (m *Main) Exit() error {
	// ignore file close error
	m.lg.Close()
	return nil
}

func (m *Main) createLogger() {
	m.lg = MustValue(logger.New(os.Stderr, logger.WeightInfo))
}

func (m *Main) updateLogger() {
	path := m.vpr.GetString(property.LogFile)
	if path == "" {
		path = paths.DefaultLogPath(time.Now())
	}
	// Continue to use the temporary logger if the output is for stderr
	if path == "stderr" {
		return
	}
	weight, err := logger.WeightForLevel(m.vpr.GetString(property.LogLevel))
	f, err := m.createGetLogFile(path)
	if err != nil {
		m.lg.Info("Failed to open the log file, using stderr: %s", err.Error())
		return
	}
	m.lg.SetWriter(f)
	m.lg.SetWeight(weight)
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

func (m *Main) runInitializers(cc *internal.CommandContext) error {
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
				p = &cobra.Command{Use: ps[i-1]}
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
		cc := internal.NewCommandContext(cmd, m.vpr, m.isInteractive)
		if ci, ok := c.Item.(plug.Initializer); ok {
			if err := ci.Init(cc); err != nil {
				return fmt.Errorf("initializing command: %w", err)
			}
		}
		parent.AddGroup(cc.Groups()...)
		if cc.TopLevel() {
			// since this is a top level command, it should always display the help.
			cmd.Args = func(cmd *cobra.Command, args []string) error {
				Must(cmd.Help())
				if !m.isInteractive {
					os.Exit(0)
				}
				return nil
			}
		}
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			cfs := cmd.Flags()
			props := plug.NewProperties()
			cfs.Visit(func(f *pflag.Flag) {
				props.Set(f.Name, convertFlagValue(cfs, f.Name, f.Value))
			})
			cf := func(ctx context.Context) (*hazelcast.Client, error) {
				return m.ensureClient(ctx, props)
			}
			ec := internal.NewExecContext(m.lg, m.stdout, m.stderr, args, props, cf, m.isInteractive, cmd)
			if err := m.runAugmentors(ec, props); err != nil {
				return err
			}
			if _, err := m.loadConfig(props); err != nil {
				return err
			}
			for k, v := range m.vpr.AllSettings() {
				m.setConfigProps(props, k, v)
			}
			m.updateLogger()
			err := c.Item.Exec(ec)
			if err != nil {
				return err
			}
			if ic, ok := c.Item.(plug.InteractiveCommander); ok {
				return ic.ExecInteractive(ec)
			}
			return ec.FlushOutput()
		}
		parent.AddCommand(cmd)
		m.cmds[c.Name] = cmd
	}
	return nil
}

func (m *Main) ensureClient(ctx context.Context, props plug.ReadOnlyProperties) (*hazelcast.Client, error) {
	var err error
	if m.client == nil {
		cfg := m.makeConfiguration(props)
		m.client, err = hazelcast.StartNewClientWithConfig(ctx, cfg)
		if err != nil {
			return nil, err
		}
	}
	return m.client, nil
}

func (m *Main) makeConfiguration(props plug.ReadOnlyProperties) hazelcast.Config {
	var cfg hazelcast.Config
	if ca := props.GetString("cluster.address"); ca != "" {
		cfg.Cluster.Network.SetAddresses(ca)
	}
	if m.lg != nil {
		cfg.Logger.CustomLogger = m.lg
	}
	sd := props.GetString(property.SchemaDir)
	if sd == "" {
		sd = filepath.Join(paths.HomeDir(), "schemas")
	}
	if err := serialization.UpdateConfigWithRecursivePaths(&cfg, sd); err != nil {
		m.lg.Error(fmt.Errorf("setting schema dir: %w", err))
	}
	return cfg
}

func (m *Main) loadConfig(props plug.ReadOnlyProperties) (bool, error) {
	if m.configLoaded {
		return true, nil
	}
	m.configLoaded = true
	defaultPath := paths.DefaultConfigPath()
	path := props.GetString(property.ConfigPath)
	if path == "" {
		path = defaultPath
	}
	m.vpr.SetConfigFile(path)
	if err := m.vpr.ReadInConfig(); err != nil {
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
	}
	return v
}
