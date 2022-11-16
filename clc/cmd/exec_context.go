package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/spf13/cobra"
	"github.com/theckman/yacspin"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/shell"
	"github.com/hazelcast/hazelcast-commandline-client/errors"
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
	mu            *sync.Mutex
	rows          []output.Row
	ci            *hazelcast.ClientInternal
	isInteractive bool
	cmd           *cobra.Command
	main          *Main
	spinnerWait   time.Duration
}

func NewExecContext(lg log.Logger, stdout, stderr io.Writer, props *plug.Properties, clientFn ClientFn, interactive bool) *ExecContext {
	return &ExecContext{
		lg:            lg,
		stdout:        stdout,
		stderr:        stderr,
		props:         props,
		clientFn:      clientFn,
		isInteractive: interactive,
		mu:            &sync.Mutex{},
		spinnerWait:   1 * time.Second,
	}
}

func (ec *ExecContext) SetArgs(args []string) {
	ec.args = args
}

func (ec *ExecContext) SetCmd(cmd *cobra.Command) {
	ec.cmd = cmd
}

func (ec *ExecContext) SetMain(main *Main) {
	ec.main = main
}

func (ec *ExecContext) Main() *Main {
	return ec.main
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
	client, err := ec.ExecuteBlocking(ctx, "Connecting to the cluster", func(ctx context.Context) (any, error) {
		return ec.clientFn(ctx)
	})
	if err != nil {
		return nil, err
	}
	ec.ci = hazelcast.NewClientInternal(client.(*hazelcast.Client))
	return ec.ci, nil
}

func (ec *ExecContext) Interactive() bool {
	return ec.isInteractive
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

func (ec *ExecContext) FlushOutput() error {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	rows := ec.outputRows()
	pn := ec.props.GetString(clc.PropertyFormat)
	pr, ok := plug.Registry.Printers()[pn]
	if !ok {
		return fmt.Errorf("printer %s is not available", pn)
	}
	if len(rows) == 0 {
		return nil
	}
	return pr.Print(os.Stdout, output.NewSimpleRows(rows))
}

func (ec *ExecContext) SetInteractive(value bool) {
	ec.isInteractive = value
}

func (ec *ExecContext) ExecuteBlocking(ctx context.Context, hint string, f func(context.Context) (any, error)) (any, error) {
	ch := make(chan any)
	go func() {
		v, err := f(ctx)
		if err != nil {
			ch <- err
			return
		}
		ch <- v
	}()
	// setup the Ctrl+C handler
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	defer stop()
	timer := time.NewTimer(ec.spinnerWait)
	defer timer.Stop()
	var s *yacspin.Spinner
	if ec.isInteractive && !shell.IsPipe() {
		if hint != "" {
			hint = fmt.Sprintf("%s ", hint)
		}
		hint = fmt.Sprintf("%s(Ctrl+C to cancel) ", hint)
		sc := yacspin.Config{
			Frequency:    100 * time.Millisecond,
			CharSet:      yacspin.CharSets[59],
			Prefix:       hint,
			SpinnerAtEnd: true,
		}
		// ignoring the error here
		s, _ = yacspin.New(sc)
		defer s.Stop()
	}
	for {
		select {
		case <-ctx.Done():
			return nil, errors.ErrUserCancelled
		case v := <-ch:
			if err, ok := v.(error); ok {
				return nil, err
			}
			return v, nil
		case <-timer.C:
			if ec.isInteractive && s != nil {
				// ignoring the error here
				_ = s.Start()
			}
		}
	}
}

func (ec *ExecContext) outputRows() []output.Row {
	// assumes called under lock
	r := ec.rows
	ec.rows = nil
	return r
}
