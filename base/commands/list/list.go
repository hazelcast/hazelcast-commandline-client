//go:build std || list

package list

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/base"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type ListCommand struct {
}

func (mc *ListCommand) Init(cc plug.InitContext) error {
	cc.AddCommandGroup(clc.GroupDDSID, clc.GroupDDSTitle)
	cc.SetCommandGroup(clc.GroupDDSID)
	cc.AddStringFlag(base.FlagName, "n", base.DefaultName, false, "list name")
	cc.AddBoolFlag(base.FlagShowType, "", false, false, "add the type names to the output")
	cc.SetTopLevel(true)
	cc.SetCommandUsage("list")
	help := "List operations"
	cc.SetCommandHelp(help, help)
	return nil
}

func (mc *ListCommand) Exec(context.Context, plug.ExecContext) error {
	return nil
}

func init() {
	check.Must(plug.Registry.RegisterCommand("list", &ListCommand{}))
}
