package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"time"

	"github.com/fatih/color"
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/spf13/cobra"
	"github.com/theckman/yacspin"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/config"
	cmderrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/log"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/terminal"
)

const (
	cancelMsg = " (Ctrl+C to cancel) "
)

type ClientFn func(ctx context.Context, cfg hazelcast.Config) (*hazelcast.ClientInternal, error)

type ExecContext struct {
	lg            log.Logger
	stdout        io.Writer
	stderr        io.Writer
	stdin         io.Reader
	args          []string
	props         *plug.Properties
	isInteractive bool
	cmd           *cobra.Command
	main          *Main
	spinnerWait   time.Duration
	printer       plug.Printer
	cp            config.Provider
}

func NewExecContext(lg log.Logger, sio clc.IO, props *plug.Properties, interactive bool) (*ExecContext, error) {
	return &ExecContext{
		lg:            lg,
		stdout:        sio.Stdout,
		stderr:        sio.Stderr,
		stdin:         sio.Stdin,
		props:         props,
		isInteractive: interactive,
		spinnerWait:   1 * time.Second,
	}, nil
}

func (ec *ExecContext) SetConfigProvider(cfgProvider config.Provider) {
	ec.cp = cfgProvider
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

func (ec *ExecContext) Stdin() io.Reader {
	return ec.stdin
}

func (ec *ExecContext) Args() []string {
	return ec.args
}

func (ec *ExecContext) Props() plug.ReadOnlyProperties {
	return ec.props
}

func (ec *ExecContext) ClientInternal(ctx context.Context) (*hazelcast.ClientInternal, error) {
	ci := ec.main.clientInternal()
	if ci != nil {
		return ci, nil
	}
	cfg, err := ec.cp.ClientConfig(ctx, ec)
	if err != nil {
		return nil, err
	}
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("Connecting to the cluster")
		if err := ec.main.ensureClient(ctx, cfg); err != nil {
			return nil, err
		}
		return ec.main.clientInternal(), nil
	})
	if err != nil {
		return nil, err
	}
	stop()
	ci = ec.main.clientInternal()
	verbose := ec.Props().GetBool(clc.PropertyVerbose)
	if verbose || ec.Interactive() {
		cn := ci.ClusterService().FailoverService().Current().ClusterName
		ec.PrintlnUnnecessary(fmt.Sprintf("Connected to cluster: %s", cn))
	}
	return ci, nil
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
func (ec *ExecContext) ExecuteBlocking(ctx context.Context, f func(context.Context, clc.Spinner) (any, error)) (value any, stop context.CancelFunc, err error) {
	// setup the Ctrl+C handler
	ctx, stop = signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	ch := make(chan any)
	var sp clc.Spinner
	if !ec.Quiet() {
		sc := yacspin.Config{
			Frequency:    100 * time.Millisecond,
			CharSet:      yacspin.CharSets[59],
			SpinnerAtEnd: true,
			Writer:       ec.stderr,
		}
		// ignoring the error here
		s, err := yacspin.New(sc)
		if err == nil {
			// note that checking whether there's no error
			defer s.Stop()
		}
		sp = &simpleSpinner{sp: s}
	} else {
		sp = nopSpinner{}
	}
	go func() {
		v, err := f(ctx, sp)
		if err != nil {
			ch <- err
			return
		}
		ch <- v
	}()
	timer := time.NewTimer(ec.spinnerWait)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			// calling stop but also returning no-op just in case...
			stop()
			if errors.Is(ctx.Err(), context.Canceled) {
				return nil, func() {}, cmderrors.ErrUserCancelled
			}
			return nil, func() {}, ctx.Err()
		case v := <-ch:
			if err, ok := v.(error); ok {
				// if an error came out from the channel, return that as the error
				// calling stop but also returning no-op just in case...
				stop()
				return nil, func() {}, err
			}
			return v, stop, nil
		case <-timer.C:
			if !ec.Quiet() {
				if s, ok := sp.(clc.SpinnerStarter); ok {
					s.Start()
				}
			}
		}
	}
}

func (ec *ExecContext) WrapResult(f func() error) error {
	t := time.Now()
	err := f()
	took := time.Since(t)
	verbose := ec.Props().GetBool(clc.PropertyVerbose)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, cmderrors.ErrUserCancelled) {
			return nil
		}
		msg := MakeErrStr(err)
		if ec.Interactive() {
			I2(fmt.Fprintln(ec.stderr, color.RedString(msg)))
		} else {
			I2(fmt.Fprintln(ec.stderr, msg))
		}
		return cmderrors.WrappedError{Err: err}
	}
	if ec.Quiet() {
		return nil
	}
	if verbose || ec.Interactive() {
		msg := fmt.Sprintf("OK (%d ms)", took.Milliseconds())
		I2(fmt.Fprintln(ec.stderr, msg))
	} else {
		I2(fmt.Fprintln(ec.stderr, "OK"))
	}
	return nil
}

func (ec *ExecContext) PrintlnUnnecessary(text string) {
	if !ec.Quiet() {
		I2(fmt.Fprintln(ec.Stdout(), text))
	}
}

func (ec *ExecContext) Quiet() bool {
	return ec.Props().GetBool(clc.PropertyQuiet) || terminal.IsPipe(ec.Stdin()) || terminal.IsPipe(ec.Stdout())
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

func makeErrorStringFromHTTPResponse(text string) string {
	m := map[string]any{}
	if err := json.Unmarshal([]byte(text), &m); err != nil {
		return text
	}
	if v, ok := m["errorCode"]; ok {
		if v == "ClusterTokenNotFound" {
			return "Discovery token is not valid for this cluster"
		}
	}
	if v, ok := m["message"]; ok {
		if vs, ok := v.(string); ok {
			return vs
		}
	}
	return text
}

type simpleSpinner struct {
	sp   *yacspin.Spinner
	text string
}

func (s *simpleSpinner) Start() {
	// ignoring the error here
	_ = s.sp.Start()
}

func (s *simpleSpinner) SetText(text string) {
	s.text = text
	if text == "" {
		s.sp.Prefix("")
		return
	}
	s.sp.Prefix(text + cancelMsg)
}

func (s *simpleSpinner) SetProgress(progress float32) {
	if progress > 1 {
		progress = 1
	}
	if progress <= 0 {
		s.sp.Suffix("")
		return
	}
	pc := int(progress * 100)
	s.sp.Suffix(fmt.Sprintf(" %3d%%", pc))
}

type nopSpinner struct{}

func (n nopSpinner) SetText(text string) {
	// pass
}

func (n nopSpinner) SetProgress(progress float32) {
	// pass
}
