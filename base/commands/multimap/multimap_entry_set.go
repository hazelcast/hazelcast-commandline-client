//go:build std || multimap

package multimap

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
)

type MultiMapEntrySetCommand struct{}

func (mc *MultiMapEntrySetCommand) Init(cc plug.InitContext) error {
	help := "Get all entries of a MultiMap"
	cc.SetCommandHelp(help, help)
	cc.SetCommandUsage("entry-set")
	cc.SetPositionalArgCount(0, 0)
	return nil
}

func (mc *MultiMapEntrySetCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	mmName := ec.Props().GetString(multiMapFlagName)
	showType := ec.Props().GetBool(multiMapFlagShowType)
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	req := codec.EncodeMultiMapEntrySetRequest(mmName)
	rv, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Getting entries of multimap %s", mmName))
		return ci.InvokeOnRandomTarget(ctx, req, nil)
	})
	if err != nil {
		return err
	}
	stop()
	pairs := codec.DecodeMultiMapEntrySetResponse(rv.(*hazelcast.ClientMessage))
	rows := output.DecodePairs(ci, pairs, showType)
	if len(rows) > 0 {
		return ec.AddOutputRows(ctx, rows...)
	}
	ec.PrintlnUnnecessary("No entries found.")
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("multi-map:entry-set", &MultiMapEntrySetCommand{}))
}
