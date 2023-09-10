//go:build std || set

package set

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/base"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type SetCommand struct{}

func (sc *SetCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("set")
	cc.AddCommandGroup(clc.GroupDDSID, clc.GroupDDSTitle)
	cc.SetCommandGroup(clc.GroupDDSID)
	cc.SetTopLevel(true)
	help := "Set operations"
	cc.SetCommandHelp(help, help)
	cc.AddStringFlag(base.FlagName, "n", base.DefaultName, false, "set name")
	cc.AddBoolFlag(base.FlagShowType, "", false, false, "add the type names to the output")
	return nil
}

func (sc *SetCommand) Exec(context.Context, plug.ExecContext) error {
	return nil
}

func init() {
	check.Must(plug.Registry.RegisterCommand("set", &SetCommand{}))
}
