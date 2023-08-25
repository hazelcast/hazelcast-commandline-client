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
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

type MultiMapKeySetCommand struct{}

func (mc *MultiMapKeySetCommand) Init(cc plug.InitContext) error {
	help := "Get all keys of a MultiMap"
	cc.SetCommandHelp(help, help)
	cc.SetCommandUsage("key-set")
	cc.SetPositionalArgCount(0, 0)
	return nil
}

func (mc *MultiMapKeySetCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	mmName := ec.Props().GetString(multiMapFlagName)
	showType := ec.Props().GetBool(multiMapFlagShowType)
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	req := codec.EncodeMultiMapKeySetRequest(mmName)
	rv, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Getting keys of %s", mmName))
		return ci.InvokeOnRandomTarget(ctx, req, nil)
	})
	if err != nil {
		return err
	}
	stop()
	raw := codec.DecodeMultiMapKeySetResponse(rv.(*hazelcast.ClientMessage))
	var rows []output.Row
	for _, r := range raw {
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
	if len(rows) > 0 {
		return ec.AddOutputRows(ctx, rows...)
	}
	ec.PrintlnUnnecessary("No entries found.")
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("multi-map:key-set", &MultiMapKeySetCommand{}))
}
