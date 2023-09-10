//go:build std || queue

package queue

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/base"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type QueueCommand struct{}

func (QueueCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("queue")
	cc.AddCommandGroup(clc.GroupDDSID, clc.GroupDDSTitle)
	cc.SetCommandGroup(clc.GroupDDSID)
	cc.SetTopLevel(true)
	help := "Queue operations"
	cc.SetCommandHelp(help, help)
	cc.AddStringFlag(base.FlagName, "n", base.DefaultName, false, "queue name")
	cc.AddBoolFlag(base.FlagShowType, "", false, false, "add the type names to the output")
	return nil
}

func (QueueCommand) Exec(context.Context, plug.ExecContext) error {
	return nil
}

func init() {
	cmd := &QueueCommand{}
	check.Must(plug.Registry.RegisterCommand("queue", cmd))
}
