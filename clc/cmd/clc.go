package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/config"
	"github.com/hazelcast/hazelcast-commandline-client/clc/logger"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	puberrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

var (
	// client is currently global in order to have a single client.
	// This is bad.
	// TODO: make the client unique without making it global.
	clientInternal atomic.Pointer[hazelcast.ClientInternal]
)

func getClientInternal() *hazelcast.ClientInternal {
	return clientInternal.Load()
}

func setClientInternal(ci *hazelcast.ClientInternal) {
	clientInternal.Store(ci)
}

func ServerVersionOf(ci *hazelcast.ClientInternal) string {
	return ci.ConnectionManager().RandomConnection().ServerVersion()
}

type Main struct {
	root          *cobra.Command
	cmds          map[string]*cobra.Command
	lg            *logger.Logger
	stderr        io.WriteCloser
	stdout        io.WriteCloser
	stdin         io.Reader
	isInteractive bool
	outputFormat  string
	configLoaded  bool
	props         *plug.Properties
	cc            *CommandContext
	cp            config.Provider
}

func NewMain(arg0, cfgPath string, cfgProvider config.Provider, logPath, logLevel string, sio clc.IO) (*Main, error) {
	rc := &cobra.Command{
		Use:               arg0,
		Short:             "Hazelcast CLC",
		Long:              "Hazelcast CLC",
		Args:              cobra.NoArgs,
		CompletionOptions: cobra.CompletionOptions{DisableDescriptions: true},
		SilenceErrors:     true,
	}
	rc.SetOut(sio.Stdout)
	rc.SetErr(sio.Stderr)
	m := &Main{
		root:   rc,
		cmds:   map[string]*cobra.Command{},
		stdout: clc.NopWriteCloser{W: sio.Stdout},
		stderr: clc.NopWriteCloser{W: sio.Stderr},
		stdin:  sio.Stdin,
		props:  plug.NewProperties(),
		cp:     cfgProvider,
	}
	if logPath == "" {
		logPath = cfgProvider.GetString(clc.PropertyLogPath)
	}
	logPath = paths.ResolveLogPath(logPath)
	if logLevel == "" {
		logLevel = cfgProvider.GetString(clc.PropertyLogLevel)
		if logLevel == "" {
			logLevel = "info"
		}
	}
	if err := m.createLogger(logPath, logLevel); err != nil {
		return nil, err
	}
	for k, v := range cfgProvider.All() {
		m.setConfigProps(m.props, k, v)
	}
	// these properties are managed manually
	m.props.Set(clc.PropertyConfig, cfgPath)
	m.props.Set(clc.PropertyLogPath, logPath)
	m.props.Set(clc.PropertyLogLevel, logLevel)
	m.cc = NewCommandContext(rc, cfgProvider, m.isInteractive)
	if err := m.runInitializers(m.cc); err != nil {
		return nil, err
	}
	if err := m.createCommands(); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *Main) Clone(interactive bool) (*Main, error) {
	mc := *m
	mc.isInteractive = true
	rc := &cobra.Command{
		SilenceErrors: true,
	}
	rc.SetOut(m.stdout)
	rc.SetErr(m.stderr)
	mc.root = rc
	// disable completions command in the interactive mode
	rc.CompletionOptions.DisableDefaultCmd = true
	rc.SetHelpCommand(&cobra.Command{
		Use:   `\help`,
		Short: "Help about commands",
		RunE: func(cmd *cobra.Command, args []string) error {
			return mc.root.Help()
		},
	})
	mc.cmds = map[string]*cobra.Command{}
	mc.cc = NewCommandContext(rc, mc.cp, interactive)
	if err := mc.runInitializers(mc.cc); err != nil {
		return nil, err
	}
	if err := mc.createCommands(); err != nil {
		return nil, err
	}
	return &mc, nil
}

func (m *Main) Root() *cobra.Command {
	return m.root
}

func (m *Main) Execute(ctx context.Context, args ...string) error {
	var cm *cobra.Command
	var cmdArgs []string
	var err error
	if !m.isInteractive {
		cm, cmdArgs, err = m.root.Find(args)
		if err != nil {
			return err
		}
		if cm.Use == "clc" {
			// check whether help or completion is requested
			useShell := true
			for i, arg := range cmdArgs {
				if arg == "--help" || arg == "-h" {
					useShell = false
					break
				}
				if i == 0 && (arg == "help" || arg == "completion" || arg == cobra.ShellCompRequestCmd || arg == cobra.ShellCompNoDescRequestCmd) {
					useShell = false
					break
				}
			}
			// if help was not requested, set shell as the command
			if useShell {
				// check that the first argument is not an invalid commands
				args = append([]string{"shell"}, cmdArgs...)
			}
		}
	} else {
		cm, _, err = m.root.Find(args)
		if err != nil {
			if len(args) > 0 && args[0] == "\\help" {
				return m.root.Help()
			}
			return convertUnknownCommandError(err)
		}
	}
	m.root.SetArgs(args)
	m.props.Push()
	err = m.root.ExecuteContext(ctx)
	m.props.Pop()
	// set all flags to their defaults
	// XXX: it may not work with slices, see: https://github.com/spf13/cobra/issues/1488#issuecomment-1205104931
	if cm != nil {
		cm.Flags().VisitAll(func(f *pflag.Flag) {
			if f.Changed {
				// ignoring the error
				_ = f.Value.Set(f.DefValue)
				f.Changed = false
			}
		})
	}
	return err
}

