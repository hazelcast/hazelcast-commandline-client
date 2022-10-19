package plug

import (
	"context"
	"io"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/internal/log"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
)

type InitContext interface {
	AddStringFlag(long, short, value string, required bool, help string)
	AddBoolFlag(long, short string, value bool, required bool, help string)
	AddIntFlag(long, short string, value int64, required bool, help string)
	SetPositionalArgCount(min, max int)
	Interactive() bool
	SetCommandHelp(long, short string)
	SetCommandUsage(usage string)
	AddCommandGroup(id, title string)
	SetCommandGroup(id string)
	AddStringConfig(name, value, flag string, help string)
	SetTopLevel(b bool)
}

type ExecContext interface {
	Logger() log.Logger
	Stdout() io.Writer
	Stderr() io.Writer
	Args() []string
	Props() ReadOnlyProperties
	ClientInternal(ctx context.Context) (*hazelcast.ClientInternal, error)
	Interactive() bool
	AddOutputRows(row ...output.Row)
}

type ConfigContext interface {
	AddStringConfig(key, value, help string)
}
