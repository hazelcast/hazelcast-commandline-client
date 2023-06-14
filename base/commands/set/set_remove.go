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
	"github.com/hazelcast/hazelcast-go-client"
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
	setName := ec.Props().GetString(setFlagName)
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	var rows []output.Row
	for _, arg := range ec.Args() {
		vd, err := makeValueData(ec, ci, arg)
		if err != nil {
			return err
		}
		req := codec.EncodeSetRemoveRequest(setName, vd)
		pID, err := stringToPartitionID(ci, setName)
		if err != nil {
			return err
		}
		sv, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
			sp.SetText(fmt.Sprintf("Removing from set %s", setName))
			return ci.InvokeOnPartition(ctx, req, pID, nil)
		})
		if err != nil {
			return err
		}
		stop()
		resp := codec.DecodeSetRemoveResponse(sv.(*hazelcast.ClientMessage))
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
	return ec.AddOutputRows(ctx, rows...)
}

func init() {
	Must(plug.Registry.RegisterCommand("set:remove", &SetRemoveCommand{}))
}
