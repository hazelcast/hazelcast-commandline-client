package clc

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/hazelcast/hazelcast-commandline-client/clc/internal"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type Main struct {
	root          *cobra.Command
	cmds          map[string]*cobra.Command
	lg            *internal.Logger
	stdout        io.WriteCloser
	stderr        io.WriteCloser
	ec            plug.ExecContext
	isInteractive bool
}

func NewMain() *Main {
	rc := &cobra.Command{
		Use:   "clc",
		Short: "Hazelcast CLC",
		Long:  "Hazelcast Command Line Client",
	}
	m := &Main{
		root:   rc,
		cmds:   map[string]*cobra.Command{},
		lg:     internal.NewLogger(nopWriteCloser{W: os.Stderr}),
		stdout: nopWriteCloser{W: os.Stdout},
		stderr: nopWriteCloser{W: os.Stderr},
	}
	err := m.createCommands()
	if err != nil {
		panic(err)
	}
	return m
}

func (m *Main) Execute() error {
	return m.root.Execute()
}

func (m *Main) runAugmentors(ec plug.ExecContext, props *plug.Properties) error {
	for _, a := range plug.Registry.Augmentors() {
		if err := a.Item.Augment(ec, props); err != nil {
			return fmt.Errorf("augmenting %s: %w", a.Name, err)
		}
	}
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
				p = &cobra.Command{
					Use: ps[i-1],
				}
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
		cc := internal.NewCommandContext(cmd, m.isInteractive)
		if err := c.Item.Init(cc); err != nil {
			return fmt.Errorf("initializing command: %w", err)
		}
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			fs := cmd.Flags()
			props := plug.NewProperties()
			fs.Visit(func(f *pflag.Flag) {
				props.Set(f.Name, convertFlagValue(fs, f.Name, f.Value))
			})
			ec := internal.NewExecContext(m.lg, m.stdout, m.stderr, args, props, m.ensureClient)
			if err := m.runAugmentors(ec, props); err != nil {
				return err
			}
			err := c.Item.Exec(ec)
			if err != nil {
				return err
			}
			return m.printRows(ec)
		}
		parent.AddCommand(cmd)
		m.cmds[c.Name] = cmd
	}
	return nil
}

func (m *Main) ensureClient(ctx context.Context) (*hazelcast.Client, error) {
	return hazelcast.StartNewClient(ctx)
}

func (m *Main) printRows(ec *internal.ExecContext) error {
	prs := map[string]plug.Printer{}
	for _, pr := range plug.Registry.Printers() {
		prs[pr.Name] = pr.Item
	}
	pr := prs["table"]
	return pr.Print(os.Stdout, ec.OutputRows())
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
