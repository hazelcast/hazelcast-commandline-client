//go:build std || map

package _map

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/base"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

type MapValuesCommand struct{}

func (mc *MapValuesCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("values")
	help := "Get all values of a Map"
	cc.SetCommandHelp(help, help)
	return nil
}

func (mc *MapValuesCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	mapName := ec.Props().GetString(base.FlagName)
	showType := ec.Props().GetBool(base.FlagShowType)
	rowsV, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		ci, err := cmd.ClientInternal(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		sp.SetText(fmt.Sprintf("Getting values of %s", mapName))
		req := codec.EncodeMapValuesRequest(mapName)
		resp, err := ci.InvokeOnRandomTarget(ctx, req, nil)
		if err != nil {
			return nil, err
		}
		data := codec.DecodeMapValuesResponse(resp)
		var rows []output.Row
		for _, r := range data {
			var row output.Row
			t := r.Type()
			v, err := ci.DecodeData(*r)
			if err != nil {
				v = serialization.NondecodedType(serialization.TypeToLabel(t))
			}
			row = append(row, output.NewValueColumn(t, v))
			if showType {
				row = append(row, output.NewValueTypeColumn(t))
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
		ec.PrintlnUnnecessary("OK the map has no values.")
		return nil
	}
	return ec.AddOutputRows(ctx, rows...)
}

func init() {
	Must(plug.Registry.RegisterCommand("map:values", &MapValuesCommand{}))
}
