package commands

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/clc/groups"
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
	cc.SetCommandGroup(groups.DDSID)
	cc.AddStringFlag(mapFlagName, "n", "", true, "IMap name")
	cc.AddBoolFlag(mapFlagShowType, "", false, false, "add the type names to the output")
	cc.SetTopLevel(true)
	cc.SetCommandUsage("map [command]")
	return nil
}

func (mc *MapCommand) Exec(ec plug.ExecContext) error {
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
		client, err := ec.Client(ctx)
		if err != nil {
			return nil, err
		}
		m, err := client.GetMap(ctx, mapName)
		if err != nil {
			return nil, err
		}
		mc.m = m
		return m, nil
	})
	return nil
}

func init() {
	cmd := &MapCommand{}
	Must(plug.Registry.RegisterCommand("map", cmd))
	plug.Registry.RegisterAugmentor("20-map", cmd)
}
