//go:build std || atomiclong

package atomiclong

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

const argValue = "value"

type AtomicLongSetCommand struct{}

func (mc *AtomicLongSetCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("set")
	help := "Set the value of the AtomicLong"
	cc.SetCommandHelp(help, help)
	cc.AddInt64Arg(argValue, "value")
	return nil
}

func (mc *AtomicLongSetCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	al, err := ec.Props().GetBlocking(atomicLongPropertyName)
	if err != nil {
		return err
	}
	value := ec.GetInt64Arg(argValue)
	ali := al.(*hazelcast.AtomicLong)
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Setting value of AtomicLong %s", ali.Name()))
		err := ali.Set(ctx, int64(value))
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return err
	}
	stop()
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("atomic-long:set", &AtomicLongSetCommand{}))
}
