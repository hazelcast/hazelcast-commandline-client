package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/spf13/cobra"
	"github.com/theckman/yacspin"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/config"
	cmderrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/log"
	"github.com/hazelcast/hazelcast-commandline-client/internal/maps"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/str"
	"github.com/hazelcast/hazelcast-commandline-client/internal/terminal"
	"github.com/hazelcast/hazelcast-commandline-client/internal/types"
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
	kwargs        map[string]any
	props         *plug.Properties
	mode          Mode
	cmd           *cobra.Command
	main          *Main
	spinnerWait   time.Duration
	printer       plug.Printer
	cp            config.Provider
	spinnerPaused atomic.Bool
}

func NewExecContext(lg log.Logger, sio clc.IO, props *plug.Properties, mode Mode) (*ExecContext, error) {
	return &ExecContext{
		lg:          lg,
		stdout:      sio.Stdout,
		stderr:      sio.Stderr,
		stdin:       sio.Stdin,
		props:       props,
		mode:        mode,
		spinnerWait: 1 * time.Second,
		kwargs:      map[string]any{},
	}, nil
}

func (ec *ExecContext) SetConfigProvider(cfgProvider config.Provider) {
	ec.cp = cfgProvider
}

func (ec *ExecContext) SetArgs(args []string, argSpecs []ArgSpec) error {
	ec.args = args
	kw, err := makeKeywordArgs(args, argSpecs)
	if err != nil {
		return err
	}
	ec.kwargs = kw
	return nil
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

func (ec *ExecContext) Arg0() string {
	return ec.main.Arg0()
}

func (ec *ExecContext) GetStringArg(key string) string {
	return maps.GetString(ec.kwargs, key)
}

func (ec *ExecContext) GetStringSliceArg(key string) []string {
	return maps.GetStringSlice(ec.kwargs, key)
}

func (ec *ExecContext) GetKeyValuesArg(key string) types.KeyValues[string, string] {
	return maps.GetKeyValues[string, any, string, string](ec.kwargs, key)
}

func (ec *ExecContext) GetInt64Arg(key string) int64 {
	return maps.GetInt64(ec.kwargs, key)
}

func (ec *ExecContext) Props() plug.ReadOnlyProperties {
	return ec.props
}

func (ec *ExecContext) ConfigPath() string {
	return ec.cp.GetString(clc.PropertyConfig)
}

func (ec *ExecContext) ClientInternal(ctx context.Context) (*hazelcast.ClientInternal, error) {
	ci := ec.main.clientInternal()
	if ci != nil {
		return ci, nil
	}
	ec.pauseSpinner()
	cfg, err := ec.cp.ClientConfig(ctx, ec)
	if err != nil {
		// unpausing here, since can't use defer
		ec.unpauseSpinner()
		return nil, err
	}
	ec.unpauseSpinner()
	if err := ec.main.ensureClient(ctx, cfg); err != nil {
		return nil, err
	}
	return ec.main.clientInternal(), nil
}

func (ec *ExecContext) Interactive() bool {
	return ec.mode == ModeInteractive
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
	if !ec.Interactive() {
		os.Exit(0)
	}
}

func (ec *ExecContext) CommandName() string {
	return ec.cmd.CommandPath()
}

func (ec *ExecContext) SetMode(mode Mode) {
	ec.mode = mode
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
	var started bool
	ticker := time.NewTicker(ec.spinnerWait)
	defer ticker.Stop()
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
		case <-ticker.C:
			if !ec.Quiet() {
				if s, ok := sp.(clc.SpinnerStarter); ok {
					if !ec.spinnerPaused.Load() {
						if !started {
							started = true
							s.Start()
						}
					}
				}
			}
		}
	}
}

func (ec *ExecContext) PrintlnUnnecessary(text string) {
	if !ec.Quiet() {
		I2(fmt.Fprintln(ec.Stdout(), str.Colorize(text)))
	}
}

func (ec *ExecContext) pauseSpinner() {
	ec.spinnerPaused.Store(true)
}

func (ec *ExecContext) unpauseSpinner() {
	ec.spinnerPaused.Store(false)
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

func makeKeywordArgs(args []string, argSpecs []ArgSpec) (map[string]any, error) {
	kw := make(map[string]any, len(argSpecs))
	var maxCnt int
	for i, s := range argSpecs {
		spec := argSpecs[i]
		maxCnt = addWithOverflow(maxCnt, s.Max)
		if s.Max-s.Min > 0 {
			if i == len(argSpecs)-1 {
				// if this is the last spec and a range of orguments is expected
				arg := args[i:]
				if len(arg) < spec.Min {
					return nil, fmt.Errorf("expected at least %d %s arguments, but received %d", spec.Min, spec.Title, len(arg))
				}
				if len(arg) > spec.Max {
					return nil, fmt.Errorf("expected at most %d %s arguments, but received %d", spec.Max, spec.Title, len(arg))
				}
				vs, err := convertSliceArg(arg, spec.Type)
				if err != nil {
					return nil, fmt.Errorf("converting argument %s: %w", spec.Title, err)
				}
				kw[s.Key] = vs
				break
			}
			return nil, errors.New("invalid argument spec: only the last argument may take a range")
		}
		// note that this code is never executed under normal circumstances
		// since the arguments are validated before running this function
		if i >= len(args) {
			return nil, fmt.Errorf("%s is required", spec.Title)
		}
		value, err := convertArg(args[i], spec.Type)
		if err != nil {
			return nil, fmt.Errorf("converting argument %s: %w", spec.Title, err)
		}
		kw[s.Key] = value
	}
	// note that this code is never executed under normal circumstances
	// since the arguments are validated before running this function
	if len(args) > maxCnt {
		return nil, fmt.Errorf("unexpected arguments")
	}
	return kw, nil
}

func convertArg(value string, typ ArgType) (any, error) {
	switch typ {
	case ArgTypeString:
		return value, nil
	case ArgTypeInt64:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
	return nil, fmt.Errorf("unknown type: %d", typ)
}

func convertSliceArg(values []string, typ ArgType) (any, error) {
	switch typ {
	case ArgTypeStringSlice:
		args := make([]string, len(values))
		copy(args, values)
		return args, nil
	case ArgTypeInt64Slice:
		args := make([]int64, len(values))
		for i, v := range values {
			vi, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return nil, err
			}
			args[i] = vi
		}
		return args, nil
	case ArgTypeKeyValueSlice:
		args := make(types.KeyValues[string, string], len(values))
		for i, kv := range values {
			k, v := str.ParseKeyValue(kv)
			if k == "" {
				continue
			}
			args[i] = types.KeyValue[string, string]{Key: k, Value: v}
		}
		return args, nil
	}
	return nil, fmt.Errorf("unknown type: %d", typ)
}

type simpleSpinner struct {
	sp   *yacspin.Spinner
	text string
}

func (s *simpleSpinner) Start() {
	// ignoring the error here
	_ = s.sp.Start()
}

func (s *simpleSpinner) Stop() {
	// ignoring the error here
	_ = s.sp.Stop()
}

func (s *simpleSpinner) SetText(text string) {
	s.text = text
	if text == "" {
		s.sp.Prefix("")
		return
	}
	s.sp.Prefix("      " + text + cancelMsg)
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
