package plug

import (
	"context"
	"io"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/internal/log"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
)

type InitContext interface {
	AddBoolFlag(long, short string, value bool, required bool, help string)
	AddCommandGroup(id, title string)
	AddIntFlag(long, short string, value int64, required bool, help string)
	AddStringConfig(name, value, flag string, help string)
	AddStringFlag(long, short, value string, required bool, help string)
	AddStringArg(key, title, help string)
	Hide()
	Interactive() bool
	SetCommandGroup(id string)
	SetCommandHelp(long, short string)
	SetCommandUsage(usage string)
	SetPositionalArgCount(min, max int)
	SetTopLevel(b bool)
}

type ExecContext interface {
	AddOutputRows(ctx context.Context, rows ...output.Row) error
	AddOutputStream(ctx context.Context, ch <-chan output.Row) error
	Args() []string
	ClientInternal(ctx context.Context) (*hazelcast.ClientInternal, error)
	CommandName() string
	Interactive() bool
	Logger() log.Logger
	Props() ReadOnlyProperties
	ShowHelpAndExit()
	Stderr() io.Writer
	Stdout() io.Writer
	Stdin() io.Reader
	ExecuteBlocking(ctx context.Context, f func(ctx context.Context, sp clc.Spinner) (any, error)) (value any, stop context.CancelFunc, err error)
	PrintlnUnnecessary(text string)
}

type ResultWrapper interface {
	WrapResult(f func() error) error
}
