//go:build std || queue

package queue

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/base"
	"github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	metric "github.com/hazelcast/hazelcast-commandline-client/clc/metrics"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

type QueueOfferCommand struct{}

func (QueueOfferCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("offer")
	help := "Add values to the given Queue"
	cc.SetCommandHelp(help, help)
	commands.AddValueTypeFlag(cc)
	cc.AddStringSliceArg(base.ArgValue, base.ArgTitleValue, 1, clc.MaxArgs)
	return nil
}

func (QueueOfferCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	name := ec.Props().GetString(base.FlagName)
	rowsV, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		ci, err := cmd.ClientInternal(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		cid, vid := cmd.FindClusterIDs(ctx, ec)
		ec.Metrics().Increment(metric.NewKey(cid, vid), "total.queue."+cmd.RunningMode(ec))
		q, err := ci.Client().GetQueue(ctx, name)
		if err != nil {
			return nil, err
		}
		sp.SetText(fmt.Sprintf("Adding values into Queue '%s'", name))
		var rows []output.Row
		for _, arg := range ec.GetStringSliceArg(base.ArgValue) {
			vd, err := commands.MakeValueData(ec, ci, arg)
			if err != nil {
				return nil, err
			}
			v, err := q.Add(ctx, vd)
			if err != nil {
				return nil, err
			}
			rows = append(rows, output.Row{
				output.Column{
					Name:  "Value",
					Type:  serialization.TypeString,
					Value: arg,
				},
				output.Column{
					Name:  "Added",
					Type:  serialization.TypeBool,
					Value: v,
				},
			})
		}
		return rows, nil
	})
	if err != nil {
		return nil
	}
	stop()
	return ec.AddOutputRows(ctx, rowsV.([]output.Row)...)
}

func init() {
	check.Must(plug.Registry.RegisterCommand("queue:offer", &QueueOfferCommand{}))
}
