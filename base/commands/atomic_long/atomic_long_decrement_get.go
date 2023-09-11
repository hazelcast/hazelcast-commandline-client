//go:build std || atomiclong

package atomiclong

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type DecrementGetCommand struct{}

func (DecrementGetCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("decrement-get")
	help := "Decrement the AtomicLong by the given value"
	cc.SetCommandHelp(help, help)
	cc.AddIntFlag(flagBy, "", 1, false, "value to decrement by")
	return nil
}

func (DecrementGetCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	return atomicLongChangeValue(ctx, ec, "Decrement", func(i int64) int64 { return -1 * i })
}

func init() {
	check.Must(plug.Registry.RegisterCommand("atomic-long:decrement-get", &DecrementGetCommand{}))
}
