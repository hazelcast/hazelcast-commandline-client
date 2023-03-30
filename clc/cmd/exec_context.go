package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"time"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/spf13/cobra"
	"github.com/theckman/yacspin"

	"github.com/hazelcast/hazelcast-commandline-client/base/commands/wizard"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/clc/shell"
	"github.com/hazelcast/hazelcast-commandline-client/errors"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/log"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type ClientFn func(ctx context.Context) (*hazelcast.ClientInternal, error)

type ConfigFn func(ctx context.Context, path string) error

type ExecContext struct {
	lg            log.Logger
	stdout        io.Writer
	stderr        io.Writer
	args          []string
	props         *plug.Properties
	clientFn      ClientFn
	configFn      ConfigFn
	isInteractive bool
	cmd           *cobra.Command
	main          *Main
	spinnerWait   time.Duration
	printer       plug.Printer
}

func NewExecContext(lg log.Logger, stdout, stderr io.Writer, props *plug.Properties, clientFn ClientFn, configFn ConfigFn, interactive bool) (*ExecContext, error) {
	return &ExecContext{
		lg:            lg,
		stdout:        stdout,
		stderr:        stderr,
		props:         props,
		clientFn:      clientFn,
		configFn:      configFn,
		isInteractive: interactive,
		spinnerWait:   1 * time.Second,
	}, nil
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
	if clientInternal != nil {
		return clientInternal, nil
	}
	if ec.Interactive() && !paths.Exists(ec.Props().GetString(clc.PropertyConfig)) {
		wiz := wizard.WizardCommand{}
		wiz.Exec(ctx, ec)
	}
	ci, stop, err := ec.ExecuteBlocking(ctx, "Connecting to the cluster", func(ctx context.Context) (any, error) {
		return ec.clientFn(ctx)
	})
	if err != nil {
		return nil, err
	}
	stop()
	clientInternal = ci.(*hazelcast.ClientInternal)
	if ec.Interactive() && !shell.IsPipe() {
		I2(fmt.Fprintf(ec.stdout, "Connected to cluster: %s\n\n", clientInternal.ClusterService().FailoverService().Current().ClusterName))
	}
	return clientInternal, nil
}

func (ec *ExecContext) ChangeConfig(ctx context.Context, path string) error {
	_, stop, err := ec.ExecuteBlocking(ctx, "", func(ctx context.Context) (any, error) {
		return nil, ec.configFn(ctx, path)
	})
	if err != nil {
		return err
	}
	stop()
	return nil
}

func (ec *ExecContext) Interactive() bool {
	return ec.isInteractive
}

func (ec *ExecContext) AddOutputRows(ctx context.Context, rows ...output.Row) error {
	if len(rows) == 0 {
		return nil
	}
	if err := ec.ensurePrinter(); err != nil {
		return err
	}
	return ec.printer.PrintRows(ctx, ec.stdout, rows)
}

func (ec *ExecContext) AddOutputStream(ctx context.Context, ch <-chan output.Row) error {
	if err := ec.ensurePrinter(); err != nil {
		return err
	}
	return ec.printer.PrintStream(ctx, ec.stdout, output.NewChanRows(ch))
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

func (ec *ExecContext) SetInteractive(value bool) {
	ec.isInteractive = value
}

// ExecuteBlocking runs the given blocking function.
// It displays a spinner in the interactive mode after a timeout.
// The returned stop function must be called at least once to prevent leaks if there's no error.
// Calling returned stop more than once has no effect.
func (ec *ExecContext) ExecuteBlocking(ctx context.Context, hint string, f func(context.Context) (any, error)) (value any, stop context.CancelFunc, err error) {
	// setup the Ctrl+C handler
	ctx, stop = signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	ch := make(chan any)
	go func() {
		v, err := f(ctx)
		if err != nil {
			ch <- err
			return
		}
		ch <- v
	}()
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
			// calling stop but also returning no-op just in case...
			stop()
			return nil, func() {}, errors.ErrUserCancelled
		case v := <-ch:
			if err, ok := v.(error); ok {
				// if an error came out from the channel, return that as the error
				// calling stop but also returning no-op just in case...
				stop()
				return nil, func() {}, err
			}
			return v, stop, nil
		case <-timer.C:
			if ec.isInteractive && s != nil {
				// ignoring the error here
				_ = s.Start()
			}
		}
	}
}

func (ec *ExecContext) ensurePrinter() error {
	if ec.printer != nil {
		return nil
	}
	pn := ec.props.GetString(clc.PropertyFormat)
	pr, ok := plug.Registry.Printers()[pn]
	if !ok {
		return fmt.Errorf("printer %s is not available", pn)
	}
	ec.printer = pr
	return nil
}
