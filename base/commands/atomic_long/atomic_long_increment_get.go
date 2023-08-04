//go:build std || atomiclong

package atomiclong

import (
	"context"

	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type AtomicLongIncrementGetCommand struct{}

func (mc *AtomicLongIncrementGetCommand) Init(cc plug.InitContext) error {
	cc.SetPositionalArgCount(0, 0)
	help := "Increment the atomic long by the given value"
	cc.AddIntFlag(atomicLongFlagBy, "", 1, false, "value to increment by")
	cc.SetCommandHelp(help, help)
	cc.SetCommandUsage("increment-get [flags]")
	return nil
}

func (mc *AtomicLongIncrementGetCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	return atomicLongChangeValue(ctx, ec, "Increment", func(i int64) int64 { return i })
}

func init() {
	Must(plug.Registry.RegisterCommand("atomic-long:increment-get", &AtomicLongIncrementGetCommand{}))
}
