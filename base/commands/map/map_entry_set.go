//go:build base || map

package _map

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"

	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
)

type MapEntrySetCommand struct{}

func (mc *MapEntrySetCommand) Init(cc plug.InitContext) error {
	help := "Get all entries of a Map"
	cc.SetCommandHelp(help, help)
	cc.SetCommandUsage("entry-set [-n map-name] [flags]")
	cc.SetPositionalArgCount(0, 0)
	return nil
}

func (mc *MapEntrySetCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	mapName := ec.Props().GetString(mapFlagName)
	showType := ec.Props().GetBool(mapFlagShowType)
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	req := codec.EncodeMapEntrySetRequest(mapName)
	rv, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Getting entries of %s", mapName))
		return ci.InvokeOnRandomTarget(ctx, req, nil)
	})
	if err != nil {
		return err
	}
	stop()
	pairs := codec.DecodeMapEntrySetResponse(rv.(*hazelcast.ClientMessage))
	rows := output.DecodePairs(ci, pairs, showType)
	if len(rows) > 0 {
		return ec.AddOutputRows(ctx, rows...)
	}
	if !ec.Props().GetBool(clc.PropertyQuiet) {
		I2(fmt.Fprintln(ec.Stdout(), "No entries found."))
	}
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("map:entry-set", &MapEntrySetCommand{}))
}
