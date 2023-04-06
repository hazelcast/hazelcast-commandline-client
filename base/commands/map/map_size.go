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
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

type MapSizeCommand struct{}

func (mc *MapSizeCommand) Init(cc plug.InitContext) error {
	help := "Return the size of the given Map"
	cc.SetCommandHelp(help, help)
	cc.SetCommandUsage("size [-n map-name]")
	cc.SetPositionalArgCount(0, 0)
	return nil
}

func (mc *MapSizeCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	mapName := ec.Props().GetString(mapFlagName)
	mv, err := ec.Props().GetBlocking(mapPropertyName)
	if err != nil {
		return err
	}
	m := mv.(*hazelcast.Map)
	sv, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Getting the size of the map %s", mapName))
		return m.Size(ctx)
	})
	if err != nil {
		return err
	}
	stop()
	return ec.AddOutputRows(ctx, output.Row{
		{
			Name:  "Size",
			Type:  serialization.TypeInt32,
			Value: int32(sv.(int)),
		},
	})
}

func init() {
	Must(plug.Registry.RegisterCommand("map:size", &MapSizeCommand{}))
}
