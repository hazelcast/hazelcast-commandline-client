package commands

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-go-client"

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

func (mc *MapCommand) Init(cc plug.CommandContext) error {
	cc.AddStringFlag(mapFlagName, "n", "", true, "IMap name")
	cc.AddBoolFlag(mapFlagShowType, "", false, false, "add the type names to the output")
	return nil
}

func (mc *MapCommand) Exec(ctx plug.ExecContext) error {
	I2(fmt.Fprintln(ctx.Stdout(), "Map:", ctx.Props().GetString("name")))
	return nil
}

func (mc *MapCommand) Augment(ec plug.ExecContext, props *plug.Properties) error {
	ctx := context.TODO()
	mapName := ec.Props().GetString(mapFlagName)
	if mapName == "" {
		return fmt.Errorf("map name is required")
	}
	props.SetBlocking(mapPropertyName, func() (any, error) {
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
