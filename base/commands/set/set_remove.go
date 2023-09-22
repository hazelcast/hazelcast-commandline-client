//go:build std || set

package set

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/base"
	"github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

type RemoveCommand struct{}

func (RemoveCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("remove")
	help := "Remove values from the given Set"
	cc.SetCommandHelp(help, help)
	commands.AddValueTypeFlag(cc)
	cc.AddStringSliceArg(base.ArgValue, base.ArgTitleValue, 1, clc.MaxArgs)
	return nil
}

func (sc *RemoveCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	name := ec.Props().GetString(base.FlagName)
	rows, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		ci, err := cmd.ClientInternal(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		cmd.IncrementClusterMetric(ctx, ec, "total.set")
		sp.SetText(fmt.Sprintf("Removing from Set '%s'", name))
		showType := ec.Props().GetBool(base.FlagShowType)
		var rows []output.Row
		for _, arg := range ec.GetStringSliceArg(base.ArgValue) {
			vd, err := commands.MakeValueData(ec, ci, arg)
			if err != nil {
				return nil, err
			}
			req := codec.EncodeSetRemoveRequest(name, vd)
			pID, err := internal.StringToPartitionID(ci, name)
			if err != nil {
				return nil, err
			}
			sv, err := ci.InvokeOnPartition(ctx, req, pID, nil)
			if err != nil {
				return nil, err
			}
			resp := codec.DecodeSetRemoveResponse(sv)
			row := output.Row{
				output.Column{
					Name:  output.NameValue,
					Type:  serialization.TypeBool,
					Value: resp,
				},
			}
			if showType {
				row = append(row, output.Column{
					Name:  output.NameValueType,
					Type:  serialization.TypeString,
					Value: serialization.TypeToLabel(serialization.TypeBool),
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
	return ec.AddOutputRows(ctx, rows.([]output.Row)...)
}

func init() {
	check.Must(plug.Registry.RegisterCommand("set:remove", &RemoveCommand{}))
}
