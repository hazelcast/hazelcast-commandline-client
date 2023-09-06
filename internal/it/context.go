package it

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/log"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/terminal"
	"github.com/hazelcast/hazelcast-commandline-client/internal/types"
)

type CommandContext struct {
	LongHelp      string
	ShortHelp     string
	Use           string
	IsInteractive bool
}

func (c CommandContext) AddKeyValueSliceArg(key, title string, min, max int) {
	panic("implement me")
}

func (c CommandContext) AddStringArg(key, title string) {
	panic("implement me")
}

func (c CommandContext) AddStringSliceArg(key, title string, min, max int) {
	panic("implement me")
}

func (c CommandContext) AddInt64Arg(key, title string) {
	panic("implement me")
}

func (c CommandContext) AddStringFlag(long, short, value string, required bool, help string) {
	panic("implement me")
}

func (c CommandContext) AddBoolFlag(long, short string, value bool, required bool, help string) {
	panic("implement me")
}

func (c CommandContext) AddIntFlag(long, short string, value int64, required bool, help string) {
	panic("implement me")
}

func (c CommandContext) SetPositionalArgCount(min, max int) {
	panic("implement me")
}

func (c CommandContext) Hide() {
	panic("implement me")
}

func (c CommandContext) Interactive() bool {
	return c.IsInteractive
}

func (c *CommandContext) SetCommandHelp(long, short string) {
	c.LongHelp = long
	c.ShortHelp = short
}

func (c *CommandContext) SetCommandUsage(usage string) {
	c.Use = usage
}

func (c CommandContext) AddCommandGroup(id, title string) {
	panic("implement me")
}

func (c CommandContext) SetCommandGroup(id string) {
	panic("implement me")
}

func (c CommandContext) AddStringConfig(name, value, flag string, help string) {
	panic("implement me")
}

func (c CommandContext) SetTopLevel(b bool) {
	panic("implement me")
}

type ExecContext struct {
	lg      *Logger
	stdout  *bytes.Buffer
	stderr  *bytes.Buffer
	stdin   *bytes.Buffer
	args    []string
	props   *plug.Properties
	Rows    []output.Row
	Spinner *Spinner
}

func NewExecuteContext(args []string) *ExecContext {
	return &ExecContext{
		lg:      NewLogger(),
		stdout:  &bytes.Buffer{},
		stderr:  &bytes.Buffer{},
		stdin:   &bytes.Buffer{},
		args:    args,
		props:   plug.NewProperties(),
		Spinner: NewSpinner(),
	}
}

func (ec *ExecContext) GetKeyValuesArg(key string) types.KeyValues[string, string] {
	panic("implement me")
}

func (ec *ExecContext) ExecuteBlocking(ctx context.Context, f func(context.Context, clc.Spinner) (any, error)) (any, context.CancelFunc, error) {
	v, err := f(ctx, ec.Spinner)
	stop := func() {}
	return v, stop, err
}

func (ec *ExecContext) Props() plug.ReadOnlyProperties {
	return ec.props
}

func (ec *ExecContext) ClientInternal(ctx context.Context) (*hazelcast.ClientInternal, error) {
	panic("implement me")
}

func (ec *ExecContext) Interactive() bool {
	return false
}

func (ec *ExecContext) AddOutputStream(ctx context.Context, ch <-chan output.Row) error {
	panic("implement me")
}

func (ec *ExecContext) AddOutputRows(ctx context.Context, rows ...output.Row) error {
	ec.Rows = append(ec.Rows, rows...)
	return nil
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

func (ec *ExecContext) ShowHelpAndExit() {
	panic("implement me")
}

func (ec *ExecContext) CommandName() string {
	panic("implement me")
}

func (ec *ExecContext) StdoutText() string {
	b, err := io.ReadAll(ec.stdout)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func (ec *ExecContext) StderrText() string {
	b, err := io.ReadAll(ec.stderr)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func (ec *ExecContext) LoggerText() string {
	return ec.lg.String()
}

func (ec *ExecContext) Set(name string, value any) {
	ec.props.Set(name, value)
}
func (ec *ExecContext) Get(name string) (any, bool) {
	return ec.props.Get(name)
}

func (ec *ExecContext) PrintlnUnnecessary(text string) {
	quiet := ec.Props().GetBool(clc.PropertyQuiet) || terminal.IsPipe(ec.Stdout())
	if !quiet {
		check.I2(fmt.Fprintln(ec.Stdout(), text))
	}
}

func (ec *ExecContext) GetStringArg(key string) string {
	panic("implement me")
}

func (ec *ExecContext) GetStringSliceArg(key string) []string {
	panic("implement me")
}

func (ec *ExecContext) GetInt64Arg(key string) int64 {
	panic("implement me")
}

func (ec *ExecContext) WrapResult(f func() error) error {
	return f()
}

type Spinner struct {
	Texts      []string
	Progresses []float32
}

func NewSpinner() *Spinner {
	return &Spinner{}
}

func (s *Spinner) Reset() {
	s.Texts = nil
	s.Progresses = nil
}

func (s *Spinner) SetText(text string) {
	s.Texts = append(s.Texts, text)
}

func (s *Spinner) SetProgress(progress float32) {
	s.Progresses = append(s.Progresses, progress)
}
