//go:build base || map

package _map

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

type MapValuesCommand struct{}

func (mc *MapValuesCommand) Init(cc plug.InitContext) error {
	help := "Get all values of a Map"
	cc.SetCommandHelp(help, help)
	cc.SetCommandUsage("values [flags]")
	cc.SetPositionalArgCount(0, 0)
	return nil
}

func (mc *MapValuesCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	mapName := ec.Props().GetString(mapFlagName)
	showType := ec.Props().GetBool(mapFlagShowType)
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	req := codec.EncodeMapValuesRequest(mapName)
	rv, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Getting values of %s", mapName))
		return ci.InvokeOnRandomTarget(ctx, req, nil)
	})
	if err != nil {
		return err
	}
	stop()
	raw := codec.DecodeMapValuesResponse(rv.(*hazelcast.ClientMessage))
	var rows []output.Row
	for _, r := range raw {
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
	if len(rows) > 0 {
		return ec.AddOutputRows(ctx, rows...)
	}
	ec.PrintlnUnnecessary("No values found.")
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("map:values", &MapValuesCommand{}))
}