func (m *Main) Exit() error {
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
	lg, err := logger.New(f, weight)
	if err != nil {
		return err
	}
	m.lg = lg
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
			return err
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
		// check if current command available in current mode
		if !plug.Registry.IsAvailable(m.isInteractive, c.Name) {
			continue
		}
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
		cc := NewCommandContext(cmd, m.cp, m.isInteractive)
		if ci, ok := c.Item.(plug.Initializer); ok {
			if err := ci.Init(cc); err != nil {
				if errors.Is(err, puberrors.ErrNotAvailable) {
					continue
				}
				return fmt.Errorf("initializing command: %w", err)
			}
		}
		// add the backslash prefix for top-level commands in the interactive mode
		if m.isInteractive && parent == m.root {
			cmd.Use = fmt.Sprintf("\\%s", cmd.Use)
		}
		parent.AddGroup(cc.Groups()...)
		if !cc.TopLevel() {
			cmd.RunE = func(cmd *cobra.Command, args []string) error {
				cfs := cmd.Flags()
				props := m.props
				cfs.VisitAll(func(f *pflag.Flag) {
					// skip managed flags
					if f.Name == clc.PropertyConfig || f.Name == clc.PropertyLogPath || f.Name == clc.PropertyLogLevel {
						return
					}
					props.Set(f.Name, convertFlagValue(cfs, f.Name, f.Value))
				})
				sio := clc.IO{
					Stdin:  m.stdin,
					Stderr: m.stderr,
					Stdout: m.stdout,
				}
				ec, err := NewExecContext(m.lg, sio, m.props, func(ctx context.Context, cfg hazelcast.Config) (*hazelcast.ClientInternal, error) {
					if err := m.ensureClient(ctx, cfg); err != nil {
						return nil, err
					}
					return clientInternal.Load(), nil
				}, m.isInteractive)
				if err != nil {
					return err
				}
				ec.SetConfigProvider(m.cp)
				ec.SetMain(m)
				ec.SetArgs(args)
				ec.SetCmd(cmd)
				ctx := context.Background()
				t, err := parseDuration(ec.Props().GetString(clc.PropertyTimeout))
				if err != nil {
					return err
				}
				var cancel context.CancelFunc
				if t != 0 {
					ctx, cancel = context.WithTimeout(ctx, t)
					defer cancel()
				}
				if err := m.runAugmentors(ec, props); err != nil {
					return err
				}
				// to wrap or not to wrap
				// that's the problem
				if _, ok := c.Item.(plug.UnwrappableCommander); ok {
					err = c.Item.Exec(ctx, ec)
				} else {
					err = ec.WrapResult(func() error {
						return c.Item.Exec(ctx, ec)
					})
				}
				if err != nil {
					return err
				}
				if ic, ok := c.Item.(plug.InteractiveCommander); ok {
					ec.SetInteractive(true)
					if _, ok := c.Item.(plug.UnwrappableCommander); ok {
						err = ic.ExecInteractive(ctx, ec)
					} else {
						err = ec.WrapResult(func() error {
							return ic.ExecInteractive(ctx, ec)
						})
					}
					if errors.Is(err, puberrors.ErrNotAvailable) {
						return nil
					}
					return err
				}
				return nil
			}
		}
		parent.AddCommand(cmd)
		m.cmds[c.Name] = cmd
	}
	return nil
}

func (m *Main) ensureClient(ctx context.Context, cfg hazelcast.Config) error {
	if getClientInternal() == nil {
		client, err := hazelcast.StartNewClientWithConfig(ctx, cfg)
		if err != nil {
			return err
		}
		setClientInternal(hazelcast.NewClientInternal(client))
	}
	return nil
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

func convertFlagValue(fs *pflag.FlagSet, name string, v pflag.Value) any {
	switch v.Type() {
	case "string":
		return MustValue(fs.GetString(name))
	case "bool":
		return MustValue(fs.GetBool(name))
	case "int64":
		return MustValue(fs.GetInt64(name))
	}
	panic(fmt.Errorf("cannot convert type: %s", v.Type()))
}

var regexpUnknownCommand = regexp.MustCompile(`unknown command "\\\\([^"]+)"`)

func convertUnknownCommandError(err error) error {
	ss := regexpUnknownCommand.FindStringSubmatch(err.Error())
	if len(ss) < 2 {
		return err
	}
	return fmt.Errorf("unknown command \\%s", ss[1])
}

func init() {
	hazelcast.SetDefaultCompactDeserializer(serialization.GenericCompactDeserializer{})
	hazelcast.SetDefaultPortableDeserializer(serialization.NewGenericPortableSerializer())
}
