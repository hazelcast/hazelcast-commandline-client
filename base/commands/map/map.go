//go:build std || map

package _map

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/base"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type MapCommand struct{}

func (mc *MapCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("map")
	cc.SetTopLevel(true)
	cc.AddCommandGroup(clc.GroupDDSID, clc.GroupDDSTitle)
	cc.SetCommandGroup(clc.GroupDDSID)
	help := "Map operations"
	cc.SetCommandHelp(help, help)
	cc.AddStringFlag(base.FlagName, "n", defaultMapName, false, "map name")
	cc.AddBoolFlag(base.FlagShowType, "", false, false, "add the type names to the output")
	return nil
}

func (mc *MapCommand) Exec(context.Context, plug.ExecContext) error {
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("map", &MapCommand{}))
}
