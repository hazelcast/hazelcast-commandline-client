//go:build std || atomiclong

package atomiclong

import (
	"context"

	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type AtomicLongDecrementGetCommand struct{}

func (mc *AtomicLongDecrementGetCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("decrement-get")
	help := "Decrement the AtomicLong by the given value"
	cc.SetCommandHelp(help, help)
	cc.AddIntFlag(atomicLongFlagBy, "", 1, false, "value to decrement by")
	return nil
}

func (mc *AtomicLongDecrementGetCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	return atomicLongChangeValue(ctx, ec, "Decrement", func(i int64) int64 { return -1 * i })
}

func init() {
	Must(plug.Registry.RegisterCommand("atomic-long:decrement-get", &AtomicLongDecrementGetCommand{}))
}
