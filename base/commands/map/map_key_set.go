//go:build std || map

package _map

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/base"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"

	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
)

type MapKeySetCommand struct{}

func (mc *MapKeySetCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("key-set")
	help := "Get all keys of a Map"
	cc.SetCommandHelp(help, help)
	return nil
}

func (mc *MapKeySetCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	mapName := ec.Props().GetString(base.FlagName)
	showType := ec.Props().GetBool(base.FlagShowType)
	rowsV, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		ci, err := cmd.ClientInternal(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		req := codec.EncodeMapKeySetRequest(mapName)
		sp.SetText(fmt.Sprintf("Getting keys of %s", mapName))
		resp, err := ci.InvokeOnRandomTarget(ctx, req, nil)
		if err != nil {
			return nil, err
		}
		data := codec.DecodeMapKeySetResponse(resp)
		var rows []output.Row
		for _, r := range data {
			var row output.Row
			t := r.Type()
			v, err := ci.DecodeData(*r)
			if err != nil {
				v = serialization.NondecodedType(serialization.TypeToLabel(t))
			}
			row = append(row, output.NewKeyColumn(t, v))
			if showType {
				row = append(row, output.NewKeyTypeColumn(t))
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
		ec.PrintlnUnnecessary("OK No entries found.")
		return nil

	}
	return ec.AddOutputRows(ctx, rows...)
}

func init() {
	Must(plug.Registry.RegisterCommand("map:key-set", &MapKeySetCommand{}))
}
