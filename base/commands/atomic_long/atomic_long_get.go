//go:build std || atomiclong

package atomiclong

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/clc/metrics"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

type GetCommand struct{}

func (GetCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("get")
	help := "Get the value of the AtomicLong"
	cc.SetCommandHelp(help, help)
	return nil
}

func (GetCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	row, stop, err := cmd.ExecuteBlocking(ctx, ec, func(ctx context.Context, sp clc.Spinner) (output.Row, error) {
		ali, err := getAtomicLong(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		cid, vid := cmd.FindClusterIDs(ctx, ec)
		ec.Metrics().Increment(metrics.NewKey(cid, vid), "total.atomiclong."+cmd.RunningMode(ec))
		sp.SetText(fmt.Sprintf("Getting value of AtomicLong %s", ali.Name()))
		val, err := ali.Get(ctx)
		if err != nil {
			return nil, err
		}
		row := output.Row{
			output.Column{
				Name:  "Value",
				Type:  serialization.TypeInt64,
				Value: val,
			},
		}
		return row, nil
	})
	if err != nil {
		return err
	}
	stop()
	return ec.AddOutputRows(ctx, row)
}

func init() {
	check.Must(plug.Registry.RegisterCommand("atomic-long:get", &GetCommand{}))
}
