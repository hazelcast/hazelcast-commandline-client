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

type AtomicLongGetCommand struct{}

func (mc *AtomicLongGetCommand) Init(cc plug.InitContext) error {
	help := "Get the value of the atomic long"
	cc.SetCommandHelp(help, help)
	cc.SetCommandUsage("get [flags]")
	cc.SetPositionalArgCount(0, 0)
	return nil
}

func (mc *AtomicLongGetCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	al, err := ec.Props().GetBlocking(atomicLongPropertyName)
	if err != nil {
		return err
	}
	ali := al.(*hazelcast.AtomicLong)
	vali, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Setting value into atomicLong %s", ali.Name()))
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
	Must(plug.Registry.RegisterCommand("atomic-long:get", &AtomicLongGetCommand{}))
}
