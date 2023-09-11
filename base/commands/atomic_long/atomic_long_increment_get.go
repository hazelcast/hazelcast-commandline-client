//go:build std || atomiclong

package atomiclong

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type IncrementGetCommand struct{}

func (mc *IncrementGetCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("increment-get")
	help := "Increment the AtomicLong by the given value"
	cc.SetCommandHelp(help, help)
	cc.AddIntFlag(flagBy, "", 1, false, "value to increment by")
	return nil
}

func (mc *IncrementGetCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	return atomicLongChangeValue(ctx, ec, "Increment", func(i int64) int64 { return i })
}

func init() {
	check.Must(plug.Registry.RegisterCommand("atomic-long:increment-get", &IncrementGetCommand{}))
}
