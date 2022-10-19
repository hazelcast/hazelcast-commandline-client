package internal

import (
	"context"
	"io"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/internal/log"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type ClientFn func(ctx context.Context) (*hazelcast.Client, error)

type ExecContext struct {
	lg       log.Logger
	stdout   io.Writer
	stderr   io.Writer
	args     []string
	props    *plug.Properties
	clientFn ClientFn
	rows     []output.Row
	ci       *hazelcast.ClientInternal
}

func NewExecContext(lg log.Logger, stdout, stderr io.Writer, args []string, props *plug.Properties, clientFn ClientFn) *ExecContext {
	return &ExecContext{
		lg:       lg,
		stdout:   stdout,
		stderr:   stderr,
		args:     args,
		props:    props,
		clientFn: clientFn,
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

func (ec *ExecContext) OutputRows() []output.Row {
	return ec.rows
}
