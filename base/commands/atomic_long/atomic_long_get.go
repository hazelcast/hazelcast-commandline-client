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

type AtomicLongGetCommand struct{}

func (mc *AtomicLongGetCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("get")
	help := "Get the value of the AtomicLong"
	cc.SetCommandHelp(help, help)
	return nil
}

func (mc *AtomicLongGetCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	val, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		ali, err := getAtomicLong(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		sp.SetText(fmt.Sprintf("Getting value of AtomicLong %s", ali.Name()))
		val, err := ali.Get(ctx)
		if err != nil {
			return nil, err
		}
		return val, nil
	})
	if err != nil {
		return err
	}
	stop()
	row := output.Row{
		output.Column{
			Name:  "Value",
			Type:  serialization.TypeInt64,
			Value: val,
		},
	}
	return ec.AddOutputRows(ctx, row)
}

func init() {
	Must(plug.Registry.RegisterCommand("atomic-long:get", &AtomicLongGetCommand{}))
}
