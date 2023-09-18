//go:build std || set

package set

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/base"
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

type GetAllCommand struct{}

func (GetAllCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("get-all")
	help := "Return the elements of the given Set"
	cc.SetCommandHelp(help, help)
	return nil
}

func (GetAllCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	name := ec.Props().GetString(base.FlagName)
	rowsV, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		ci, err := cmd.ClientInternal(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		cid, vid := cmd.FindClusterIDs(ctx, ec)
		ec.Metrics().Increment(metric.NewKey(cid, vid), "total.set."+cmd.RunningMode(ec))
		req := codec.EncodeSetGetAllRequest(name)
		pID, err := internal.StringToPartitionID(ci, name)
		if err != nil {
			return nil, err
		}
		sp.SetText(fmt.Sprintf("Removing from Set '%s'", name))
		resp, err := ci.InvokeOnPartition(ctx, req, pID, nil)
		if err != nil {
			return nil, err
		}
		data := codec.DecodeSetGetAllResponse(resp)
		showType := ec.Props().GetBool(base.FlagShowType)
		var rows []output.Row
		for _, r := range data {
			val, err := ci.DecodeData(*r)
			if err != nil {
				ec.Logger().Info("The value was not decoded, due to error: %s", err.Error())
				val = serialization.NondecodedType(serialization.TypeToLabel(r.Type()))
			}
			row := output.Row{
				{
					Name:  "Value",
					Type:  r.Type(),
					Value: val,
				},
			}
			if showType {
				row = append(row, output.Column{
					Name:  output.NameValueType,
					Type:  serialization.TypeString,
					Value: serialization.TypeToLabel(r.Type()),
				})
			}
			rows = append(rows, row)
		}
		return rows, nil
	})
	if err != nil {
		return err
	}
	stop()
	rows := rowsV.([]output.Row)
	if len(rows) == 0 {
		ec.PrintlnUnnecessary("OK No items in the set.")
		return nil
	}
	return ec.AddOutputRows(ctx, rowsV.([]output.Row)...)
}

func init() {
	check.Must(plug.Registry.RegisterCommand("set:get-all", &GetAllCommand{}))
}
