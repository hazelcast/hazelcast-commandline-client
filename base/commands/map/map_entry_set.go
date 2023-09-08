//go:build std || map

package _map

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/base"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"

	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
)

type MapEntrySetCommand struct{}

func (mc *MapEntrySetCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("entry-set")
	help := "Get all entries of a Map"
	cc.SetCommandHelp(help, help)
	return nil
}

func (mc *MapEntrySetCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	mapName := ec.Props().GetString(base.FlagName)
	showType := ec.Props().GetBool(base.FlagShowType)
	rowsV, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		ci, err := cmd.ClientInternal(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		req := codec.EncodeMapEntrySetRequest(mapName)
		sp.SetText(fmt.Sprintf("Getting entries of %s", mapName))
		resp, err := ci.InvokeOnRandomTarget(ctx, req, nil)
		if err != nil {
			return nil, err
		}
		pairs := codec.DecodeMapEntrySetResponse(resp)
		rows := output.DecodePairs(ci, pairs, showType)
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
	Must(plug.Registry.RegisterCommand("map:entry-set", &MapEntrySetCommand{}))
}
