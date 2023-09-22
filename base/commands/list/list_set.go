//go:build std || list

package list

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/base"
	"github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/clc/metrics"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
)

type ListSetCommand struct{}

func (mc *ListSetCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("set")
	help := "Set a value at the given index in the list"
	cc.SetCommandHelp(help, help)
	commands.AddValueTypeFlag(cc)
	cc.AddInt64Arg(argIndex, argTitleIndex)
	cc.AddStringArg(base.ArgValue, base.ArgTitleValue)
	return nil
}

func (mc *ListSetCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	name := ec.Props().GetString(base.FlagName)
	index := ec.GetInt64Arg(argIndex)
	valueStr := ec.GetStringArg(base.ArgValue)
	rowV, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		ci, err := cmd.ClientInternal(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		cid, vid := cmd.FindClusterIDs(ctx, ec)
		ec.Metrics().Increment(metrics.NewKey(cid, vid), "total.list."+cmd.RunningModeString(ec))
		// get the list just to ensure the corresponding proxy is created
		_, err = getList(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		vd, err := commands.MakeValueData(ec, ci, valueStr)
		if err != nil {
			return nil, err
		}
		pid, err := internal.StringToPartitionID(ci, name)
		if err != nil {
			return nil, err
		}
		sp.SetText(fmt.Sprintf("Setting the value of the List '%s'", name))
		req := codec.EncodeListSetRequest(name, int32(index), vd)
		resp, err := ci.InvokeOnPartition(ctx, req, pid, nil)
		if err != nil {
			return nil, err
		}
		data := codec.DecodeListSetResponse(resp)
		return convertDataToRow(ci, "Last Value", data, ec.Props().GetBool(base.FlagShowType))
	})
	if err != nil {
		return err
	}
	stop()
	ec.PrintlnUnnecessary("OK Set the value in the List.\n")
	return ec.AddOutputRows(ctx, rowV.(output.Row))
}

func init() {
	check.Must(plug.Registry.RegisterCommand("list:set", &ListSetCommand{}))
}
