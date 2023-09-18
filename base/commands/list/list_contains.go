//go:build std || list

package list

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/base"
	"github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	metric "github.com/hazelcast/hazelcast-commandline-client/clc/metrics"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

type ListContainsCommand struct{}

func (mc *ListContainsCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("contains")
	help := "Check if the value is present in the list"
	cc.SetCommandHelp(help, help)
	commands.AddValueTypeFlag(cc)
	cc.AddStringArg(base.ArgValue, base.ArgTitleValue)
	return nil
}

func (mc *ListContainsCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	name := ec.Props().GetString(base.FlagName)
	ok, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		ci, err := cmd.ClientInternal(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		cid, vid := cmd.FindClusterIDs(ctx, ec)
		ec.Metrics().Increment(metric.NewKey(cid, vid), "total.list."+cmd.RunningMode(ec))
		// get the list just to ensure the corresponding proxy is created
		_, err = getList(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		valueStr := ec.GetStringArg(base.ArgValue)
		vd, err := commands.MakeValueData(ec, ci, valueStr)
		if err != nil {
			return nil, err
		}
		pid, err := internal.StringToPartitionID(ci, name)
		if err != nil {
			return nil, err
		}
		sp.SetText(fmt.Sprintf("Checking if value exists in the List '%s'", name))
		req := codec.EncodeListContainsRequest(name, vd)
		resp, err := ci.InvokeOnPartition(ctx, req, pid, nil)
		if err != nil {
			return nil, err
		}
		contains := codec.DecodeListContainsResponse(resp)
		return contains, nil
	})
	if err != nil {
		return err
	}
	stop()
	return ec.AddOutputRows(ctx, output.Row{
		{
			Name:  "Contains",
			Type:  serialization.TypeBool,
			Value: ok,
		},
	})
}

func init() {
	check.Must(plug.Registry.RegisterCommand("list:contains", &ListContainsCommand{}))
}
