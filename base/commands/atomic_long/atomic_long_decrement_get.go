//go:build base || atomiclong

package _atomiclong

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

type AtomicLongDecrementGetCommand struct{}

func (mc *AtomicLongDecrementGetCommand) Init(cc plug.InitContext) error {
	cc.SetPositionalArgCount(0, 0)
	help := "Decrement the atomic long by the given value"
	cc.AddIntFlag(atomicLongFlagBy, "", 1, false, "value to decrement by")
	cc.SetCommandHelp(help, help)
	cc.SetCommandUsage("decrement-get [flags]")
	return nil
}

func (mc *AtomicLongDecrementGetCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	al, err := ec.Props().GetBlocking(atomicLongPropertyName)
	if err != nil {
		return err
	}
	dec := ec.Props().GetInt(atomicLongFlagBy)
	ali := al.(*hazelcast.AtomicLong)
	vali, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Decrementing the atomic long %s", ali.Name()))
		val, err := ali.AddAndGet(ctx, -1*dec)
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
	Must(plug.Registry.RegisterCommand("atomic-long:decrement-get", &AtomicLongDecrementGetCommand{}))
}
