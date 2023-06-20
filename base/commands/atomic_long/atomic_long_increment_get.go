//go:build base || atomiclong

package atomiclong

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
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
	al, err := ec.Props().GetBlocking(atomicLongPropertyName)
	if err != nil {
		return err
	}
	inc := ec.Props().GetInt(atomicLongFlagBy)
	ali := al.(*hazelcast.AtomicLong)
	vali, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Incrementing the AtomicLong %s", ali.Name()))
		val, err := ali.AddAndGet(ctx, inc)
		if err != nil {
			return nil, err
		}
		return val, nil
	})
	if err != nil {
		return err
	}
	stop()
	val := vali.(int64)
	row := output.Row{
		output.Column{
			Name:  output.NameValue,
			Type:  serialization.TypeInt64,
			Value: val,
		},
	}
	return ec.AddOutputRows(ctx, row)
}

func init() {
	Must(plug.Registry.RegisterCommand("atomic-long:increment-get", &AtomicLongIncrementGetCommand{}))
}
