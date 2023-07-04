package set

import (
	"context"
	"fmt"
	"math"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

type SetRemoveCommand struct{}

func (sc *SetRemoveCommand) Init(cc plug.InitContext) error {
	addValueTypeFlag(cc)
	cc.SetPositionalArgCount(1, math.MaxInt)
	help := "Remove values from the given Set"
	cc.SetCommandHelp(help, help)
	cc.SetCommandUsage("remove [values] [flags]")
	return nil
}

func (sc *SetRemoveCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	name := ec.Props().GetString(setFlagName)
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}

	rows, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Removing from set %s", name))
		var rows []output.Row
		for _, arg := range ec.Args() {
			vd, err := makeValueData(ec, ci, arg)
			if err != nil {
				return nil, err
			}
			req := codec.EncodeSetRemoveRequest(name, vd)
			pID, err := stringToPartitionID(ci, name)
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
			if ec.Props().GetBool(setFlagShowType) {
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
	Must(plug.Registry.RegisterCommand("set:remove", &SetRemoveCommand{}))
}
