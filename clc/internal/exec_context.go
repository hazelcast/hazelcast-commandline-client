package internal

import (
	"context"
	"io"
	"os"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/spf13/cobra"

	"github.com/hazelcast/hazelcast-commandline-client/clc/property"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/log"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type ClientFn func(ctx context.Context) (*hazelcast.Client, error)

type ExecContext struct {
	lg            log.Logger
	stdout        io.Writer
	stderr        io.Writer
	args          []string
	props         *plug.Properties
	clientFn      ClientFn
	rows          []output.Row
	ci            *hazelcast.ClientInternal
	isInteractive bool
	cmd           *cobra.Command
}

func NewExecContext(lg log.Logger, stdout, stderr io.Writer, args []string, props *plug.Properties, clientFn ClientFn, interactive bool, cmd *cobra.Command) *ExecContext {
	return &ExecContext{
		lg:            lg,
		stdout:        stdout,
		stderr:        stderr,
		args:          args,
		props:         props,
		clientFn:      clientFn,
		isInteractive: interactive,
		cmd:           cmd,
	}
}

func (ec *ExecContext) Logger() log.Logger {
	return ec.lg
}

func (ec *ExecContext) Stdout() io.Writer {
	return ec.stdout
}

func (ec *ExecContext) Stderr() io.Writer {
	return ec.stderr
}

func (ec *ExecContext) Args() []string {
	return ec.args
}

func (ec *ExecContext) Props() plug.ReadOnlyProperties {
	return ec.props
}

func (ec *ExecContext) ClientInternal(ctx context.Context) (*hazelcast.ClientInternal, error) {
	if ec.ci != nil {
		return ec.ci, nil
	}
	client, err := ec.clientFn(ctx)
	if err != nil {
		return nil, err
	}
	ec.ci = hazelcast.NewClientInternal(client)
	return ec.ci, nil
}

func (ec *ExecContext) Interactive() bool {
	return false
}

func (ec *ExecContext) AddOutputRows(row ...output.Row) {
	ec.rows = append(ec.rows, row...)
}

func (ec *ExecContext) ShowHelpAndExit() {
	Must(ec.cmd.Help())
	if !ec.isInteractive {
		os.Exit(0)
	}
}

func (ec *ExecContext) CommandName() string {
	return ec.cmd.CommandPath()
}

func (ec *ExecContext) QuitCh() <-chan struct{} {
	return nil
}

func (ec *ExecContext) FlushOutput() error {
	prs := map[string]plug.Printer{}
	for _, pr := range plug.Registry.Printers() {
		prs[pr.Name] = pr.Item
	}
	name := ec.Props().GetString(property.OutputType)
	pr := prs[name]
	err := pr.Print(os.Stdout, ec.OutputRows())
	ec.rows = nil
	if err != nil {
		return err
	}
	return nil
}

func (ec *ExecContext) OutputRows() []output.Row {
	return ec.rows
}
