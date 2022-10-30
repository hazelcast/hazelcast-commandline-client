package plug

import (
	"context"
	"io"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/internal/log"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
)

type InitContext interface {
	AddBoolFlag(long, short string, value bool, required bool, help string)
	AddCommandGroup(id, title string)
	AddIntFlag(long, short string, value int64, required bool, help string)
	AddStringConfig(name, value, flag string, help string)
	AddStringFlag(long, short, value string, required bool, help string)
	Interactive() bool
	SetCommandGroup(id string)
	SetCommandHelp(long, short string)
	SetCommandUsage(usage string)
	SetPositionalArgCount(min, max int)
	SetTopLevel(b bool)
}

type ExecContext interface {
	AddOutputRows(row ...output.Row)
	Args() []string
	ClientInternal(ctx context.Context) (*hazelcast.ClientInternal, error)
	CommandName() string
	Interactive() bool
	Logger() log.Logger
	Props() ReadOnlyProperties
	ShowHelpAndExit()
	Stderr() io.Writer
	Stdout() io.Writer
	FlushOutput() error
	ExecuteBlocking(ctx context.Context, hint string, f func(context.Context) (any, error)) (any, error)
}
