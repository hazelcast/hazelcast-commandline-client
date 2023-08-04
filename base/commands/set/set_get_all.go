//go:build std || set

package set

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

type SetGetAllCommand struct{}

func (sc *SetGetAllCommand) Init(cc plug.InitContext) error {
	help := "Return the elements of the given Set"
	cc.SetCommandHelp(help, help)
	cc.SetCommandUsage("get-all")
	cc.SetPositionalArgCount(0, 0)
	return nil
}

func (sc *SetGetAllCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	name := ec.Props().GetString(setFlagName)
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	req := codec.EncodeSetGetAllRequest(name)
	pID, err := stringToPartitionID(ci, name)
	if err != nil {
		return err
	}
	sv, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Removing from set %s", name))
		return ci.InvokeOnPartition(ctx, req, pID, nil)
	})
	if err != nil {
		return err
	}
	stop()
	resp := codec.DecodeSetGetAllResponse(sv.(*hazelcast.ClientMessage))
	var rows []output.Row
	for _, r := range resp {
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
		if ec.Props().GetBool(setFlagShowType) {
			row = append(row, output.Column{
				Name:  output.NameValueType,
				Type:  serialization.TypeString,
				Value: serialization.TypeToLabel(r.Type()),
			})
		}
		rows = append(rows, row)
	}
	return ec.AddOutputRows(ctx, rows...)
}

func init() {
	Must(plug.Registry.RegisterCommand("set:get-all", &SetGetAllCommand{}))
}
