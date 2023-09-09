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
	cc.SetCommandUsage("list")
	cc.AddCommandGroup(clc.GroupDDSID, clc.GroupDDSTitle)
	cc.SetCommandGroup(clc.GroupDDSID)
	cc.SetTopLevel(true)
	help := "List operations"
	cc.SetCommandHelp(help, help)
	cc.AddStringFlag(base.FlagName, "n", base.DefaultName, false, "list name")
	cc.AddBoolFlag(base.FlagShowType, "", false, false, "add the type names to the output")
	return nil
}

func (mc *ListCommand) Exec(context.Context, plug.ExecContext) error {
	return nil
}

func init() {
	check.Must(plug.Registry.RegisterCommand("list", &ListCommand{}))
}
