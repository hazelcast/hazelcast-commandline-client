//go:build std || atomiclong

package atomiclong

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

const (
	argValue      = "value"
	argTitleValue = "value"
)

type AtomicLongSetCommand struct{}

func (mc *AtomicLongSetCommand) Unwrappable() {}

func (mc *AtomicLongSetCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("set")
	help := "Set the value of the AtomicLong"
	cc.SetCommandHelp(help, help)
	cc.AddInt64Arg(argValue, argTitleValue)
	return nil
}

func (mc *AtomicLongSetCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	valueV, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		ali, err := getAtomicLong(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		sp.SetText(fmt.Sprintf("Setting value of AtomicLong %s", ali.Name()))
		value := ec.GetInt64Arg(argValue)
		err = ali.Set(ctx, value)
		if err != nil {
			return nil, err
		}
		msg := fmt.Sprintf("OK Set AtomicLong %s.\n", ali.Name())
		ec.PrintlnUnnecessary(msg)
		return value, nil
	})
	if err != nil {
		return err
	}
	stop()
	return ec.AddOutputRows(ctx, output.Row{
		output.Column{
			Name:  "Value",
			Type:  serialization.TypeInt64,
			Value: valueV,
		},
	})
}

func init() {
	Must(plug.Registry.RegisterCommand("atomic-long:set", &AtomicLongSetCommand{}))
}
