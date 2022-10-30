package _map

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"

	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
)

type MapEntrySetCommand struct{}

func (mc *MapEntrySetCommand) Init(cc plug.InitContext) error {
	help := "Get all entries of an IMap"
	cc.SetCommandHelp(help, help)
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
	hint := fmt.Sprintf("Getting entries of %s", mapName)
	rv, err := ec.ExecuteBlocking(ctx, hint, func(ctx context.Context) (any, error) {
		return ci.InvokeOnRandomTarget(ctx, req, nil)
	})
	if err != nil {
		return err
	}
	pairs := codec.DecodeMapEntrySetResponse(rv.(*hazelcast.ClientMessage))
	rows := output.DecodePairs(ci, pairs, showType)
	ec.AddOutputRows(rows...)
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("map:entry-set", &MapEntrySetCommand{}))
}
