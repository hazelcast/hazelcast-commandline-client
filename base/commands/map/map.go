package _map

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

const (
	mapFlagName     = "name"
	mapFlagShowType = "show-type"
	mapPropertyName = "map"
)

type MapCommand struct {
	m *hazelcast.Map
}

func (mc *MapCommand) Init(cc plug.InitContext) error {
	cc.SetCommandGroup(clc.GroupDDSID)
	cc.AddStringFlag(mapFlagName, "n", "", true, "map name")
	cc.AddBoolFlag(mapFlagShowType, "", false, false, "add the type names to the output")
	if !cc.Interactive() {
		cc.AddStringFlag(clc.PropertySchemaDir, "", paths.Schemas(), false, "set the schema directory")
	}
	cc.SetTopLevel(true)
	cc.SetCommandUsage("map COMMAND [flags]")
	help := "Map operations"
	cc.SetCommandHelp(help, help)
	return nil
}

func (mc *MapCommand) Exec(context.Context, plug.ExecContext) error {
	return nil
}

func (mc *MapCommand) Augment(ec plug.ExecContext, props *plug.Properties) error {
	ctx := context.TODO()
	props.SetBlocking(mapPropertyName, func() (any, error) {
		mapName := ec.Props().GetString(mapFlagName)
		if mapName == "" {
			return nil, fmt.Errorf("map name is required")
		}
		if mc.m != nil {
			return mc.m, nil
		}
		ci, err := ec.ClientInternal(ctx)
		if err != nil {
			return nil, err
		}
		hint := fmt.Sprintf("Getting map %s", mapName)
		mv, err := ec.ExecuteBlocking(ctx, hint, func(ctx context.Context) (any, error) {
			m, err := ci.Client().GetMap(ctx, mapName)
			if err != nil {
				return nil, err
			}
			return m, nil
		})
		mc.m = mv.(*hazelcast.Map)
		return mc.m, nil
	})
	return nil
}

func init() {
	cmd := &MapCommand{}
	Must(plug.Registry.RegisterCommand("map", cmd))
	plug.Registry.RegisterAugmentor("20-map", cmd)
}
